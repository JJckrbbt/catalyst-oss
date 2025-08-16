-- name: CreateItem :one
-- Inserts a new item record into database
-- Go is responsible for constructing the custom_properties JSONB
INSERT INTO items (
	item_type, 
	scope,
	business_key,
	status,
	custom_properties,
	embedding
) VALUES (
	$1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: CreateUserFromAuthProvider :one
-- Creates a new user record from the authentication provider's details
INSERT INTO "users" (
	auth_provider_subject,
	email,
	display_name,
	is_active,
	is_admin
) VALUES (
	$1, $2, $3, TRUE, FALSE
)
RETURNING *;

-- name: CreateIngestionJob :one
-- Inserts a new file ingestion job record.
INSERT INTO ingestion_jobs (
	id, 
	source_type,
	source_details,
	report_type,
	status, 
	user_id,
	source_uri
) VALUES (
	$1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: CreateTempItemsStagingTable :exec
-- Creates a temporary table for staging items during ingest
-- This table is dropped on commit
CREATE TEMP TABLE temp_items_staging (LIKE items INCLUDING DEFAULTS) ON COMMIT DROP;
