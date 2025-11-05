-- name: GetUserByGithubUserNodeID :one
SELECT 
    *
FROM 
    users u
INNER JOIN github_users ghu
    ON u.id=ghu.user_id
WHERE ghu.node_id=$1;

-- name: UpsertUser :exec
INSERT INTO users (id, email, name)
VALUES ($1, $2, $3)
ON CONFLICT (id)
DO UPDATE
SET email = EXCLUDED.email,
    name = EXCLUDED.name;

-- name: UpsertGithubUser :exec
INSERT INTO github_users (user_id, node_id)
VALUES ($1, $2)
ON CONFLICT (user_id)
DO UPDATE
SET node_id = EXCLUDED.node_id;