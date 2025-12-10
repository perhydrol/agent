#!/bin/bash

# Define colors
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}Starting Development Environment (MySQL + Redis)...${NC}"

# Stop existing containers
echo "Stopping old containers..."
podman rm -f insurai-mysql insurai-redis 2>/dev/null

# Start Redis
echo "Starting Redis..."
podman run -d --name insurai-redis \
    -p 6379:6379 \
    docker.io/library/redis:latest

# Start MySQL
echo "Starting MySQL..."
podman run -d --name insurai-mysql \
    -p 3306:3306 \
    -e MYSQL_ROOT_PASSWORD=123456 \
    -e MYSQL_DATABASE=insurai \
    docker.io/library/mysql:latest

# Wait for MySQL to be ready
echo "Waiting for MySQL to initialize..."
for i in {1..30}; do
    if podman exec insurai-mysql mysqladmin ping -h localhost -u root -p123456 --silent; then
        echo -e "${GREEN}MySQL is up!${NC}"
        break
    fi
    sleep 2
    echo -n "."
done

echo -e "\n${GREEN}Environment Ready!${NC}"
echo "MySQL: 127.0.0.1:3306 (root/123456, db: insurai)"
echo "Redis: 127.0.0.1:6379"
