#!/bin/bash
# Webhook deploy script
# Executado pelo webhook quando há push no GitHub

set -e

LOG_FILE="/var/log/ifinu-backend-deploy.log"

log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "🚀 Deploy iniciado por webhook"

cd /var/www/ifinu-radio/ifinu-radio-back || exit 1

# Pull latest changes
log "📥 Pulling latest changes from GitHub..."
git pull origin main

# Configure .env if not exists
if [ ! -f .env ]; then
  log "⚙️  Configurando .env inicial..."
  cp .env.example .env

  # Generate secure API key
  API_KEY=$(openssl rand -base64 32)
  sed -i "s/your-secret-api-key-here-change-in-production/$API_KEY/" .env
  sed -i "s|ALLOWED_ORIGINS=.*|ALLOWED_ORIGINS=https://ifinu.io,https://www.ifinu.io|" .env

  log "✅ .env configurado"
fi

# Rebuild and restart containers
log "🐳 Stopping containers..."
docker compose stop api postgres || true

log "🔨 Building containers (no cache)..."
docker compose build --no-cache api

log "▶️  Starting containers..."
docker compose up -d api postgres

# Wait for services to start
sleep 5

# Show logs
log "📋 Container logs:"
docker compose logs --tail=20 api | tee -a "$LOG_FILE"

log "✅ Deploy concluído com sucesso!"
