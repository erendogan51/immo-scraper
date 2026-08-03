package sql

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"sync"

	"github.com/erendogan51/immo-scrapper/pkg/config"
	"github.com/erendogan51/immo-scrapper/pkg/db/sql/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var migrations embed.FS

var (
	dbCfg    config.DBConfig
	connPool *pgxpool.Pool
	dbInit   sync.Once
)

func GetDBPool(ctx context.Context, cfg ...config.DBConfig) (*pgxpool.Pool, error) {
	var initErr error
	dbInit.Do(func() {
		if len(cfg) != 1 {
			initErr = fmt.Errorf("no db config provided")
			return
		}

		dbCfg = cfg[0]

		pool, err := InitDb(ctx, dbCfg)
		if err != nil {
			initErr = err
			return
		}

		connPool = pool
	})
	if initErr != nil {
		return nil, initErr
	}

	return connPool, nil
}

func GetDBQueries(ctx context.Context) (*db.Queries, error) {
	pool, err := GetDBPool(ctx)
	if err != nil {
		return nil, err
	}

	return db.New(pool), nil
}

func CreateConnectionString(dbConfig config.DBConfig) string {
	host := dbConfig.Host
	if dbConfig.Port != 0 {
		host = fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port)
	}

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", dbConfig.Username, dbConfig.Password, host, dbConfig.Database, dbConfig.SslMode)
}

func InitDb(ctx context.Context, dbConfig config.DBConfig) (*pgxpool.Pool, error) {
	err := MigrateDBUp(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("could not migrate DB: %w", err)
	}

	connectionURL := CreateConnectionString(dbConfig)
	pool, err := ConnectPg(ctx, connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB, %w", err)
	}

	log.Info().Msg("initialized database")

	return pool, nil
}

func CheckPg(ctx context.Context, pool *pgxpool.Pool) error {
	err := pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("could not ping DB: %w", err)
	}
	return nil
}

func ConnectPg(ctx context.Context, connectionURL string) (*pgxpool.Pool, error) {
	pool, err := CreateConnPool(ctx, connectionURL)
	if err != nil {
		return nil, err
	}

	err = CheckPg(ctx, pool)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Connection success to database")

	return pool, nil
}

func CreateConnPool(ctx context.Context, connectionURL string) (*pgxpool.Pool, error) {
	pgCfg, err := pgxpool.ParseConfig(connectionURL)
	if err != nil {
		return nil, fmt.Errorf("invalid postgres config: %w", err)
	}
	pgCfg.AfterConnect = func(_ context.Context, c *pgx.Conn) error {
		pgxdecimal.Register(c.TypeMap())
		return nil
	}

	_connPool, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("could not connect to DB: %w", err)
	}

	err = _connPool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not ping DB: %w", err)
	}

	log.Info().Msg("successfully created connection pool")

	connPool = _connPool

	return _connPool, nil
}

func MigrateDBUp(_ context.Context, dbConfig config.DBConfig) error {
	log.Info().Msg("migrating DB")
	dbConn, err := sql.Open("pgx", CreateConnectionString(dbConfig))
	if err != nil {
		return fmt.Errorf("could not connect to DB: %w", err)
	}

	defer func(dbConn *sql.DB) {
		err := dbConn.Close()
		if err != nil {
			log.Err(err).Msg("failed to close DB conn after migration")
		}
	}(dbConn)

	dbDriver, err := pgxmigrate.WithInstance(dbConn, &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("could not get migration driver: %w", err)
	}
	defer func(dbDriver database.Driver) {
		err := dbDriver.Close()
		if err != nil {
			log.Err(err).Msg("failed to close DB driver after migration")
		}
	}(dbDriver)

	sourceDriver, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("could not get migration source: %w", err)
	}

	migration, err := migrate.NewWithInstance("embedded migrations", sourceDriver, dbConfig.Database, dbDriver)
	if err != nil {
		return fmt.Errorf("could not get migration: %w", err)
	}

	log.Info().Msg("running migrations")
	err = migration.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("no new migrations")
			return nil
		}

		return err
	}

	log.Info().Msg("completed DB migrations")
	return nil
}

func ResetDb(ctx context.Context, dbConfig config.DBConfig) error {
	log.Info().Msg("resetting DB")

	err := MigrateDBDown(dbConfig)
	if err != nil {
		return err
	}

	err = MigrateDBUp(ctx, dbConfig)
	if err != nil {
		return err
	}

	log.Info().Msg("resetting connection pool")
	connPool.Reset()
	log.Info().Msg("done resetting connection pool")
	log.Info().Msg("finished resetting DB")

	return nil
}

func MigrateDBDown(dbConfig config.DBConfig) error {
	log.Info().Msg("starting DB down migration")

	dbConn, err := sql.Open("pgx", CreateConnectionString(dbConfig))
	if err != nil {
		return fmt.Errorf("could not migrate DB down: %w", err)
	}
	defer func(dbConn *sql.DB) {
		err := dbConn.Close()
		if err != nil {
			log.Err(err).Msgf("failed to close db conn")
		}
	}(dbConn)

	pgxMigrateDriver, err := pgxmigrate.WithInstance(dbConn, &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("could not create pgxmigrate driver: %w", err)
	}
	defer func(pgxMigrateDriver database.Driver) {
		err := pgxMigrateDriver.Close()
		if err != nil {
			log.Err(err).Msgf("failed to close pgx migration driver")
		}
	}(pgxMigrateDriver)

	migrationsSource, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("could not load migrations files: %w", err)
	}

	migration, err := migrate.NewWithInstance("embedded migrations", migrationsSource, dbConfig.Database, pgxMigrateDriver)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %w", err)
	}

	err = migration.Down()
	if err != nil {
		return fmt.Errorf("could not migrate down: %w", err)
	}

	log.Info().Msg("completed down migration")

	return nil
}
