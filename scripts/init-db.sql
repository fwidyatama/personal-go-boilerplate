-- Database initialization script
-- This script is executed when the PostgreSQL container starts

-- Create database if it doesn't exist
-- Note: This is handled by the POSTGRES_DB environment variable in docker-compose.yml

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set timezone
SET timezone = 'UTC';

-- Create a read-only user for monitoring (optional)
-- CREATE USER readonly WITH PASSWORD 'readonly_password';
-- GRANT CONNECT ON DATABASE microservice_db TO readonly;
-- GRANT USAGE ON SCHEMA public TO readonly;
-- GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO readonly;

-- Create a read-write user for the application (optional)
-- CREATE USER app_user WITH PASSWORD 'app_password';
-- GRANT ALL PRIVILEGES ON DATABASE microservice_db TO app_user;
-- GRANT ALL PRIVILEGES ON SCHEMA public TO app_user;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app_user;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO app_user;

-- Log successful initialization
DO $$
BEGIN
    RAISE NOTICE 'Database initialization completed successfully';
END $$; 