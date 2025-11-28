-- TaxiFleet Database Schema Migration
-- Initial schema creation

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable pgcrypto for password hashing
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Tenants table (multi-tenant support)
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(100) UNIQUE NOT NULL,
    logo TEXT,
    settings JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    permission INTEGER DEFAULT 3 NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT unique_email_per_tenant UNIQUE (tenant_id, email)
);

-- Sessions table
CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Taxis table
CREATE TABLE taxis (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    license_plate VARCHAR(50) NOT NULL,
    model VARCHAR(255),
    year INTEGER,
    color VARCHAR(50),
    vin VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    assigned_driver_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Weekly Reports table
CREATE TABLE weekly_reports (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    taxi_id INTEGER NOT NULL REFERENCES taxis(id) ON DELETE CASCADE,
    driver_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    week_start_date DATE NOT NULL,
    earnings DECIMAL(10, 2) NOT NULL DEFAULT 0,
    total_expenses DECIMAL(10, 2) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    notes TEXT,
    submitted_at TIMESTAMP WITH TIME ZONE,
    approved_at TIMESTAMP WITH TIME ZONE,
    approved_by_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Expenses table
CREATE TABLE expenses (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    report_id INTEGER REFERENCES weekly_reports(id) ON DELETE SET NULL,
    taxi_id INTEGER REFERENCES taxis(id) ON DELETE SET NULL,
    category VARCHAR(50) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    reason TEXT,
    receipt_url TEXT,
    date DATE NOT NULL,
    created_by_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Bank Deposits table
CREATE TABLE bank_deposits (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL,
    deposit_date DATE NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    bank_account VARCHAR(255),
    proof_url TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Maintenance Logs table
CREATE TABLE maintenance_logs (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    taxi_id INTEGER NOT NULL REFERENCES taxis(id) ON DELETE CASCADE,
    description TEXT,
    cost DECIMAL(10, 2),
    date DATE NOT NULL,
    mechanic_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for performance

-- Tenants indexes
CREATE INDEX idx_tenants_subdomain ON tenants(subdomain);
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at);

-- Users indexes
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Sessions indexes
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_deleted_at ON sessions(deleted_at);

-- Taxis indexes
CREATE INDEX idx_taxis_tenant_id ON taxis(tenant_id);
CREATE INDEX idx_taxis_assigned_driver_id ON taxis(assigned_driver_id);
CREATE INDEX idx_taxis_status ON taxis(status);
CREATE INDEX idx_taxis_deleted_at ON taxis(deleted_at);

-- Weekly Reports indexes
CREATE INDEX idx_weekly_reports_tenant_id ON weekly_reports(tenant_id);
CREATE INDEX idx_weekly_reports_taxi_id ON weekly_reports(taxi_id);
CREATE INDEX idx_weekly_reports_driver_id ON weekly_reports(driver_id);
CREATE INDEX idx_weekly_reports_status ON weekly_reports(status);
CREATE INDEX idx_weekly_reports_week_start_date ON weekly_reports(week_start_date);
CREATE INDEX idx_weekly_reports_deleted_at ON weekly_reports(deleted_at);

-- Expenses indexes
CREATE INDEX idx_expenses_tenant_id ON expenses(tenant_id);
CREATE INDEX idx_expenses_report_id ON expenses(report_id);
CREATE INDEX idx_expenses_taxi_id ON expenses(taxi_id);
CREATE INDEX idx_expenses_category ON expenses(category);
CREATE INDEX idx_expenses_date ON expenses(date);
CREATE INDEX idx_expenses_deleted_at ON expenses(deleted_at);

-- Bank Deposits indexes
CREATE INDEX idx_bank_deposits_tenant_id ON bank_deposits(tenant_id);
CREATE INDEX idx_bank_deposits_deposit_date ON bank_deposits(deposit_date);
CREATE INDEX idx_bank_deposits_deleted_at ON bank_deposits(deleted_at);

-- Maintenance Logs indexes
CREATE INDEX idx_maintenance_logs_tenant_id ON maintenance_logs(tenant_id);
CREATE INDEX idx_maintenance_logs_taxi_id ON maintenance_logs(taxi_id);
CREATE INDEX idx_maintenance_logs_date ON maintenance_logs(date);
CREATE INDEX idx_maintenance_logs_deleted_at ON maintenance_logs(deleted_at);

-- Create function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at
CREATE TRIGGER trigger_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_taxis_updated_at
    BEFORE UPDATE ON taxis
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_weekly_reports_updated_at
    BEFORE UPDATE ON weekly_reports
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_expenses_updated_at
    BEFORE UPDATE ON expenses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_bank_deposits_updated_at
    BEFORE UPDATE ON bank_deposits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_maintenance_logs_updated_at
    BEFORE UPDATE ON maintenance_logs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

