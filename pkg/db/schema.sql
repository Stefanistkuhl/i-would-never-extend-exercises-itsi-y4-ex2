create table captures (
	id integer primary key autoincrement,
	hostname text not null,
	scenario text not null,
	capture_datetime datetime not null,
	file_path text unique not null,
	file_size integer not null,
	compressed boolean default 0,
	archived boolean default 0,
	created_at datetime default current_timestamp,
	updated_at datetime default current_timestamp
);

create table capture_stats (
	id integer primary key autoincrement,
	packet_count integer,
	capture_id integer not null unique,
	protocol_distribution text,  -- JSON: {"TCP": 45, "UDP": 30, "ICMP": 5}
	top_src_ips text,             -- JSON: ["192.168.1.10", "10.0.0.5", ...]
	top_dst_ips text,             -- JSON: ["8.8.8.8", "1.1.1.1", ...]
	top_ports text,               -- JSON: {"443": 150, "80": 120, "22": 45}
	packet_rate real,             -- packets/sec
	avg_packet_size real,
	tls_versions text,            -- JSON: {"TLS1.2": 100, "TLS1.3": 50}
	dns_queries integer,
	duration_seconds integer,
	first_packet_time datetime,
	last_packet_time datetime,
	created_at datetime default current_timestamp,
	foreign key(capture_id) references captures(id) on delete cascade
);

create table config (

	watch_dir text default './data/captures/incoming',
	organized_dir text default './data/captures/organized',
	archive_dir text default './data/captures/archive',
	compression_enabled boolean default 1,
	archive_days integer default 30,
	max_retention_days integer default 90,
	cleanup_interval_hours integer default 24,
	batch_size integer default 10,
	log_level text default 'info',
	updated_at datetime default current_timestamp

);

create index idx_captures_hostname on captures(hostname);
create index idx_captures_scenario on captures(scenario);
create index idx_captures_datetime on captures(capture_datetime);
create index idx_captures_archived on captures(archived);
create index idx_capture_stats_capture_id on capture_stats(capture_id);

insert or ignore into config default values;
