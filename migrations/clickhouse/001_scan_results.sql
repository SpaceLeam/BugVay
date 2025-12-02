-- ClickHouse schema for massive scan result analytics
-- Using MergeTree engine with monthly partitioning for optimal performance

CREATE DATABASE IF NOT EXISTS bugvay;

CREATE TABLE IF NOT EXISTS bugvay.scan_results (
    id UUID DEFAULT generateUUIDv4(),
    timestamp DateTime DEFAULT now(),
    program_id UInt32,
    asset_id UInt32,
    endpoint_id UInt32,
    scanner LowCardinality(String),
    scan_type LowCardinality(String),
    
    -- Request data
    method LowCardinality(String),
    url String,
    payload String,
    headers String,
    
    -- Response data
    status_code UInt16,
    content_length UInt32,
    response_time_ms UInt16,
    body_snippet String,
    response_headers String,
    
    -- Detection
    vulnerable Bool,
    confidence Float32,
    evidence_hash FixedString(32),
    evidence String,
    
    -- Metadata
    worker_id String,
    created_at DateTime DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (program_id, scanner, endpoint_id, created_at)
TTL created_at + INTERVAL 6 MONTH
SETTINGS index_granularity = 8192;

-- Materialized view for real-time analytics
CREATE MATERIALIZED VIEW IF NOT EXISTS bugvay.scan_stats_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (program_id, scanner, hour)
AS SELECT
    program_id,
    scanner,
    toStartOfHour(created_at) AS hour,
    countIf(vulnerable) AS vuln_count,
    count() AS total_scans,
    avg(response_time_ms) AS avg_response_time,
    quantile(0.95)(response_time_ms) AS p95_response_time
FROM bugvay.scan_results
GROUP BY program_id, scanner, hour;

-- Index for fast evidence lookups
CREATE TABLE IF NOT EXISTS bugvay.evidence_index (
    evidence_hash FixedString(32),
    endpoint_id UInt32,
    scanner LowCardinality(String),
    first_seen DateTime,
    last_seen DateTime,
    occurrences UInt32
)
ENGINE = ReplacingMergeTree(last_seen)
ORDER BY (evidence_hash, endpoint_id, scanner);

COMMENT ON TABLE bugvay.scan_results IS 'Raw scan results for analytics (billions of rows)';
COMMENT ON TABLE bugvay.scan_stats_hourly IS 'Hourly aggregated scan statistics';
COMMENT ON TABLE bugvay.evidence_index IS 'Deduplication index for findings';
