-- FILE: platform/database/migrations/004_auth_schema.sql
-- Auth database schema

CREATE TABLE IF NOT EXISTS users (
    -- MySQL doesn't have a UUID type. BINARY(16) is the recommended way to store UUIDs efficiently.
    -- The UUID() function generates a standard UUID string, and UUID_TO_BIN() converts it to a compact binary format.
    -- This DEFAULT expression requires MySQL 8.0.13 or newer.
    id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),

    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    client_id VARCHAR(100) NOT NULL,
    subscription_tier VARCHAR(50) DEFAULT 'free',
    is_active BOOLEAN DEFAULT TRUE,

    -- TIMESTAMPTZ is a PostgreSQL type. The MySQL equivalent is TIMESTAMP,
    -- which stores the timestamp in UTC and converts it to the connection's time zone on retrieval.
    -- CURRENT_TIMESTAMP is the standard function to get the current time.
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Using ON UPDATE CURRENT_TIMESTAMP is a common pattern to automatically update
    -- this field whenever a row's data is changed.
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );


CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_client ON users(client_id);

CREATE TABLE IF NOT EXISTS auth_tokens (
    -- Use BINARY(16) for UUIDs and the UUID_TO_BIN(UUID()) function for the default value.
                                           id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),

    -- The user_id must also be BINARY(16) to match the 'id' column in the 'users' table.
    user_id BINARY(16) NOT NULL,

    token_hash VARCHAR(255) NOT NULL,

    -- TIMESTAMPTZ is replaced with TIMESTAMP.
    expires_at TIMESTAMP NOT NULL,

    -- TIMESTAMPTZ is replaced with TIMESTAMP, and NOW() is replaced with the standard CURRENT_TIMESTAMP.
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- It's good practice to explicitly define the foreign key constraint.
    -- ON DELETE CASCADE will automatically delete a user's tokens if the user is deleted.
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );

CREATE INDEX idx_tokens_user ON auth_tokens(user_id);
CREATE INDEX idx_tokens_expires ON auth_tokens(expires_at);

CREATE INDEX idx_users_email_active ON users(email, is_active);