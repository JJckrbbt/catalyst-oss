-- backend/sql/apps/insurance/queries/insurance_queries.sql

-- name: ListClaims :many
-- Fetches a paginated and filtered list of insurance claims.
SELECT *
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
