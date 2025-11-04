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

-- name: InsertCaptureStats :exec
INSERT INTO capture_stats (
    capture_id,
    packet_count,
    protocol_distribution,
    top_src_ips,
    top_dst_ips,
    top_tcp_src_ports,
    top_tcp_dst_ports,
    top_udp_src_ports,
    top_udp_dst_ports,
    packet_rate,
    avg_packet_size,
    duration_seconds,
    first_packet_time,
    last_packet_time
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: InsertCapture :one
INSERT INTO captures (
    hostname,
    scenario,
    capture_datetime,
    file_path,
    file_size,
    compressed,
    archived,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING id;


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

-- name: MarkCaptureAsArchived :exec
UPDATE captures
SET archived = 1, updated_at = current_timestamp
WHERE id = ?;

-- name: GetPendingCompressions :many
SELECT id, file_path, file_size
FROM captures
WHERE compressed = 0
  AND archived = 0
ORDER BY capture_datetime ASC
LIMIT ?;

-- name: MarkCaptureAsCompressed :exec
UPDATE captures
SET compressed = 1, updated_at = current_timestamp
WHERE id = ?;

-- name: UpdateFilePath :exec
UPDATE captures
SET file_path = ?, updated_at = current_timestamp
WHERE id = ?;

-- name: GetOldArchivedCaptures :many
SELECT id, file_path
FROM captures
WHERE archived = 1
  AND capture_datetime < datetime('now', ?)
ORDER BY capture_datetime ASC;

-- name: DeleteCapture :exec
DELETE FROM captures
WHERE id = ?;


