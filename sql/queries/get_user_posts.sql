-- name: GetPostsForUser :many
SELECT posts.* FROM posts
INNER JOIN feeds
ON posts.feed_id = feeds.id
WHERE feeds.user_id = $1
ORDER BY published_at DESC
LIMIT $2;