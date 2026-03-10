CREATE DATABASE auth_db;
CREATE USER auth_user WITH PASSWORD 'generate-a-password';
GRANT ALL ON DATABASE auth_db TO auth_user;

ALTER USER auth_user WITH PASSWORD 'Authy!123!Password';