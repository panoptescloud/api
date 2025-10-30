-- name: GetUserByGithubUserNodeID :one
SELECT 
    *
FROM 
    users u
INNER JOIN github_users ghu
    ON u.id=ghu.user_id
WHERE ghu.node_id=$1;