FROM alpine:latest

# Install PostgreSQL client and MySQL client
RUN apk add --no-cache \
    postgresql16-client \
    mysql-client \
    bash \
    curl \
    jq

# Create app directory
WORKDIR /app

# Copy seeding scripts and data files
COPY docker/scripts/seed-data.sh /app/
COPY docker/data/ /app/data/
COPY docker/scripts/wait-for-services.sh /app/

# Make scripts executable
RUN chmod +x /app/seed-data.sh /app/wait-for-services.sh

# Set default command
CMD ["/app/seed-data.sh"]