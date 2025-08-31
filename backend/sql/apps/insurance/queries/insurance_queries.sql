-- backend/sql/apps/insurance/queries/insurance_queries.sql

-- name: ListClaims :many
-- Fetches a paginated and filtered list of insurance claims.
SELECT 
    id,
    item_type,
    claim_id,
    policy_number,
    system_status,
    created_at,
    updated_at,
    policyholder_id,
    claim_type,
    date_of_loss,
    description_of_loss,
    claim_amount,
    business_status,
    adjuster_assigned
FROM vw_insurance_claims
WHERE
    adjuster_assigned = COALESCE(sqlc.narg('adjuster_assigned'), adjuster_assigned)
AND
    business_status = COALESCE(sqlc.narg('status'), business_status)
AND
    policy_number = COALESCE(sqlc.narg('policy_number'), policy_number)
ORDER BY
    date_of_loss DESC
LIMIT $1
OFFSET $2;

-- name: GetPolicyholderByID :one
-- Fetches a single policyholder by their unique PolicyHolder_ID.
SELECT *
FROM vw_policyholders
WHERE policyholder_id = $1;

-- name: ListPolicyholders :many
-- Fetches a paginated and filtered list of policyholders.
SELECT *
FROM vw_policyholders
WHERE
    state = COALESCE(sqlc.narg('state'), state)
AND
    customer_level = COALESCE(sqlc.narg('customer_level'), customer_level)
ORDER BY
    policyholder_name
LIMIT $1
OFFSET $2;

-- name: GetClaimDetails :one
-- Fetches a single claim joined with its correspondng policyholder data
SELECT
    c.id,
    c.item_type,
    c.claim_id,
    c.policy_number,
    c.system_status,
    c.created_at,
    c.updated_at,
    c.policyholder_id,
    c.claim_type,
    c.date_of_loss,
    c.description_of_loss,
    c.claim_amount,
    c.business_status,
    c.adjuster_assigned,
    p.policyholder_name,
    p.city,
    p.state,
    p.customer_since_date,
    p.customer_level
FROM
    vw_insurance_claims c
JOIN
    vw_policyholders p ON c.policyholder_id = p.policyholder_id
WHERE
    c.id = $1;

-- name: GetClaimStatusHistory :many
-- Fetches the business status change history for a specific claim item
SELECT
    ie.id AS event_id,
    ie.created_at AS event_timestamp,
    ie.event_data, 
    u.display_name AS user_name
FROM
    items_events ie
JOIN
    users u ON ie.created_by = u.id
WHERE
    ie.item_id = $1
AND
    ie.event_type = 'CLAIM_STATUS_CHANGED'
ORDER BY
    ie.created_at DESC;
