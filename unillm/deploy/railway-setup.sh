#!/bin/bash
set -e

echo "=========================================="
echo "  UniLLM Railway Deployment Setup"
echo "=========================================="
echo ""

# Check prerequisites
command -v railway >/dev/null 2>&1 || { echo "ERROR: Railway CLI not installed. Run: npm i -g @railway/cli"; exit 1; }
command -v gh >/dev/null 2>&1 || { echo "WARNING: GitHub CLI not installed. You'll need to connect the repo manually."; }

# Step 1: Login
echo "[1/6] Logging into Railway..."
railway login 2>/dev/null || true

# Step 2: Create project
echo ""
echo "[2/6] Creating Railway project..."
railway init

# Step 3: Add PostgreSQL
echo ""
echo "[3/6] Adding PostgreSQL database..."
echo "  → Go to Railway dashboard and add a PostgreSQL plugin"
echo "  → Or run: railway add --plugin postgresql"
railway add --plugin postgresql 2>/dev/null || echo "  (Add PostgreSQL manually via dashboard)"

# Step 4: Add Redis
echo ""
echo "[4/6] Adding Redis..."
echo "  → Go to Railway dashboard and add a Redis plugin"
echo "  → Or run: railway add --plugin redis"
railway add --plugin redis 2>/dev/null || echo "  (Add Redis manually via dashboard)"

# Step 5: Generate and set secrets
echo ""
echo "[5/6] Setting environment variables..."
JWT_SECRET=$(openssl rand -hex 32)
ENCRYPTION_KEY=$(openssl rand -hex 32)

echo "  Generated JWT_SECRET and ENCRYPTION_KEY"

railway vars set JWT_SECRET="$JWT_SECRET" 2>/dev/null || echo "  Set JWT_SECRET manually in dashboard"
railway vars set ENCRYPTION_KEY="$ENCRYPTION_KEY" 2>/dev/null || echo "  Set ENCRYPTION_KEY manually in dashboard"
railway vars set ENVIRONMENT=prod 2>/dev/null || echo "  Set ENVIRONMENT=prod manually"
railway vars set PORT=8080 2>/dev/null || echo "  Set PORT=8080 manually"

# Step 6: Deploy
echo ""
echo "[6/6] Deploying..."
railway up

echo ""
echo "=========================================="
echo "  Deployment complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Add a custom domain:  railway domain"
echo "  2. Check logs:           railway logs"
echo "  3. Open dashboard:       railway open"
echo ""
echo "Environment variables to set for frontend service:"
echo "  BACKEND_URL = \${{backend.RAILWAY_PRIVATE_DOMAIN}}:8080"
echo ""
