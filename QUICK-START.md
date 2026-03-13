# Quick Start - IFINU Radio API

## 🚀 Deploy em 5 Passos

### 1. Clone e Configure

```bash
# No servidor
cd /var/www/ifinu-radio
git clone https://github.com/SEU_USUARIO/ifinu-radio-back.git
cd ifinu-radio-back

# Copie e configure .env
cp .env.example .env
nano .env
```

**Configure no .env**:
```bash
# Gere uma API key forte:
openssl rand -base64 32

# Cole no .env:
ADMIN_API_KEY=sua_key_gerada_aqui
ALLOWED_ORIGINS=https://ifinu.io,https://www.ifinu.io
DB_PASSWORD=senha_forte_aqui
```

### 2. Inicie os Serviços

```bash
docker compose up -d
```

### 3. Verifique o Status

```bash
# Logs em tempo real
docker compose logs -f api

# Verificar health
curl https://api.ifinu.io/health
```

Resposta esperada:
```json
{
  "status": "ok",
  "database": "ok",
  "total_radios": 0,
  "version": "1.0.0"
}
```

### 4. Force o Sync Inicial

```bash
# Substitua YOUR_API_KEY pela key do .env
curl -X POST https://api.ifinu.io/api/v1/admin/sync \
  -H "X-API-Key: YOUR_API_KEY"
```

Aguarde ~30-60 segundos. Resposta esperada:
```json
{
  "sucesso": true,
  "mensagem": "Sync completed successfully",
  "total_radios": 1500
}
```

### 5. Teste a API

```bash
# Listar rádios
curl https://api.ifinu.io/api/v1/radios?limit=5

# Buscar por rock
curl https://api.ifinu.io/api/v1/radios/search?q=rock

# Health check novamente
curl https://api.ifinu.io/health
```

Agora `total_radios` deve estar > 0!

## 🔧 Troubleshooting

### Problema: `{"dados":[],"sucesso":true,"total":0}`

**Causa**: Banco vazio, sync ainda não executou.

**Solução**:
```bash
# 1. Verifique logs
docker compose logs api | grep -i sync

# 2. Force sync manual
curl -X POST https://api.ifinu.io/api/v1/admin/sync \
  -H "X-API-Key: SUA_KEY"

# 3. Verifique health
curl https://api.ifinu.io/health
```

### Problema: `"database": "error"`

**Causa**: PostgreSQL não conectado.

**Solução**:
```bash
# Verifique status do postgres
docker compose ps postgres

# Reinicie se necessário
docker compose restart postgres api
```

### Problema: `401 Unauthorized` no /admin/sync

**Causa**: API key não configurada ou incorreta.

**Solução**:
```bash
# 1. Verifique se ADMIN_API_KEY está no .env
grep ADMIN_API_KEY .env

# 2. Se não estiver, adicione:
echo "ADMIN_API_KEY=$(openssl rand -base64 32)" >> .env

# 3. Reinicie
docker compose restart api
```

### Problema: CORS error no frontend

**Causa**: ALLOWED_ORIGINS incorreto.

**Solução**:
```bash
# No .env:
ALLOWED_ORIGINS=https://ifinu.io,https://www.ifinu.io

# Reinicie
docker compose restart api
```

## 📊 Comandos Úteis

```bash
# Ver logs
docker compose logs -f api

# Reiniciar API
docker compose restart api

# Parar tudo
docker compose down

# Ver status
docker compose ps

# Entrar no container
docker compose exec api sh

# Backup do banco
docker compose exec postgres pg_dump -U postgres radio_db > backup.sql
```

## ✅ Checklist de Produção

- [ ] SSL configurado (`certbot --nginx -d api.ifinu.io`)
- [ ] ADMIN_API_KEY forte configurada
- [ ] ALLOWED_ORIGINS apenas com domínios necessários
- [ ] Senha do PostgreSQL forte
- [ ] Backup automático configurado
- [ ] Monitoring (uptime, logs)
- [ ] Firewall configurado (apenas 80, 443, 22)

## 🆘 Precisa de Ajuda?

1. Verifique logs: `docker compose logs -f api`
2. Verifique health: `curl https://api.ifinu.io/health`
3. Leia [SECURITY.md](SECURITY.md)
4. Abra uma issue no GitHub
