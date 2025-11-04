-- Insert queries

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

