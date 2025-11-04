-- Delete queries

-- name: DeleteCapture :exec
DELETE FROM captures
WHERE id = ?;

