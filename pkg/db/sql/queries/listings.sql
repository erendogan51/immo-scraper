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
                            search_path)
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
        sqlc.narg(search_path))
ON CONFLICT (ad_id) DO UPDATE SET vertical_id      = excluded.vertical_id,
                                   ad_type_id       = excluded.ad_type_id,
                                   product_id       = excluded.product_id,
                                   advert_status_id = excluded.advert_status_id,
                                   advert_status    = excluded.advert_status,
                                   description      = excluded.description,
                                   heading          = excluded.heading,
                                   property_type    = excluded.property_type,
                                   rooms_layout     = excluded.rooms_layout,
                                   number_of_rooms  = excluded.number_of_rooms,
                                   floor            = excluded.floor,
                                   address          = excluded.address,
                                   location         = excluded.location,
                                   district         = excluded.district,
                                   state            = excluded.state,
                                   country          = excluded.country,
                                   postcode         = excluded.postcode,
                                   latitude         = excluded.latitude,
                                   longitude        = excluded.longitude,
                                   living_area_sqm  = excluded.living_area_sqm,
                                   usable_area_sqm  = excluded.usable_area_sqm,
                                   estate_size_sqm  = excluded.estate_size_sqm,
                                   price            = excluded.price,
                                   price_display    = excluded.price_display,
                                   price_per_sqm    = excluded.price_per_sqm,
                                   rent_per_month   = excluded.rent_per_month,
                                   org_id           = excluded.org_id,
                                   org_name         = excluded.org_name,
                                   is_private       = excluded.is_private,
                                   main_image       = excluded.main_image,
                                   seo_url          = excluded.seo_url,
                                   published_at     = excluded.published_at,
                                   attributes       = excluded.attributes,
                                   search_path      = excluded.search_path,
                                   last_seen_at     = now();

-- name: GetListing :one
SELECT *
FROM immo_listings
WHERE ad_id = sqlc.arg(ad_id);

-- name: ListListings :many
SELECT *
FROM immo_listings
WHERE sqlc.narg(search_path)::varchar IS NULL OR search_path = sqlc.narg(search_path)
ORDER BY last_seen_at DESC, ad_id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountListings :one
SELECT count(*)
FROM immo_listings
WHERE sqlc.narg(search_path)::varchar IS NULL OR search_path = sqlc.narg(search_path);

-- name: DeleteListing :exec
DELETE
FROM immo_listings
WHERE ad_id = sqlc.arg(ad_id);
