ALTER TABLE research_results ADD COLUMN result_type VARCHAR(50);

-- Add data column to research_results table
ALTER TABLE research_results
    ADD COLUMN IF NOT EXISTS data JSONB;

-- Or if you prefer 'findings' as the column name (which the fixed code tries first):
ALTER TABLE research_results
    ADD COLUMN IF NOT EXISTS findings JSONB;