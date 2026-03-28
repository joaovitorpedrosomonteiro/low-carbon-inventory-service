-- Migration: Create tables for Inventory Service

-- GWP Standards table
CREATE TABLE gwp_standards (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    source VARCHAR(255) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Conversion Factors table
CREATE TABLE conversion_factors (
    id UUID PRIMARY KEY,
    gwp_standard_id UUID NOT NULL REFERENCES gwp_standards(id),
    gas_type VARCHAR(50) NOT NULL,
    value DECIMAL(20, 10) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Scopes table (seeded reference data)
CREATE TABLE scopes (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Categories table (seeded reference data)
CREATE TABLE categories (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    scope_id UUID NOT NULL REFERENCES scopes(id),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Emission Templates table
CREATE TABLE emission_templates (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    inventory_count INTEGER DEFAULT 0,
    is_frozen BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Links table (for evidences)
CREATE TABLE links (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    path TEXT NOT NULL,
    storage_type VARCHAR(20) NOT NULL CHECK (storage_type IN ('local', 'gcs')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Variables table
CREATE TABLE variables (
    id UUID PRIMARY KEY,
    emission_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    value DECIMAL(20, 10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Inventories table (aggregate root)
CREATE TABLE inventories (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    month INTEGER NOT NULL CHECK (month >= 1 AND month <= 12),
    year INTEGER NOT NULL CHECK (year >= 2000),
    state VARCHAR(50) NOT NULL DEFAULT 'to_report_emissions',
    template_id UUID REFERENCES emission_templates(id),
    company_branch_id UUID NOT NULL,
    gwp_standard_id UUID REFERENCES gwp_standards(id),
    review_message TEXT,
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (company_branch_id, month, year)
);

-- Emissions table
CREATE TABLE emissions (
    id UUID PRIMARY KEY,
    inventory_id UUID NOT NULL REFERENCES inventories(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    gas_type VARCHAR(50) NOT NULL,
    formula TEXT NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    reliability_job_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User Emails table (read model from Identity Service)
CREATE TABLE user_emails (
    user_id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    company_id UUID,
    branch_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Redis cache table (for caching)
CREATE TABLE redis_cache (
    key VARCHAR(255) PRIMARY KEY,
    data TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Seed Scopes
INSERT INTO scopes (id, name, description) VALUES 
    ('11111111-1111-1111-1111-111111111111', 'Escopo 1', 'Emissões diretas de fontes fixas e móveis'),
    ('22222222-2222-2222-2222-222222222222', 'Escopo 2', 'Emissões indiretas de energia elétrica e térmica'),
    ('33333333-3333-3333-3333-333333333333', 'Escopo 3', 'Demais emissões indiretas da cadeia de valor');

-- Seed Categories
INSERT INTO categories (id, name, scope_id, description) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Combustão Fixa', '11111111-1111-1111-1111-111111111111', 'Fontes fixas de combustão'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaab', 'Combustão Móvel', '11111111-1111-1111-1111-111111111111', 'Fontes móveis de combustão'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaac', 'Processos Industriais', '11111111-1111-1111-1111-111111111111', 'Processos industriais sem combustão'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaad', 'Fugitivas', '11111111-1111-1111-1111-111111111111', 'Emissões fugitivas'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaae', 'Energia Elétrica', '22222222-2222-2222-2222-222222222222', 'Consumo de energia elétrica'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaf', 'Energia Térmica', '22222222-2222-2222-2222-222222222222', 'Consumo de energia térmica'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaag', 'Compras', '33333333-3333-3333-3333-333333333333', 'Bens e serviços adquiridos'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaah', 'Transporte', '33333333-3333-3333-3333-333333333333', 'Transporte de cargas e passageiros'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaai', 'Resíduos', '33333333-3333-3333-3333-333333333333', 'Resíduos operacionais');

-- Seed GWP Standard (IPCC 2021)
INSERT INTO gwp_standards (id, name, source, is_default) VALUES
    ('aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'IPCC 2021', 'IPCC AR6', true);

-- Seed Conversion Factors
INSERT INTO conversion_factors (id, gwp_standard_id, gas_type, value, unit) VALUES
    ('bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'CO2', 1, 'tCO2/t'),
    ('bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'CH4', 27.9, 'tCO2e/tCH4'),
    ('bbbbbbb3-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'N2O', 273, 'tCO2e/tN2O'),
    ('bbbbbbb4-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'SF6', 25200, 'tCO2e/tSF6'),
    ('bbbbbbb5-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'NF3', 17300, 'tCO2e/tNF3');

-- Create indexes
CREATE INDEX idx_inventories_company_branch ON inventories(company_branch_id);
CREATE INDEX idx_inventories_state ON inventories(state);
CREATE INDEX idx_inventories_created_at ON inventories(created_at);
CREATE INDEX idx_emissions_inventory ON emissions(inventory_id);
CREATE INDEX idx_emissions_category ON emissions(category_id);
CREATE INDEX idx_links_emission ON links(id);
CREATE INDEX idx_variables_emission ON variables(emission_id);
CREATE INDEX idx_user_emails_branch ON user_emails(branch_id);
CREATE INDEX idx_user_emails_company ON user_emails(company_id);