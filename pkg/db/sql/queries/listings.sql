-- name: UpsertListing :exec
INSERT INTO immo_listings (ad_id,
                           vertical_id,
                           ad_type_id,
                           product_id,
                           advert_status_id,
                           advert_status,
                           description,
                           heading,
                           property_type,
                           rooms_layout,
                           number_of_rooms,
                           floor,
                           address,
                           location,
                           district,
                           state,
                           country,
                           postcode,
                           latitude,
                           longitude,
                           living_area_sqm,
                           usable_area_sqm,
                           estate_size_sqm,
                           price,
                           price_display,
                           price_per_sqm,
                           rent_per_month,
                           org_id,
                           org_name,
                           is_private,
                           main_image,
                           seo_url,
                           published_at,
                           attributes,
                           search_path,
                           ad_uuid,
                           start_date,
                           end_date,
                           first_published_at,
                           created_at,
                           changed_at,
                           municipality,
                           canonical_url,
                           images)
VALUES (sqlc.arg(ad_id),
        sqlc.arg(vertical_id),
        sqlc.arg(ad_type_id),
        sqlc.arg(product_id),
        sqlc.arg(advert_status_id),
        sqlc.arg(advert_status),
        sqlc.arg(description),
        sqlc.narg(heading),
        sqlc.narg(property_type),
        sqlc.narg(rooms_layout),
        sqlc.narg(number_of_rooms),
        sqlc.narg(floor),
        sqlc.narg(address),
        sqlc.narg(location),
        sqlc.narg(district),
        sqlc.narg(state),
        sqlc.narg(country),
        sqlc.narg(postcode),
        sqlc.narg(latitude),
        sqlc.narg(longitude),
        sqlc.narg(living_area_sqm),
        sqlc.narg(usable_area_sqm),
        sqlc.narg(estate_size_sqm),
        sqlc.narg(price),
        sqlc.narg(price_display),
        sqlc.narg(price_per_sqm),
        sqlc.narg(rent_per_month),
        sqlc.narg(org_id),
        sqlc.narg(org_name),
        sqlc.arg(is_private),
        sqlc.narg(main_image),
        sqlc.arg(seo_url),
        sqlc.narg(published_at),
        sqlc.arg(attributes),
        sqlc.narg(search_path),
        sqlc.narg(ad_uuid),
        sqlc.narg(start_date),
        sqlc.narg(end_date),
        sqlc.narg(first_published_at),
        sqlc.narg(created_at),
        sqlc.narg(changed_at),
        sqlc.narg(municipality),
        sqlc.narg(canonical_url),
        sqlc.arg(images))
ON CONFLICT (ad_id) DO UPDATE SET vertical_id        = excluded.vertical_id,
                                  ad_type_id         = excluded.ad_type_id,
                                  product_id         = excluded.product_id,
                                  advert_status_id   = excluded.advert_status_id,
                                  advert_status      = excluded.advert_status,
                                  description        = excluded.description,
                                  heading            = excluded.heading,
                                  property_type      = excluded.property_type,
                                  rooms_layout       = excluded.rooms_layout,
                                  number_of_rooms    = excluded.number_of_rooms,
                                  floor              = excluded.floor,
                                  address            = excluded.address,
                                  location           = excluded.location,
                                  district           = excluded.district,
                                  state              = excluded.state,
                                  country            = excluded.country,
                                  postcode           = excluded.postcode,
                                  latitude           = excluded.latitude,
                                  longitude          = excluded.longitude,
                                  living_area_sqm    = excluded.living_area_sqm,
                                  usable_area_sqm    = excluded.usable_area_sqm,
                                  estate_size_sqm    = excluded.estate_size_sqm,
                                  price              = excluded.price,
                                  price_display      = excluded.price_display,
                                  price_per_sqm      = excluded.price_per_sqm,
                                  rent_per_month     = excluded.rent_per_month,
                                  org_id             = excluded.org_id,
                                  org_name           = excluded.org_name,
                                  is_private         = excluded.is_private,
                                  main_image         = excluded.main_image,
                                  seo_url            = excluded.seo_url,
                                  published_at       = excluded.published_at,
                                  attributes         = excluded.attributes,
                                  search_path        = excluded.search_path,
                                  ad_uuid            = excluded.ad_uuid,
                                  start_date         = excluded.start_date,
                                  end_date           = excluded.end_date,
                                  first_published_at = excluded.first_published_at,
                                  created_at         = excluded.created_at,
                                  changed_at         = excluded.changed_at,
                                  municipality       = excluded.municipality,
                                  canonical_url      = excluded.canonical_url,
                                  images             = excluded.images,
                                  last_seen_at       = now();

