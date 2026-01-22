-- Migration: Create import_jobs table for CSV import tracking
-- Created: 2026-01-17

-- Create import_jobs table
CREATE TABLE IF NOT EXISTS import_jobs (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,  -- clientes, polizas, recibos, siniestros
    mode VARCHAR(50) NOT NULL,  -- add, update, replace
    status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, processing, completed, failed, cancelled
    total_rows INTEGER NOT NULL DEFAULT 0,
    processed_rows INTEGER NOT NULL DEFAULT 0,
    successful_rows INTEGER NOT NULL DEFAULT 0,
    failed_rows INTEGER NOT NULL DEFAULT 0,
    duplicate_rows INTEGER NOT NULL DEFAULT 0,
    skipped_rows INTEGER NOT NULL DEFAULT 0,
    errors JSONB,  -- Array of error objects
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    validate_first BOOLEAN DEFAULT FALSE,
    duplicate_handling VARCHAR(50) DEFAULT 'skip',  -- skip, update, error
    username VARCHAR(255),
    filename VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_import_jobs_status ON import_jobs(status);
CREATE INDEX IF NOT EXISTS idx_import_jobs_type ON import_jobs(type);
CREATE INDEX IF NOT EXISTS idx_import_jobs_started_at ON import_jobs(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_import_jobs_username ON import_jobs(username);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_import_jobs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for updated_at
DROP TRIGGER IF EXISTS trigger_update_import_jobs_updated_at ON import_jobs;
CREATE TRIGGER trigger_update_import_jobs_updated_at
    BEFORE UPDATE ON import_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_import_jobs_updated_at();

-- Add comments
COMMENT ON TABLE import_jobs IS 'Tracks CSV import jobs with progress and error details';
COMMENT ON COLUMN import_jobs.id IS 'Unique job identifier (UUID)';
COMMENT ON COLUMN import_jobs.type IS 'Type of data being imported: clientes, polizas, recibos, siniestros';
COMMENT ON COLUMN import_jobs.mode IS 'Import mode: add (new only), update (existing only), replace (all)';
COMMENT ON COLUMN import_jobs.status IS 'Current status: pending, processing, completed, failed, cancelled';
COMMENT ON COLUMN import_jobs.errors IS 'JSON array of error objects with row, field, message, value';
COMMENT ON COLUMN import_jobs.duplicate_handling IS 'How to handle duplicates: skip, update, error';

-- Add import_id column to existing tables for tracking
-- This allows reverting imports by marking records as inactive

ALTER TABLE clientes ADD COLUMN IF NOT EXISTS import_id VARCHAR(255);
ALTER TABLE clientes ADD COLUMN IF NOT EXISTS creado_en TIMESTAMP DEFAULT NOW();
ALTER TABLE clientes ADD COLUMN IF NOT EXISTS actualizado_en TIMESTAMP DEFAULT NOW();

ALTER TABLE polizas ADD COLUMN IF NOT EXISTS import_id VARCHAR(255);
ALTER TABLE polizas ADD COLUMN IF NOT EXISTS creado_en TIMESTAMP DEFAULT NOW();
ALTER TABLE polizas ADD COLUMN IF NOT EXISTS actualizado_en TIMESTAMP DEFAULT NOW();

ALTER TABLE recibos ADD COLUMN IF NOT EXISTS import_id VARCHAR(255);
ALTER TABLE recibos ADD COLUMN IF NOT EXISTS creado_en TIMESTAMP DEFAULT NOW();
ALTER TABLE recibos ADD COLUMN IF NOT EXISTS actualizado_en TIMESTAMP DEFAULT NOW();

ALTER TABLE siniestros ADD COLUMN IF NOT EXISTS import_id VARCHAR(255);
ALTER TABLE siniestros ADD COLUMN IF NOT EXISTS creado_en TIMESTAMP DEFAULT NOW();
ALTER TABLE siniestros ADD COLUMN IF NOT EXISTS actualizado_en TIMESTAMP DEFAULT NOW();

-- Create indexes on import_id for reverting
CREATE INDEX IF NOT EXISTS idx_clientes_import_id ON clientes(import_id) WHERE import_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_polizas_import_id ON polizas(import_id) WHERE import_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_recibos_import_id ON recibos(import_id) WHERE import_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_siniestros_import_id ON siniestros(import_id) WHERE import_id IS NOT NULL;

-- Sample query to get import job statistics
-- SELECT
--     id,
--     type,
--     status,
--     total_rows,
--     successful_rows,
--     failed_rows,
--     ROUND((successful_rows::NUMERIC / NULLIF(total_rows, 0) * 100), 2) as success_rate,
--     started_at,
--     completed_at,
--     EXTRACT(EPOCH FROM (COALESCE(completed_at, NOW()) - started_at)) as duration_seconds
-- FROM import_jobs
-- ORDER BY started_at DESC
-- LIMIT 10;
