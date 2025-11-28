-- Migration: Change role column to permission (integer)
-- This migration converts the role-based system to a permission-based system

-- Add new permission column
ALTER TABLE users ADD COLUMN permission INTEGER DEFAULT 3; -- Default to driver permissions (view + add reports)

-- Migrate existing role data to permissions
UPDATE users SET permission = CASE
    WHEN role = 'admin' THEN 4294967295 -- All permissions (0xFFFFFFFF)
    WHEN role = 'owner' THEN 1048575    -- All business permissions (0xFFFFF)
    WHEN role = 'manager' THEN 28695        -- View and Add reports only
    WHEN role = 'mechanic' THEN 17     -- Taxi management (0x70)
    WHEN role = 'driver' THEN 3          -- View and Add reports
    ELSE 3                               -- Default to driver
END;

-- Make permission NOT NULL after migration
ALTER TABLE users ALTER COLUMN permission SET NOT NULL;

-- Drop the old role column
ALTER TABLE users DROP COLUMN role;

-- Drop the old role index
DROP INDEX IF EXISTS idx_users_role;

-- Create index on permission column
CREATE INDEX idx_users_permission ON users(permission);

