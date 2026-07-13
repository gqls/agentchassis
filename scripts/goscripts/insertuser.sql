-- SQL to insert a new user into the 'users' table
-- Replace 'YOUR_HASHED_PASSWORD_HERE' with the bcrypt hash generated in step 1.
-- Replace 'your_email@example.com' with a valid email.
-- 'demo_client' is the client_id we'll use for testing.

INSERT INTO users (id, email, password_hash, role, client_id, subscription_tier, is_active, email_verified, created_at, updated_at)
VALUES (
           UUID_TO_BIN('123e4567-e89b-12d3-a456-426614174000'), -- This is a sample UUID. You can generate a new one if preferred.
           -- The UUID_TO_BIN() function converts the string UUID to BINARY(16).
           'testuser@example.com',
           '$2a$10$ZKVVMvdtAFGawUIJy.p2Pe.N8ghhVjX2ZYre/nK0oNq7FvvfCZ/0i', -- PASTE THE HASHED PASSWORD HERE ToilAndTrouble123!
           'admin',                     -- Set as 'admin' for full access in tests
           'demo_client',               -- This client_id must match the one used for agent instances
           'enterprise',                -- Give them a high tier for testing quotas
           TRUE,
           TRUE,
           NOW(),
           NOW()
       );

-- Optional: Insert a basic subscription for this user
INSERT INTO subscriptions (id, user_id, tier, status, start_date, payment_method, created_at, updated_at)
VALUES (
           UUID_TO_BIN('00000000-0000-0000-0000-000000000005'), -- Unique UUID for subscription
           UUID_TO_BIN('123e4567-e89b-12d3-a456-426614174000'), -- User's UUID
           'enterprise',
           'active',
           NOW(),
           'test_method',
           NOW(),
           NOW()
       );

-- Optional: Insert a user profile
INSERT INTO user_profiles (user_id, first_name, last_name, company, created_at, updated_at)
VALUES (
           UUID_TO_BIN('123e4567-e89b-12d3-a456-426614174000'), -- User's UUID
           'Test',
           'User',
           'Test Company',
           NOW(),
           NOW()
       );

-- Optional: Grant admin permissions (if not already handled by role)
-- Assuming 'admin.users' and '*' permissions exist from your migrations
INSERT INTO user_permissions (user_id, permission_id, granted_at)
VALUES (
           UUID_TO_BIN('123e4567-e89b-12d3-a456-426614174000'),
           (SELECT id FROM permissions WHERE name = '*'), -- Grant super admin permission
           NOW()
       );