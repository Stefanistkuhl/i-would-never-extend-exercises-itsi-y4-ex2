-- Update queries

-- name: MarkCaptureAsArchived :exec
UPDATE captures
SET archived = 1, updated_at = current_timestamp
WHERE id = ?;

-- name: MarkCaptureAsCompressed :exec
UPDATE captures
SET compressed = 1, updated_at = current_timestamp
WHERE id = ?;

-- name: UpdateFilePath :exec
UPDATE captures
SET file_path = ?, updated_at = current_timestamp
WHERE id = ?;

