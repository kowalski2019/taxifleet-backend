-- Add unique constraint to phone number
-- Phone numbers must be unique globally (country code + number combination)

-- First, remove any duplicate phone numbers (keep the first one)
WITH duplicates AS (
  SELECT id, phone, ROW_NUMBER() OVER (PARTITION BY phone ORDER BY id) as rn
  FROM users
  WHERE phone IS NOT NULL AND phone != ''
)
UPDATE users
SET phone = NULL
WHERE id IN (
  SELECT id FROM duplicates WHERE rn > 1
);

-- Add unique constraint
CREATE UNIQUE INDEX idx_users_phone_unique ON users(phone) WHERE phone IS NOT NULL AND phone != '';

