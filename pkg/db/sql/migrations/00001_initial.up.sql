CREATE TABLE IF NOT EXISTS immo_listings
(
    ad_id            VARCHAR(64)                            NOT NULL PRIMARY KEY,
    description      text                                   NOT NULL,

    published_at     timestamptz,
    price            decimal,
    is_private       boolean                                not null,

    vertical_id      INTEGER                                NOT NULL,
    ad_type_id       INTEGER                                NOT NULL,
    product_id       INTEGER                                NOT NULL,
    advert_status_id VARCHAR(64)                            NOT NULL,
    advert_status    VARCHAR(256)                           NOT NULL,
    heading          TEXT,

    property_type    VARCHAR(256),
    rooms_layout     VARCHAR(64),
    number_of_rooms  NUMERIC,
    floor            INTEGER,

    address          VARCHAR(512),
    location         VARCHAR(512),
    district         VARCHAR(256),
    state            VARCHAR(256),
    country          VARCHAR(256),
    postcode         VARCHAR(32),
    latitude         DOUBLE PRECISION,
    longitude        DOUBLE PRECISION,

    living_area_sqm  INTEGER,
    usable_area_sqm  INTEGER,
    estate_size_sqm  INTEGER,

    price_display    VARCHAR(64),
    price_per_sqm    NUMERIC,
    rent_per_month   NUMERIC,

    org_id           VARCHAR(64),
    org_name         VARCHAR(256),

    main_image       TEXT,
    seo_url          TEXT                                   NOT NULL,

    -- Populated only once a listing's detail page (not just its search
    -- result) has been scraped.
    ad_uuid          UUID,
    start_date       TIMESTAMPTZ,
    end_date         TIMESTAMPTZ,
    first_published_at TIMESTAMPTZ,
    created_at       TIMESTAMPTZ,
    changed_at       TIMESTAMPTZ,
    municipality     VARCHAR(256),
    canonical_url    TEXT,

    -- Full willhaben attribute bag ({"attribute": [{"name", "values"}, ...]}
    -- as sent by the API), kept alongside the columns above so nothing is
    -- lost even for attributes that don't have a dedicated column.
    attributes       JSONB                    DEFAULT '{}'  NOT NULL,

    -- Listing detail page's image gallery ([{"id", "description",
    -- "mainImageUrl", "thumbnailImageUrl", "referenceImageUrl"}, ...]),
    -- empty until the detail page has been scraped.
    images           JSONB                    DEFAULT '[]'  NOT NULL,

    -- SEO path of the search this row was last seen through, e.g.
    -- "immobilien/mietwohnungen/wien".
    search_path      VARCHAR(512),

    first_seen_at    TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    last_seen_at     TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

CREATE INDEX IF NOT EXISTS immo_listings_district_idx ON immo_listings (district);
CREATE INDEX IF NOT EXISTS immo_listings_price_idx ON immo_listings (price);
CREATE INDEX IF NOT EXISTS immo_listings_last_seen_at_idx ON immo_listings (last_seen_at);
CREATE INDEX IF NOT EXISTS immo_listings_search_path_idx ON immo_listings (search_path);
