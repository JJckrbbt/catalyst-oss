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

-- name: ListCommentsForItem :many
SELECT
	c.id,
	c.comment,
	c.created_at,
	u.display_name,
	-- Aggregate mentioned user IDs and names into JSON array
	(
		SELECT COALESCE(json_agg(json_build_object('user_id', mu.id, 'display_name', mu.display_name)), '[]')
		FROM comment_mentions cm
		JOIN users mu ON cm.user_id = mu.id
		WHERE cm.comment_id = c.id
	) AS mentioned_users
FROM
	comments c
JOIN
	users u ON c.user_id = u.id
WHERE
	c.item_id = $1
ORDER BY
	c.created_at ASC;


