# Webhook Auto-Deploy Setup

Como o GitHub Actions está bloqueado pelo firewall, vamos usar **Webhooks** para deploy automático.

## Como Funciona

```
Git Push → GitHub → Webhook → Servidor → Git Pull → Docker Rebuild → Deploy ✅
```

## 📋 Instalação no Servidor

### 1. Compile o webhook server

```bash
# No servidor
cd /var/www/ifinu-radio/ifinu-radio-back
go build -o webhook-server webhook-server.go
chmod +x webhook-server webhook-deploy.sh
```

### 2. Gere um secret para o webhook

```bash
# Gere um secret seguro
openssl rand -base64 32
```

**Copie esse valor** - você vai usar no GitHub e no serviço.

### 3. Configure o systemd service

```bash
# Edite o service file
sudo nano /etc/systemd/system/webhook-deploy.service
```

Cole este conteúdo (substitua `YOUR_SECRET_HERE` pelo secret gerado):

```ini
[Unit]
Description=IFINU Radio Backend Webhook Deploy Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/ifinu-radio/ifinu-radio-back
Environment="WEBHOOK_SECRET=YOUR_SECRET_HERE"
ExecStart=/var/www/ifinu-radio/ifinu-radio-back/webhook-server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 4. Inicie o serviço

```bash
# Recarregue systemd
sudo systemctl daemon-reload

# Inicie o webhook server
sudo systemctl start webhook-deploy

# Ative para iniciar no boot
sudo systemctl enable webhook-deploy

# Verifique status
sudo systemctl status webhook-deploy
```

### 5. Configure Nginx para proxy do webhook

```bash
sudo nano /etc/nginx/sites-available/webhook.ifinu.io
```

```nginx
server {
    listen 80;
    server_name webhook.ifinu.io;

    location /webhook {
        proxy_pass http://localhost:9000/webhook;
        proxy_http_version 1.1;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Hub-Signature-256 $http_x_hub_signature_256;
    }
}
```

```bash
# Ative o site
sudo ln -s /etc/nginx/sites-available/webhook.ifinu.io /etc/nginx/sites-enabled/

# Teste e recarregue
sudo nginx -t
sudo systemctl reload nginx

# Configure SSL
sudo certbot --nginx -d webhook.ifinu.io
```

## 🔧 Configurar Webhook no GitHub

### 1. Acesse as configurações do repositório

https://github.com/ifinu/IFINU-RADIO-BACK/settings/hooks

### 2. Clique em "Add webhook"

- **Payload URL**: `https://webhook.ifinu.io/webhook`
- **Content type**: `application/json`
- **Secret**: Cole o secret que você gerou (mesmo do step 2)
- **Which events**: Selecione "Just the push event"
- **Active**: ✅ Marcado

### 3. Clique em "Add webhook"

## ✅ Testar

### 1. Faça um commit e push

```bash
cd /Users/mikael/ifinu-radio/ifinu-radio-back-new
echo "# Test" >> README.md
git add README.md
git commit -m "test: webhook deploy"
git push origin main
```

### 2. Verifique os logs no servidor

```bash
# Logs do webhook server
sudo journalctl -u webhook-deploy -f

# Logs do deploy
tail -f /var/log/ifinu-backend-deploy.log
```

### 3. Verifique no GitHub

Vá em: https://github.com/ifinu/IFINU-RADIO-BACK/settings/hooks

Clique no webhook → "Recent Deliveries"

Deve mostrar ✅ verde com status 200.

## 🔍 Troubleshooting

### Webhook server não inicia

```bash
# Verifique logs
sudo journalctl -u webhook-deploy -n 50

# Verifique se WEBHOOK_SECRET está configurado
sudo systemctl show webhook-deploy | grep WEBHOOK_SECRET
```

### Deploy não acontece

```bash
# Verifique permissões
ls -la /var/www/ifinu-radio/ifinu-radio-back/webhook-deploy.sh

# Verifique se é executável
chmod +x /var/www/ifinu-radio/ifinu-radio-back/webhook-deploy.sh

# Teste manualmente
sudo /var/www/ifinu-radio/ifinu-radio-back/webhook-deploy.sh
```

### GitHub mostra erro no webhook

- Verifique se webhook.ifinu.io está acessível: `curl https://webhook.ifinu.io/webhook`
- Verifique SSL: Deve ter certificado válido
- Verifique secret: Deve ser exatamente igual no GitHub e no serviço

## 🎯 Resultado

Agora, **toda vez que você fizer `git push`**:

1. ✅ GitHub envia webhook para seu servidor
2. ✅ Webhook server executa `webhook-deploy.sh`
3. ✅ Script faz `git pull` e rebuilda containers
4. ✅ Backend atualizado automaticamente!

**Deploy 100% automático sem GitHub Actions!** 🚀
