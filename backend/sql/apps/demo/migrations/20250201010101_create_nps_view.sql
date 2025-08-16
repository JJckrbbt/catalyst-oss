-- +goose Up
-- This view creates a structured, application-specific representation of NPS visitation data
-- from the generic 'items' table.

CREATE OR REPLACE VIEW vw_nps_visitation AS
SELECT
    -- core item properties
    item.id,
    item.item_type,
    item.scope AS state_code,
    item.status,
    item.created_at,
    item.updated_at,

    -- Unpacked visitation-specific properties from JSONB
    (item.custom_properties->>'park_name')::VARCHAR AS park_name,
    (item.custom_properties->>'year')::INTEGER AS year,
    (item.custom_properties->>'visitor_count')::BIGINT AS visitor_count,
    (item.custom_properties->>'notes')::TEXT AS notes

FROM
    items AS item
WHERE
    item.item_type = 'PARK_VISITATION';

-- +goose Down
DROP VIEW IF EXISTS vw_nps_visitation;
