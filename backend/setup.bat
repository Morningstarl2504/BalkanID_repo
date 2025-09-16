# ================ setup.bat (Windows) ================
@echo off
echo Setting up BalkanID File Vault Backend...

cd backend

REM Initialize go module (if not already done)
if not exist "go.mod" (
    go mod init github.com/Morningstarl2504/Balkanid_repo
)

REM Install dependencies
echo Installing Go dependencies...
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/golang-migrate/migrate/v4
go get github.com/joho/godotenv
go get github.com/golang-jwt/jwt/v4
go get golang.org/x/time
go get golang.org/x/crypto/bcrypt
go get github.com/google/uuid
go get github.com/stretchr/testify
go get github.com/lib/pq

REM Create upload directory
mkdir uploads

REM Create .env file if it doesn't exist
if not exist ".env" (
    echo Creating .env file...
    (
        echo DATABASE_URL=postgres://postgres:password@localhost:5432/filevault?sslmode=disable
        echo JWT_SECRET=your-super-secret-jwt-key-change-in-production
        echo PORT=8080
        echo UPLOAD_PATH=./uploads
        echo MAX_FILE_SIZE=52428800
        echo RATE_LIMIT=2
        echo STORAGE_QUOTA=10485760
    ) > .env
)

echo Backend setup complete!
echo Make sure PostgreSQL is running on port 5432
echo Run 'go run cmd/server/main.go' to start the server

# ================ run-dev.sh (Development) ================
#!/bin/bash

# Start PostgreSQL with Docker
echo "Starting PostgreSQL database..."
docker run --name filevault-postgres -d \
  -e POSTGRES_DB=filevault \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:14

# Wait for database to be ready
echo "Waiting for database to be ready..."
sleep 10

# Run the backend server
echo "Starting backend server..."
cd backend
go run cmd/server/main.go

# ================ Quick API Test Script (test-api.sh) ================
#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "Testing BalkanID File Vault API..."

# Test health endpoint
echo "1. Testing health endpoint..."
curl -X GET "$BASE_URL/../health"
echo -e "\n"

# Test user registration
echo "2. Testing user registration..."
curl -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
echo -e "\n"

# Test user login
echo "3. Testing user login..."
RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

# Extract token (requires jq)
if command -v jq &> /dev/null; then
    TOKEN=$(echo $RESPONSE | jq -r '.data.token')
    echo "Login successful, token: ${TOKEN:0:20}..."
    
    # Test authenticated endpoint
    echo "4. Testing authenticated endpoint..."
    curl -X GET "$BASE_URL/profile" \
      -H "Authorization: Bearer $TOKEN"
    echo -e "\n"
else
    echo "Install jq to extract token automatically"
    echo $RESPONSE
fi

echo "API test complete!"

# ================ Database Init Script (database/init.sql) ================
-- Create database if not exists (optional, since Docker creates it)
-- CREATE DATABASE filevault;

-- Create extension for UUID generation (if needed)
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- The actual tables will be created by GORM migrations
-- This file can contain any initial data or custom configurations

-- Example: Create an admin user (password: admin123)
-- Note: This will be executed after GORM migrations
-- INSERT INTO users (username, email, password_hash, is_admin, created_at, updated_at) 
-- VALUES (
--   'admin', 
--   'admin@example.com', 
--   '$2a$10$N9qo8uLOickgx2ZMRZoMye1J8kWb8wZR5lp3Iyf/H7Uf4d8hqzVwy', -- bcrypt hash of 'admin123'
--   true, 
--   CURRENT_TIMESTAMP, 
--   CURRENT_TIMESTAMP
-- );