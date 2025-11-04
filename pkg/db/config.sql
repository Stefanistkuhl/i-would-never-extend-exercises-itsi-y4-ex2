-- Config queries

-- name: GetConfig :one
SELECT * FROM config LIMIT 1;

-- name: UpdateConfig :exec
UPDATE config
SET watch_dir = ?,
organized_dir = ?,
archive_dir = ?,
expose_service = ?,
port = ?,
compression_enabled = ?,
archive_days = ?,
max_retention_days = ?,
log_level = ?;

