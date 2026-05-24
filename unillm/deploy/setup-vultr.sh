#!/bin/bash
set -e

echo "=========================================="
echo "  UniLLM Vultr Deployment"
echo "=========================================="

# 1. System updates + Docker
echo "[1/7] Installing Docker + dependencies..."
apt-get update -qq
apt-get install -y -qq docker.io docker-compose-v2 git curl ufw > /dev/null 2>&1
systemctl enable docker
systemctl start docker

# 2. Install Go 1.23 (latest stable that works)
echo "[2/7] Installing Go..."
if ! command -v go &> /dev/null; then
    curl -sL https://go.dev/dl/go1.23.8.linux-amd64.tar.gz | tar -C /usr/local -xz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc
    export PATH=$PATH:/usr/local/go/bin
fi
go version

# 3. Firewall
echo "[3/7] Configuring firewall..."
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 8080/tcp
ufw --force enable

# 4. Clone and build
echo "[4/7] Cloning UniLLM..."
mkdir -p /opt/unillm
cd /opt/unillm

if [ -d ".git" ]; then
    git pull
else
    # Copy from local (will be done via scp)
    echo "  Waiting for code upload..."
fi

# 5. Start PostgreSQL + Redis via Docker
echo "[5/7] Starting PostgreSQL + Redis..."
cat > /opt/unillm/docker-compose.prod.yml << 'COMPOSE'
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: unillm
      POSTGRES_USER: unillm
      POSTGRES_PASSWORD: UniLLM_Pr0d_2026!
    ports:
      - "127.0.0.1:5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U unillm"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "127.0.0.1:6379:6379"
    command: redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru --appendonly yes
    volumes:
      - redisdata:/data
    restart: unless-stopped

volumes:
  pgdata:
  redisdata:
COMPOSE

docker compose -f docker-compose.prod.yml up -d

# Wait for PostgreSQL
echo "  Waiting for PostgreSQL..."
sleep 5
until docker compose -f docker-compose.prod.yml exec -T postgres pg_isready -U unillm > /dev/null 2>&1; do
    sleep 2
done
echo "  PostgreSQL ready."

# 6. Environment file
echo "[6/7] Creating environment config..."
JWT_SECRET=$(openssl rand -hex 32)
ENCRYPTION_KEY=$(openssl rand -hex 32)

cat > /opt/unillm/.env << ENVFILE
DATABASE_URL=postgres://unillm:UniLLM_Pr0d_2026!@127.0.0.1:5432/unillm?sslmode=disable
REDIS_URL=redis://127.0.0.1:6379/0
JWT_SECRET=${JWT_SECRET}
ENCRYPTION_KEY=${ENCRYPTION_KEY}
ENVIRONMENT=prod
PORT=8080
CORS_ORIGINS=*
MAX_BODY_BYTES=10485760
BCRYPT_COST=12
ENVFILE

chmod 600 /opt/unillm/.env

# 7. systemd service
echo "[7/7] Creating systemd service..."
cat > /etc/systemd/system/unillm.service << 'SERVICE'
[Unit]
Description=UniLLM API Server
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
WorkingDirectory=/opt/unillm
EnvironmentFile=/opt/unillm/.env
ExecStart=/opt/unillm/unillm
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
SERVICE

systemctl daemon-reload

echo ""
echo "=========================================="
echo "  Infrastructure ready!"
echo "=========================================="
echo "  PostgreSQL: 127.0.0.1:5432"
echo "  Redis: 127.0.0.1:6379"
echo "  Next: upload code, build, seed, start"
echo "=========================================="
