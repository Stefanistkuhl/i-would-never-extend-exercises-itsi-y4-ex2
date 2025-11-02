-- name: GetConfig :one
SELECT * FROM config LIMIT 1;

-- name: UpdateConfig :exec
UPDATE config
SET watch_dir = ?,
organized_dir = ?,
archive_dir = ?,
compression_enabled = ?,
archive_days = ?,
max_retention_days = ?,
cleanup_interval_hours = ?,
batch_size = ?,
log_level = ?,
updated_at = current_timestamp;

-- name: InsertCaptureStats :exec
INSERT INTO capture_stats (
    capture_id,
    packet_count,
    protocol_distribution,
    top_src_ips,
    top_dst_ips,
    top_ports,
    packet_rate,
    avg_packet_size,
    tls_versions,
    dns_queries,
    duration_seconds,
    first_packet_time,
    last_packet_time
    ) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
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
