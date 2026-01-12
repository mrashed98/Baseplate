-- Enable UUID extension for generating IDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table: Blueprints
-- Stores the "Model" definitions (e.g., 'Service', 'Cluster').
-- The 'schema' column holds the JSON Schema validation rules.
CREATE TABLE blueprints (
    id VARCHAR(50) PRIMARY KEY, -- e.g., 'service', 'employee'
    title VARCHAR(100) NOT NULL,
    schema JSONB NOT NULL DEFAULT '{}', -- The JSON Schema definition
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table: Entities
-- Stores the actual instances of data.
-- The 'data' column holds the properties defined in the Blueprint.
CREATE TABLE entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    blueprint_id VARCHAR(50) NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
    data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for Performance
-- 1. Index for fetching all entities of a specific blueprint
CREATE INDEX idx_entities_blueprint_id ON entities(blueprint_id);

-- 2. GIN Index on 'data' to allow efficient querying inside the JSONB blob
-- e.g., WHERE data @> '{"status": "active"}'
CREATE INDEX idx_entities_data ON entities USING GIN (data);