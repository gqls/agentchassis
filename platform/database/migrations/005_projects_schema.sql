// FILE: platform/database/migrations/005_projects_schema.sql
-- Projects table for auth database
CREATE TABLE IF NOT EXISTS projects (
                                        id VARCHAR(36) PRIMARY KEY,
    client_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id VARCHAR(36) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (owner_id) REFERENCES users(id),
    INDEX idx_projects_client (client_id),
    INDEX idx_projects_owner (owner_id)
    );

-- Subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
                                             id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL UNIQUE,
    tier VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ,
    trial_ends_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    payment_method VARCHAR(100),
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_subscriptions_status (status)
    );

-- Subscription tiers table
CREATE TABLE IF NOT EXISTS subscription_tiers (
                                                  id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    price_monthly DECIMAL(10,2) NOT NULL,
    price_yearly DECIMAL(10,2) NOT NULL,
    max_personas INT NOT NULL,
    max_projects INT NOT NULL,
    max_content_items INT NOT NULL,
    features JSON,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Insert default tiers
INSERT INTO subscription_tiers (id, name, display_name, description, price_monthly, price_yearly, max_personas, max_projects, max_content_items, features) VALUES
                                                                                                                                                               ('00000000-0000-0000-0000-000000000001', 'free', 'Free', 'Basic features for getting started', 0.00, 0.00, 1, 3, 10, '["Basic personas", "Limited content generation"]'),
                                                                                                                                                               ('00000000-0000-0000-0000-000000000002', 'basic', 'Basic', 'For individual users', 9.99, 99.99, 5, 10, 100, '["All persona types", "Priority support", "Advanced templates"]'),
                                                                                                                                                               ('00000000-0000-0000-0000-000000000003', 'premium', 'Premium', 'For power users', 29.99, 299.99, 20, 50, 1000, '["All basic features", "Custom personas", "API access", "Analytics"]'),
                                                                                                                                                               ('00000000-0000-0000-0000-000000000004', 'enterprise', 'Enterprise', 'For organizations', 99.99, 999.99, -1, -1, -1, '["All premium features", "Unlimited usage", "Dedicated support", "Custom integrations"]');

-- User profiles table
CREATE TABLE IF NOT EXISTS user_profiles (
                                             user_id VARCHAR(36) PRIMARY KEY,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    company VARCHAR(255),
    phone VARCHAR(50),
    avatar_url VARCHAR(500),
    preferences JSON,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id)
    );

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
                                           id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- User permissions junction table
CREATE TABLE IF NOT EXISTS user_permissions (
                                                user_id VARCHAR(36) NOT NULL,
    permission_id VARCHAR(36) NOT NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, permission_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (permission_id) REFERENCES permissions(id)
    );

-- Insert default permissions
INSERT INTO permissions (id, name, description) VALUES
                                                    ('00000000-0000-0000-0000-000000000001', 'personas.create', 'Create new personas'),
                                                    ('00000000-0000-0000-0000-000000000002', 'personas.delete', 'Delete personas'),
                                                    ('00000000-0000-0000-0000-000000000003', 'projects.manage', 'Manage all projects'),
                                                    ('00000000-0000-0000-0000-000000000004', 'admin.users', 'Manage users'),
                                                    ('00000000-0000-0000-0000-000000000005', 'admin.subscriptions', 'Manage subscriptions'),
                                                    ('00000000-0000-0000-0000-000000000006', '*', 'Super admin - all permissions');