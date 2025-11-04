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

-- name: GetArchviedCaptures :many
SELECT * FROM captures WHERE archived = 1;

-- name: GetArchiveBrief :many
SELECT id, file_path, file_size, created_at, updated_at FROM captures WHERE archived = 1;

-- name: GetCaptureStatsByID :one
SELECT cs.id, cs.packet_count, cs.capture_id, cs.protocol_distribution, cs.top_src_ips, cs.top_dst_ips, cs.top_tcp_src_ports, cs.top_tcp_dst_ports, cs.top_udp_src_ports, cs.top_udp_dst_ports, cs.packet_rate, cs.avg_packet_size, cs.duration_seconds, cs.first_packet_time, cs.last_packet_time, cs.created_at, c.hostname, c.scenario, c.capture_datetime, c.file_path
FROM captures c
LEFT JOIN capture_stats cs ON c.id = cs.capture_id
WHERE c.id = ?;

-- name: GetSummary :one
SELECT 
    COUNT(DISTINCT c.id) as total_captures,
    COUNT(DISTINCT cs.id) as captures_with_stats,
    SUM(c.file_size) as total_file_size,
    SUM(COALESCE(cs.packet_count, 0)) as total_packets,
    AVG(COALESCE(cs.packet_rate, 0)) as avg_packet_rate,
    AVG(COALESCE(cs.avg_packet_size, 0)) as avg_packet_size,
    SUM(COALESCE(cs.duration_seconds, 0)) as total_duration_seconds,
    COUNT(DISTINCT c.hostname) as unique_hostnames,
    COUNT(DISTINCT c.scenario) as unique_scenarios
FROM captures c
LEFT JOIN capture_stats cs ON c.id = cs.capture_id;

-- name: GetStatsByHostname :many
SELECT 
    c.hostname,
    COUNT(DISTINCT c.id) as capture_count,
    SUM(c.file_size) as total_file_size,
    SUM(COALESCE(cs.packet_count, 0)) as total_packets,
    AVG(COALESCE(cs.packet_rate, 0)) as avg_packet_rate,
    AVG(COALESCE(cs.avg_packet_size, 0)) as avg_packet_size,
    SUM(COALESCE(cs.duration_seconds, 0)) as total_duration_seconds
FROM captures c
LEFT JOIN capture_stats cs ON c.id = cs.capture_id
GROUP BY c.hostname
ORDER BY c.hostname;

-- name: GetStatsByScenario :many
SELECT 
    c.scenario,
    COUNT(DISTINCT c.id) as capture_count,
    SUM(c.file_size) as total_file_size,
    SUM(COALESCE(cs.packet_count, 0)) as total_packets,
    AVG(COALESCE(cs.packet_rate, 0)) as avg_packet_rate,
    AVG(COALESCE(cs.avg_packet_size, 0)) as avg_packet_size,
    SUM(COALESCE(cs.duration_seconds, 0)) as total_duration_seconds
FROM captures c
LEFT JOIN capture_stats cs ON c.id = cs.capture_id
GROUP BY c.scenario
ORDER BY c.scenario;

-- name: GetCapturesByHostname :many
SELECT * FROM captures WHERE hostname = ? ORDER BY capture_datetime DESC;

-- name: GetCapturesByScenario :many
SELECT * FROM captures WHERE scenario = ? ORDER BY capture_datetime DESC;