-- name: GetListing :one
SELECT *
FROM immo_listings
WHERE ad_id = sqlc.arg(ad_id);

-- name: ListListings :many
SELECT *
FROM immo_listings
WHERE sqlc.narg(search_path)::varchar IS NULL
   OR search_path = sqlc.narg(search_path)
ORDER BY last_seen_at DESC, ad_id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountListings :one
SELECT count(*)
FROM immo_listings
WHERE sqlc.narg(search_path)::varchar IS NULL
   OR search_path = sqlc.narg(search_path);

-- name: DeleteListing :exec
DELETE
FROM immo_listings
WHERE ad_id = sqlc.arg(ad_id);


SELECT l.price_display,
       CASE d.own_age_type
           WHEN 'Miete' THEN true -- rent
           WHEN 'Pacht'
               THEN true -- lease, treated as rent-like rather than a sale
           WHEN 'Kauf' THEN false -- sale
           ELSE NULL -- attribute missing/unrecognized
           END AS is_for_rent,
       l.property_type,
       phone,
       phone2,
       l.number_of_rooms,
       l.living_area_sqm,
       l.estate_size_sqm,
       l.address,
       l.location,
       l.district,
       l.state,
       l.postcode,
       l.canonical_url,
       l.published_at,
       l.*
FROM immo_listings AS l
         LEFT JOIN LATERAL (
    SELECT max(attr -> 'values' ->> 0) FILTER (WHERE attr ->> 'name' = 'DESCRIPTION')         AS description_text,
           max(attr -> 'values' ->> 0) FILTER (WHERE attr ->> 'name' = 'DEALER')              AS dealer_flag,
           max(attr -> 'values' ->> 0) FILTER (WHERE attr ->> 'name' = 'CONTACT/COMPANYNAME') AS company_name,
           max(attr -> 'values' ->> 0) FILTER (WHERE attr ->> 'name' = 'OWNAGETYPE')          AS own_age_type,
           max(attr -> 'values' ->> 0) FILTER (WHERE attr ->> 'name' = 'CONTACT/PHONE')       AS phone,
           max(attr -> 'values' ->> 0) FILTER (WHERE attr ->> 'name' = 'CONTACT/PHONE2')      AS phone2,
           max(attr -> 'values' ->> 0) FILTER (WHERE attr ->> 'name' = 'categorytreeids')     AS category_tree_id
    FROM jsonb_array_elements(l.attributes -> 'attribute') AS attr
    ) AS d ON TRUE
