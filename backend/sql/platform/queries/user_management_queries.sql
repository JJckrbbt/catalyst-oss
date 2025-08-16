-- name: ListRoles :many
-- Fetch all available roles in system
SELECT id, name, description FROM "roles" ORDER BY id;

-- name: AssignRoleToUser :exec
-- Assign a specific role to a user
INSERT INTO "user_roles" (user_id, role_id) VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: RemoveRoleFromUser :exec
-- Removes a specific role from a user
DELETE FROM "user_roles" WHERE user_id = $1 AND role_id = $2;

-- name: AssignScopeToUser :exec
-- Grants a user access to a specific scope
INSERT INTO "user_scope_access" (user_id, scope) VALUES ($1, $2)
ON CONFLICT (user_id, scope) DO NOTHING;

-- name: RemoveScopeFromUser :exec
--Revokes a user's access from a specific scope.
DELETE FROM "user_scope_access" WHERE user_id = $1 AND scope = $2;

-- name: RemoveAllRolesFromUser :exec
-- Removes all roles from a user. Useful when completely re-assigning roles
DELETE FROM "user_roles" WHERE user_id = $1;

-- name: RemoveAllScopesFromUser :exec
-- Removes all scope access from a user
DELETE FROM "user_scope_access" WHERE user_id = $1;



-- name: SetUserAdminStatus :one
-- Updates only the is_admin status of a specific user
-- This is a priviliged action and should be protected at API layer
UPDATE "users"
SET
	is_admin = $2
WHERE
	id = $1
RETURNING *;

