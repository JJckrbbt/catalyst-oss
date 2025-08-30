-- +goose Up

-- A view for unpacking policyholder data from the items table's JSONB column.
CREATE OR REPLACE VIEW vw_policyholders AS
SELECT
    -- Core item properties for direct access
    item.id,
    item.item_type,
    item.business_key AS policyholder_id,
    item.scope AS state,
    item.status,
    item.created_at,
    item.updated_at,

    -- Unpacked policyholder-specific properties from the JSONB field with type casting
    (item.custom_properties->>'PolicyHolder_Name')::VARCHAR AS policyholder_name,
    (item.custom_properties->>'City')::VARCHAR AS city,
    (item.custom_properties->>'Customer_Since_Date')::DATE AS customer_since_date,
    (item.custom_properties->>'Customer_Level')::VARCHAR AS customer_level,
    (item.custom_properties->>'Active_Policies')::JSONB AS active_policies
FROM
    items AS item
WHERE
    item.item_type = 'POLICYHOLDER';


-- A view for unpacking insurance claim data from the items table's JSONB column.
CREATE OR REPLACE VIEW vw_insurance_claims AS
SELECT
    -- Core item properties
    item.id,
    item.item_type,
    item.business_key AS claim_id,
    item.scope AS policy_number,
    item.status AS system_status,
    item.embedding,
    item.created_at,
    item.updated_at,

    -- Unpacked claim-specific properties from the JSONB field with type casting
    (item.custom_properties->>'PolicyHolder_ID')::VARCHAR AS policyholder_id,
    (item.custom_properties->>'Claim_Type')::VARCHAR AS claim_type,
    (item.custom_properties->>'Date_of_Loss')::DATE AS date_of_loss,
    (item.custom_properties->>'Description_of_Loss')::TEXT AS description_of_loss,
    (item.custom_properties->>'Claim_Amount')::DECIMAL(12, 2) AS claim_amount,
    (item.custom_properties->>'Status')::VARCHAR AS business_status,
    (item.custom_properties->>'Adjuster_Assigned')::VARCHAR AS adjuster_assigned
FROM
    items AS item
WHERE
    item.item_type = 'INSURANCE_CLAIM';


-- +goose Down
DROP VIEW IF EXISTS vw_insurance_claims;
DROP VIEW IF EXISTS vw_policyholders;
