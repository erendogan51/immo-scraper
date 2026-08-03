package config

type Config struct {
	DB     DBConfig     `env:", prefix=POSTGRES_, required"`
	Bucket Bucket       `env:", prefix=BUCKET_, required"`
	Logger LoggerConfig `env:", prefix=LOG_, required"`

	ListenPort  int `env:"LISTEN_PORT, required"`
	MetricsPort int `env:"METRICS_PORT, required"`
}

type DBConfig struct {
	Username string `env:"USERNAME, required"`
	Password string `env:"PASSWORD, required"`
	Host     string `env:"HOST, required"`
	Port     int    `env:"PORT"`
	Database string `env:"DB, required"`
	SslMode  string `env:"SSL_MODE, required"`
}

type Bucket struct {
	Region   string `env:"REGION, required"`
	Endpoint string `env:"ENDPOINT, required"`
	Name     string `env:"NAME, required"`
	Key      string `env:"KEY, required"`
	Secret   string `env:"SECRET, required"`
}

type LogStyle string

const (
	LogStyleJson LogStyle = "json"
	LogStyleText LogStyle = "text"
)

type LoggerConfig struct {
	LogLevel string   `env:"LEVEL"`
	LogStyle LogStyle `env:"STYLE"`
}

type TLSConfig struct {
	CertPath string `env:"CERT_PATH"`
	KeyPath  string `env:"KEY_PATH"`
}
