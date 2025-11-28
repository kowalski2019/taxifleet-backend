-- Rollback: Change permission column back to role

-- Add role column back
ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'driver';

-- Migrate permission data back to role (approximate mapping)
UPDATE users SET role = CASE
    WHEN permission = 4294967295 THEN 'admin'     -- All permissions
    WHEN permission = 1048575 THEN 'owner'        -- All business permissions
    WHEN permission = 3 THEN 'driver'             -- View and Add reports
    WHEN permission = 112 THEN 'mechanic'        -- Taxi management
    WHEN permission & 3 = 3 AND permission = 3 THEN 'manager' -- View and Add reports only
    ELSE 'driver'
END;

-- Make role NOT NULL
ALTER TABLE users ALTER COLUMN role SET NOT NULL;

-- Drop permission column
ALTER TABLE users DROP COLUMN permission;

-- Drop permission index
DROP INDEX IF EXISTS idx_users_permission;

-- Recreate role index
CREATE INDEX idx_users_role ON users(role);

