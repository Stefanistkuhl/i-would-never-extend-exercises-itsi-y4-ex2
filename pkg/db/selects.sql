-- Select queries

-- name: GetCapturesForArchive :many
SELECT id, file_path, file_size
FROM captures
WHERE archived = 0
  AND capture_datetime < datetime('now', ?)
ORDER BY capture_datetime ASC;

-- name: GetCapturesForArchiveWithLimit :many
SELECT id, file_path
FROM captures
WHERE archived = 0
  AND capture_datetime < datetime('now', ?)
ORDER BY capture_datetime ASC;

-- name: GetPendingCompressions :many
SELECT id, file_path, file_size
FROM captures
WHERE compressed = 0
  AND archived = 0
ORDER BY capture_datetime ASC
LIMIT ?;

-- name: GetOldArchivedCaptures :many
SELECT id, file_path
FROM captures
WHERE archived = 1
  AND capture_datetime < datetime('now', ?)
ORDER BY capture_datetime ASC;

-- name: GetCapture :one
SELECT * FROM captures WHERE id = ?;

-- name: GetCaptures :many
SELECT * FROM captures;