WHERE is_private = true
  and
  -- exclude listings whose title or description says agents aren't welcome
    l.heading !~*
    ('makler\w*.{0,70}(unerw[uü]nscht|nicht\s*(er|ge)?w[uü]nscht|nicht\s*beantwortet|nicht\s*entgegengenommen|ausgeschlossen|zu\s*unterlassen|ab\s*zu\s*sehen|abzusehen|absehen|verzichten|bitte\s*(um\s*)?(abstand|absehen|verzichten))' ||
     '|(unerw[uü]nscht|nicht\s*(er|ge)?w[uü]nscht|nicht\s*beantwortet|nicht\s*entgegengenommen|ausgeschlossen|zu\s*unterlassen|ab\s*zu\s*sehen|abzusehen|absehen|verzichten|bitte\s*(um\s*)?(abstand|absehen|verzichten)).{0,70}makler\w*' ||
     '|(kein\w*|keinerlei).{0,70}(anfrage\w*|kontakt\w*|anruf\w*|angebot\w*|interesse\w*|telefonat\w*).{0,70}makler\w*' ||
     '|(kein\w*|keinerlei).{0,70}makler\w*.{0,70}(anfrage\w*|kontakt\w*|anruf\w*|angebot\w*|interesse\w*|telefonat\w*)' ||
     '|ohne\s*makler|nicht\s*an\s*makler\w*')
  AND COALESCE(d.description_text, '') !~*
      ('makler\w*.{0,70}(unerw[uü]nscht|nicht\s*(er|ge)?w[uü]nscht|nicht\s*beantwortet|nicht\s*entgegengenommen|ausgeschlossen|zu\s*unterlassen|ab\s*zu\s*sehen|abzusehen|absehen|verzichten|bitte\s*(um\s*)?(abstand|absehen|verzichten))' ||
       '|(unerw[uü]nscht|nicht\s*(er|ge)?w[uü]nscht|nicht\s*beantwortet|nicht\s*entgegengenommen|ausgeschlossen|zu\s*unterlassen|ab\s*zu\s*sehen|abzusehen|absehen|verzichten|bitte\s*(um\s*)?(abstand|absehen|verzichten)).{0,70}makler\w*' ||
       '|(kein\w*|keinerlei).{0,70}(anfrage\w*|kontakt\w*|anruf\w*|angebot\w*|interesse\w*|telefonat\w*).{0,70}makler\w*' ||
       '|(kein\w*|keinerlei).{0,70}makler\w*.{0,70}(anfrage\w*|kontakt\w*|anruf\w*|angebot\w*|interesse\w*|telefonat\w*)' ||
       '|ohne\s*makler|nicht\s*an\s*makler\w*')
  -- exclude listings that mention paying no agent commission ("keine Maklerprovision",
  -- "provisionsfrei", "Maklergebühr entfällt", etc.) — even without an explicit
  -- "agents unwelcome" statement, since this is a separate, related signal
  AND l.heading !~*
      ('makler\w*?(provision\w*|kosten\w*|geb[uü]hr\w*|courtage\w*|honorar\w*).{0,60}(kein\w*|keinerlei|ohne|frei\b|entfällt|gespart|sparen\w*|nicht\b)' ||
       '|(kein\w*|keinerlei|ohne|frei\b|entfällt|gespart|sparen\w*|nicht\b).{0,60}makler\w*?(provision\w*|kosten\w*|geb[uü]hr\w*|courtage\w*|honorar\w*)' ||
       '|(kein\w*|keinerlei|ohne|frei\b|entfällt|gespart|sparen\w*|nicht\b).{0,60}(provision\w*|kosten\w*|geb[uü]hr\w*|courtage\w*|honorar\w*).{0,60}makler\w*' ||
       '|provisionsfrei|maklerfrei')
  AND COALESCE(d.description_text, '') !~*
      ('makler\w*?(provision\w*|kosten\w*|geb[uü]hr\w*|courtage\w*|honorar\w*).{0,60}(kein\w*|keinerlei|ohne|frei\b|entfällt|gespart|sparen\w*|nicht\b)' ||
       '|(kein\w*|keinerlei|ohne|frei\b|entfällt|gespart|sparen\w*|nicht\b).{0,60}makler\w*?(provision\w*|kosten\w*|geb[uü]hr\w*|courtage\w*|honorar\w*)' ||
       '|(kein\w*|keinerlei|ohne|frei\b|entfällt|gespart|sparen\w*|nicht\b).{0,60}(provision\w*|kosten\w*|geb[uü]hr\w*|courtage\w*|honorar\w*).{0,60}makler\w*' ||
       '|provisionsfrei|maklerfrei')
  -- exclude listings that appear to be posted by a company/agency
  AND COALESCE(d.dealer_flag, '0') <> '1'
  AND COALESCE(d.company_name, '') !~*
      '\y(gmbh|e\.?u\.?|ohg|kg|ag|ges\.?m\.?b\.?h\.?|immobilien|makler|realit(ä|ae)t|invest|treuhand|hausverwaltung|bau[- ]?(und|&)|projekt(e)?)\y'
  -- keep only listings with a phone number or an email address in the description
  AND (
    NULLIF(d.phone, '') IS NOT NULL
        OR NULLIF(d.phone2, '') IS NOT NULL
        OR COALESCE(d.description_text, '') ~ '[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}'
    )
  -- exclude co-working listings, checked in both the title and the description
  AND l.heading !~* 'co[\s-]?working'
  AND COALESCE(d.description_text, '') !~* 'co[\s-]?working'
  -- exclude Gewerbeimmobilien: main commercial category, commercial land plots,
  -- and listings that are actually operating as a commercial/hospitality business
  AND COALESCE(d.category_tree_id, '') <> '7278'
  AND l.property_type NOT ILIKE '%gewerbe%'
  AND l.heading !~* 'gewerbe\s*gr[uü]nd\s*st[uü]ck'
  AND COALESCE(d.description_text, '') !~*
      '(gewerbefläche|gewerbevertr|gastgewerbefläche|gewerberechtliche\s*genehmigung|beherbergungsbetrieb)'