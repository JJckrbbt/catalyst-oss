-- name: ListClaimsWithVector :many
-- Fetches and sorts claims by semantic similarity.
SELECT
    id, item_type, claim_id, policy_number, system_status, created_at, updated_at,
    policyholder_id, claim_type, date_of_loss, description_of_loss, claim_amount,
    business_status, adjuster_assigned,
    (embedding <=> @search_embedding::vector) as similarity_score
FROM vw_insurance_claims
WHERE
    (sqlc.narg('claim_id')::text IS NULL OR claim_id = sqlc.arg('claim_id'))
AND (sqlc.narg('min_amount')::decimal IS NULL OR claim_amount >= sqlc.narg('min_amount'))
AND (sqlc.narg('max_amount')::decimal IS NULL OR claim_amount <= sqlc.narg('max_amount'))
AND (sqlc.narg('adjuster_assigned')::text IS NULL OR adjuster_assigned = sqlc.arg('adjuster_assigned'))
AND (sqlc.narg('status')::text IS NULL OR business_status = sqlc.arg('status'))
AND (sqlc.narg('policy_number')::text IS NULL OR policy_number = sqlc.arg('policy_number'))
AND (embedding <=> @search_embedding::vector) < 0.8
ORDER BY similarity_score ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListClaimsWithoutVector :many
-- Fetches a paginated and filtered list of insurance claims without vector search.
SELECT
    id, item_type, claim_id, policy_number, system_status, created_at, updated_at,
    policyholder_id, claim_type, date_of_loss, description_of_loss, claim_amount,
    business_status, adjuster_assigned,
    NULL::float8 as similarity_score
FROM vw_insurance_claims
WHERE
    (sqlc.narg('claim_id')::text IS NULL OR claim_id = sqlc.arg('claim_id'))
AND (sqlc.narg('min_amount')::decimal IS NULL OR claim_amount >= sqlc.narg('min_amount'))
AND (sqlc.narg('max_amount')::decimal IS NULL OR claim_amount <= sqlc.narg('max_amount'))
AND (sqlc.narg('adjuster_assigned')::text IS NULL OR adjuster_assigned = sqlc.arg('adjuster_assigned'))
AND (sqlc.narg('status')::text IS NULL OR business_status = sqlc.arg('status'))
AND (sqlc.narg('policy_number')::text IS NULL OR policy_number = sqlc.arg('policy_number'))
ORDER BY
    CASE WHEN sqlc.arg(sort_by)::text = 'claim_amount' AND sqlc.arg(sort_direction)::text = 'asc' THEN claim_amount END ASC,
    CASE WHEN sqlc.arg(sort_by)::text = 'claim_amount' AND sqlc.arg(sort_direction)::text = 'desc' THEN claim_amount END DESC,
    date_of_loss DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: SearchKnowledgeChunks :many
-- Searches semantically the knowledge base
SELECT
    'Knowledge Chunk from ' || (custom_properties->'metadata'->>'document_name')::VARCHAR AS source,
    COALESCE((custom_properties->>'chunk_text')::TEXT, '') AS text,
    embedding <=> @embedding::vector AS similarity_score,
    custom_properties->'metadata'->'source_custom_properties' AS structured_metadata
FROM items
WHERE
    item_type = 'KNOWLEDGE_CHUNK' AND embedding IS NOT NULL
ORDER BY similarity_score ASC
LIMIT sqlc.arg('limit');

-- name: SearchComments :many
-- Searches comments semantically.
SELECT
    'Comment' AS source,
    comment::TEXT AS text,
    embedding <=> @embedding::vector AS similarity_score
FROM comments
WHERE embedding IS NOT NULL
ORDER BY similarity_score ASC
LIMIT sqlc.arg('limit');

-- name: GetDocumentHeader :one
-- Fetches the header chunk's source_custom_properties for a given document ID.
SELECT
    custom_properties->'metadata'->'source_custom_properties' as structured_metadata
FROM items
WHERE
    item_type = 'KNOWLEDGE_CHUNK'
AND
    custom_properties->'metadata'->>'document_id' = @document_id
AND
    custom_properties->'metadata'->>'chunk_number' = '0'
LIMIT 1;

-- name: GetClaimDetails :one
-- Fetches a single claim joined with its correspondng policyholder data
SELECT
    c.id, c.item_type, c.claim_id, c.policy_number, c.system_status, c.created_at, c.updated_at,
    c.policyholder_id, c.claim_type, c.date_of_loss, c.description_of_loss, c.claim_amount,
    c.business_status, c.adjuster_assigned, p.policyholder_name, p.city, p.state,
    p.customer_since_date, p.customer_level
FROM vw_insurance_claims c
JOIN vw_policyholders p ON c.policyholder_id = p.policyholder_id
WHERE c.id = $1;

-- name: GetClaimStatusHistory :many
-- Fetches the business status change history for a specific claim item
SELECT
    ie.id AS event_id, ie.created_at AS event_timestamp, ie.event_data, u.display_name AS user_name
FROM items_events ie
JOIN users u ON ie.created_by = u.id
WHERE ie.item_id = $1 AND ie.event_type = 'CLAIM_STATUS_CHANGED'
ORDER BY ie.created_at DESC;

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
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');
