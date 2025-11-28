-- Rollback initial schema

DROP TRIGGER IF EXISTS trigger_maintenance_logs_updated_at ON maintenance_logs;
DROP TRIGGER IF EXISTS trigger_bank_deposits_updated_at ON bank_deposits;
DROP TRIGGER IF EXISTS trigger_expenses_updated_at ON expenses;
DROP TRIGGER IF EXISTS trigger_weekly_reports_updated_at ON weekly_reports;
DROP TRIGGER IF EXISTS trigger_taxis_updated_at ON taxis;
DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
DROP TRIGGER IF EXISTS trigger_tenants_updated_at ON tenants;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS maintenance_logs;
DROP TABLE IF EXISTS bank_deposits;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS weekly_reports;
DROP TABLE IF EXISTS taxis;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;

