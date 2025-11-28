-- Rollback: Remove admin user
-- This migration removes the admin user created in the up migration

DELETE FROM users WHERE email = 'admin@taxifleet.ci';

-- Optionally remove system tenant if it was created by this migration
-- Uncomment the following if you want to remove the system tenant on rollback
-- DELETE FROM tenants WHERE subdomain = 'system' AND name = 'System';

