-- name: GetUserByAuthProviderSubject :one
-- Fetch a single user by their external auth provider ID
SELECT * FROM "users" WHERE auth_provider_subject = $1;

-- name: GetEventsForItem :many
-- Fetch the event history for a specific item, newest first
SELECT * FROM "items_events"
WHERE item_id = $1
ORDER BY created_at DESC;

-- name: GetItemForUpdate :one
-- Fetch a single item for update
SELECT * FROM "items"
WHERE id = $1 LIMIT 1;
