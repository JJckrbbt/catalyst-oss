-- backend/sql/apps/insurance/queries/insurance_queries.sql

-- name: ListClaims :many
SELECT *
FROM vw_insurance_claims
WHERE
	(@adjuster_assigned::TEXT IS NULL OR adjuster_assigned = @adjuster_assigned)
AND
	(@status::TEXT IS NULL OR status = @status)
AND
	(@policy_number::TEXT IS NULL OR policy_number = @policy_number)
ORDER BY
	date_of_loss DESC
LIMIT $1
OFFSET $2;

-- name: GetPolicyholderBYID :one
SELECT *
FROM vw_policyholders
WHERE policyholder_id = $1;

-- name: ListPolicyholders :many
SELECT *
FROM vw_policyholders
WHERE
	(@state::TEXT IS NULL OR state = @state)
AND
	(@customer_level::TEXT IS NULL OR customer_level = @customer_level)
ORDER BY
	policyholder_name
LIMIT $1
OFFSET $2;



