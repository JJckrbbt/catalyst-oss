-- name: UpdateItem :one
-- Updates the mutable fields of a specific item
UPDATE items
SET
	scope = $2,
	status = $3,
	custom_properties = $4,
	updated_at = NOW()
WHERE
	id = $1
RETURNING *;

-- name: UpdateUser :one
-- Updates a user's mutable details
UPDATE "users"
SET
	display_name = $2,
	is_active = $3
WHERE
	id = $1
RETURNING *;

-- name: UpdateIngestionJobStatus :exec
-- Updates the status and details of an ingestion job
UPDATE ingestion_jobs
SET
	status = $2,
	completed_at = NOW(),
	error_details = $3,
	rows_upserted = $4,
	rows_triaged = $5
WHERE
	id = $1;

-- name: SetCommentEmbedding :exec
-- Sets the embedding for a specific comment after its been created
UPDATE comments
SET
	embedding = $2
WHERE
	id = $1;


