# FILE: docker/Dockerfile.migrator
FROM alpine:latest

# Install PostgreSQL client and MySQL client
RUN apk add --no-cache \
    postgresql16-client \
    mysql-client \
    bash \
    curl

# Create app directory
WORKDIR /app

# Copy migration scripts and SQL files
COPY platform/database/migrations/ /app/migrations/
COPY docker/scripts/run-migrations.sh /app/
COPY docker/scripts/wait-for-services.sh /app/

# Make scripts executable
RUN chmod +x /app/run-migrations.sh /app/wait-for-services.sh

# Set default command
CMD ["/app/run-migrations.sh"]