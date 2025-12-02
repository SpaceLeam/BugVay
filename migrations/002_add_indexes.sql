-- Apply proper indexes and constraints to existing schema

-- PROGRAMS table already created by user
CREATE INDEX IF NOT EXISTS idx_programs_name ON programs(name);

-- ASSETS table already created by user
-- Adding composite index for common queries
CREATE INDEX IF NOT EXISTS idx_assets_program_domain ON assets(program_id, domain);
CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(type);

-- ENDPOINTS table already created by user
-- Adding index for canonical URL lookups
CREATE INDEX IF NOT EXISTS idx_endpoints_canonical ON endpoints(canonical_url);
CREATE INDEX IF NOT EXISTS idx_endpoints_crawled ON endpoints(crawled) WHERE crawled = false;

-- FINDINGS table already created by user
-- Adding composite index for filtering
CREATE INDEX IF NOT EXISTS idx_findings_status_severity ON findings(status, severity);
CREATE INDEX IF NOT EXISTS idx_findings_created_at ON findings(created_at DESC);

-- Add missing columns if not exists
ALTER TABLE findings ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMP;
ALTER TABLE findings ADD COLUMN IF NOT EXISTS false_positive BOOLEAN DEFAULT FALSE;

-- Create index for hash-based deduplication
CREATE INDEX IF NOT EXISTS idx_endpoints_hash ON endpoints(hash);

-- Performance: partial index for active findings
CREATE INDEX IF NOT EXISTS idx_findings_active ON findings(endpoint_id, severity) 
WHERE status IN ('new', 'triaged', 'verified');

COMMENT ON TABLE programs IS 'Bug bounty programs with scope definitions';
COMMENT ON TABLE assets IS 'In-scope assets (domains, wildcards, URLs) per program';
COMMENT ON TABLE endpoints IS 'Discovered endpoints ready for scanning';
COMMENT ON TABLE findings IS 'Security findings from scanner modules';
