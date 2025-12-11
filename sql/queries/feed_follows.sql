-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, feed_id, user_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT 
    iff.*,
    f.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow iff
INNER JOIN feeds f ON iff.feed_id = f.id
INNER JOIN users ON iff.user_id = users.id;

-- name: GetFeedFollowsForUser :many
SELECT 
    ff.*,
    f.name AS feed_name,
    u.name AS user_name
FROM feed_follows ff
INNER JOIN feeds f ON ff.feed_id = f.id
INNER JOIN users u ON ff.user_id = u.id
WHERE ff.user_id = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;