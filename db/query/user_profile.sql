-- name: CreateUserProfile :exec
INSERT INTO users (
    user_id, image_url, first_name, last_name
    )
VALUES ($1, $2, $3, $4);

-- name: GetUserProfileByUID :one
SELECT * FROM users WHERE user_id = $1 LIMIT 1;

-- name: UpdateImgUserProfile :exec
UPDATE users SET image_url = $2 WHERE user_id = $1;

-- name: DeleteUserProfileByID :exec
DELETE FROM users WHERE user_id = $1;

-- name: GetUserProfile :one
SELECT
  u.id,
  u.username,
  u.email,
  u.phone,
  u.created_at,
  u.is_verified,
  us.first_name,
  us.last_name,
  us.image_url
FROM
  authentications u
JOIN
  users us ON u.id = us.user_id
WHERE
  (u.username = $1 OR u.id::text = $1);