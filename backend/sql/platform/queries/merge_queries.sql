-- name: DeactivateItemsBySource :exec
UPDATE items SET status = 'inactive'
WHERE item_type= $1 AND custom_properties->>'reporting_source' = $2;

-- name: UpsertItems :execrows
--Insert new records from staging, or update existing ones based on business key
INSERT INTO items (
	item_type, scope, business_key, status, custom_properties, embedding
)
SELECT
	item_type,
	scope, 
	business_key,
	'active',
	custom_properties,
	embedding
FROM temp_items_staging
ON CONFLICT (item_type, business_key) DO UPDATE SET 
	status = EXCLUDED.status,
	scope = EXCLUDED.scope,
	custom_properties = items.custom_properties || EXCLUDED.custom_properties,
	embedding = EXCLUDED.embedding,
	updated_at = NOW();
