-- Migration: Add admin user
-- This migration creates a system tenant (if it doesn't exist) and an admin user
-- Admin credentials: admin@taxifleet.ci / admin123
-- Permission: -1 (represents 0xFFFFFFFF in signed INTEGER - all permissions - admin)
-- Note: PostgreSQL INTEGER is signed, so 0xFFFFFFFF (4294967295) is out of range.
-- Using -1 which has all bits set in two's complement, works correctly with bitwise operations.

-- Create system tenant if it doesn't exist
INSERT INTO tenants (name, subdomain, settings, created_at, updated_at)
SELECT 
    'System',
    'system',
    '{}'::jsonb,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
WHERE NOT EXISTS (
    SELECT 1 FROM tenants WHERE subdomain = 'system'
);

-- Get the system tenant ID (or use the first tenant if system tenant doesn't exist)
DO $$
DECLARE
    system_tenant_id INTEGER;
    admin_permission INTEGER := -1; -- -1 represents 0xFFFFFFFF (all permissions) in signed integer (two's complement)
    admin_password_hash TEXT := '$2a$10$dqcqcApxCPK1GxHBDM2OqukHz4KuPP2QXCyGlOfEVGdaUMErehWsq'; -- bcrypt hash for "admin123"
BEGIN
    -- Get system tenant ID, or use the first tenant if system tenant doesn't exist
    SELECT id INTO system_tenant_id 
    FROM tenants 
    WHERE subdomain = 'system' 
    LIMIT 1;
    
    -- If no system tenant found, use the first tenant
    IF system_tenant_id IS NULL THEN
        SELECT id INTO system_tenant_id 
        FROM tenants 
        ORDER BY id ASC 
        LIMIT 1;
    END IF;
    
    -- Only proceed if we have a tenant
    IF system_tenant_id IS NOT NULL THEN
        -- Create admin user if it doesn't exist
        INSERT INTO users (
            tenant_id,
            email,
            password_hash,
            permission,
            first_name,
            last_name,
            phone,
            active,
            created_at,
            updated_at
        )
        SELECT 
            system_tenant_id,
            'admin@taxifleet.ci',
            admin_password_hash,
            admin_permission,
            'Admin',
            'Admin',
            '+1234567890',
            true,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        WHERE NOT EXISTS (
            SELECT 1 FROM users WHERE email = 'admin@taxifleet.ci'
        );
    END IF;
END $$;

