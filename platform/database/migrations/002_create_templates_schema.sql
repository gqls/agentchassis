-- FILE: platform/database/migrations/002_create_templates_schema.sql
-- Templates database schema
CREATE TABLE IF NOT EXISTS persona_templates (
                                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX idx_templates_category ON persona_templates(category);
CREATE INDEX idx_templates_active ON persona_templates(is_active);
