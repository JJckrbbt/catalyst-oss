-- name: CreateIngestionError :one
-- Inserts a new ingestion error record for a row that failed processing.
INSERT INTO ingestion_errors (
    id,
    job_id,
    original_row_data,
    reason_for_failure
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ItemExistsByBusinessKey :one
-- Checks for the existence of an item by its type and business key. Returns 1 if it exists, 0 otherwise.
SELECT EXISTS(SELECT 1 FROM items WHERE item_type = $1 AND business_key = $2)::int;
