-- Rollback seed data
-- Note: This will delete the seed data, but keep the structure

DELETE FROM taxis WHERE tenant_id IN (SELECT id FROM tenants WHERE subdomain = 'gnakpa-transport');
DELETE FROM users WHERE tenant_id IN (SELECT id FROM tenants WHERE subdomain = 'gnakpa-transport');
DELETE FROM tenants WHERE subdomain = 'gnakpa-transport';

