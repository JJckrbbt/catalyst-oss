-- +goose Up
-- Create views for both structured and unstructured Apollo mission data.

-- A view for the structured mission facts
CREATE OR REPLACE VIEW vw_apollo_mission_facts AS
SELECT
    -- core item properties
    item.id,
    item.item_type,
    item.business_key,
    
    -- Unpacked mission-specific properties from JSONB
    (item.custom_properties->>'Mission')::VARCHAR AS mission_name,
    (item.custom_properties->>'Commander')::VARCHAR AS commander,
    (item.custom_properties->>'LunarModulePilot')::VARCHAR AS lunar_module_pilot,
    (item.custom_properties->>'LaunchDate')::DATE AS launch_date,
    (item.custom_properties->>'LandingSite')::VARCHAR AS landing_site

FROM
    items AS item
WHERE
    item.item_type = 'MISSION_FACTS';

-- A view for the unstructured knowledge chunks from mission reports
CREATE OR REPLACE VIEW vw_apollo_mission_knowledge AS
SELECT
    item.id,
    item.item_type,
    item.business_key,
    item.embedding,

    -- Unpacked text content from JSONB
    (item.custom_properties->>'chunk_text')::TEXT AS chunk_text
FROM
    items AS item
WHERE
    item.item_type = 'KNOWLEDGE_CHUNK';


-- +goose Down
DROP VIEW IF EXISTS vw_apollo_mission_knowledge;
DROP VIEW IF EXISTS vw_apollo_mission_facts;
