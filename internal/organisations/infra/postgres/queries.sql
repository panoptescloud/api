-- name: GetOrganisationByID :one
SELECT 
    *
FROM 
    organisations o
WHERE o.id=$1;

-- name: GetOrganisationMembers :many
SELECT * FROM organisation_members WHERE organisation_id = $1;

-- name: UpsertOrganisation :exec
INSERT INTO organisations (id, name)
VALUES ($1, $2)
ON CONFLICT (id)
DO UPDATE
SET name = EXCLUDED.name;

-- name: UpsertOrganisationMember :exec
INSERT INTO organisation_members (organisation_id, member_id, "role")
VALUES ($1, $2, $3)
ON CONFLICT (organisation_id, member_id)
DO UPDATE
SET "role" = EXCLUDED.role;

-- name: GetOrganisationsForMember :many
SELECT
    o.*
FROM organisations o 
INNER JOIN organisation_members om
    ON o.id=om.organisation_id
WHERE om.member_id = $1;

-- name: UpsertAPIKey :exec
INSERT INTO organisation_api_keys (id, organisation_id, name, token)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id)
DO UPDATE
-- name is the only property that may be updated after creation
SET name = EXCLUDED.name;

-- name: APIKeyByToken :one
SELECT
    *
FROM organisation_api_keys oak
WHERE oak.token=$1;

-- name: APIKeyByID :one
SELECT
    *
FROM organisation_api_keys oak
WHERE oak.id=$1;

-- name: AllOrganisationAPIKeys :many
SELECT
    *
FROM organisation_api_keys oak
WHERE oak.organisation_id=$1;
