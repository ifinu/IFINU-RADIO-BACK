# Guia de Segurança

## Configurações Obrigatórias para Produção

### 1. API Key para Rotas Admin

**CRÍTICO**: Configure uma API key forte para proteger rotas administrativas:

```bash
# Gere uma key segura:
openssl rand -base64 32

# Configure no .env ou docker-compose:
ADMIN_API_KEY=sua_key_gerada_aqui
```

**Uso**:
```bash
# Sync manual de rádios
curl -X POST https://api.ifinu.io/api/v1/admin/sync \
  -H "X-API-Key: sua_key_aqui"
```

### 2. CORS (Cross-Origin Resource Sharing)

Configure origens permitidas:

```bash
# Para produção - apenas domínios específicos
ALLOWED_ORIGINS=https://ifinu.io,https://www.ifinu.io

# Para desenvolvimento local
ALLOWED_ORIGINS=http://localhost:3000
```

**NUNCA** use `*` em produção.

### 3. Banco de Dados

- ✅ Use senhas fortes (mínimo 32 caracteres)
- ✅ Nunca exponha porta PostgreSQL publicamente
- ✅ Use SSL/TLS para conexões do banco
- ✅ Backup automático configurado

### 4. SSL/TLS

Configure certificado para `api.ifinu.io`:

```bash
sudo certbot --nginx -d api.ifinu.io
```

### 5. Variáveis de Ambiente Sensíveis

**NUNCA** commite no Git:
- ❌ Senhas do banco
- ❌ API keys
- ❌ Tokens de autenticação

Use `.env` e adicione ao `.gitignore`.

## Vulnerabilidades Mitigadas

### ✅ SQL Injection
- GORM com prepared statements
- Validação de entrada

### ✅ Command Injection
- Sem execução de comandos shell com entrada do usuário

### ✅ Denial of Service (DoS)
- Health check com depends_on no Docker
- Connection pooling configurado
- Timeouts apropriados

### ✅ Authentication Bypass
- API key obrigatória para rotas admin
- Constant-time comparison

### ✅ Information Disclosure
- Logs não expõem dados sensíveis
- Errors genéricos para usuários

## Checklist de Deploy

Antes de fazer deploy em produção:

- [ ] `ADMIN_API_KEY` configurada (forte, 32+ caracteres)
- [ ] `ALLOWED_ORIGINS` apenas com domínios necessários
- [ ] Senha do PostgreSQL forte
- [ ] SSL/TLS configurado (HTTPS)
- [ ] Porta PostgreSQL não exposta publicamente
- [ ] `.env` não commitado no Git
- [ ] Logs de acesso configurados
- [ ] Backup automático do banco
- [ ] Monitoring configurado

## Reporte de Vulnerabilidades

Se encontrar uma vulnerabilidade de segurança, **NÃO** abra uma issue pública.

Contato: security@ifinu.io

## Atualizações de Segurança

- Mantenha Go atualizado (1.21+)
- Atualize dependências regularmente: `go get -u ./...`
- Monitore CVEs das dependências
