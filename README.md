# Nginx Manager

A comprehensive nginx management platform with a modern web UI, configuration editor, SSL certificate management, and fail2ban security.

> **Repository**: [vm75/nginx-manager](https://github.com/vm75/nginx-manager)

## Features

- 📁 File browser with tree view
- ✏️ Code editor with nginx syntax highlighting (Monaco Editor)
- 🔄 Test and reload nginx configurations
- 📊 View access and error logs (auto-parsed from nginx.conf)
- ➕ Create, delete, rename files and folders
- 🎯 Drag and drop file operations
- 🔗 Create and manage symlinks (for sites-enabled)
- 🔐 SSL certificate management with Let's Encrypt
- 🌐 Wildcard certificate support (170+ DNS providers)
- 🔄 Automatic certificate renewal
- 🐳 Docker support with nginx and fail2ban included
- 🚀 Single binary deployment

## Tech Stack

- **Backend**: Go 1.21+ with embedded frontend
- **Frontend**: Svelte 4 + Vite 5
- **Editor**: Monaco Editor (VS Code editor)

---

## Quick Start

### Docker (Recommended)

```bash
# Using docker-compose
docker-compose up -d

# Access the web UI
open http://localhost:8080

# Nginx runs on
open http://localhost:80
```

### Standalone Binary

```bash
# Run with default nginx config directory (/etc/nginx)
./nginx-manager

# Specify custom config directory
./nginx-manager -config /path/to/nginx/config -port 8080
```

---

## Building from Source

```bash
# Build frontend
cd frontend
npm install
npm run build
cd ..

# Build Go binary
go build -o nginx-manager

# Or use build script
chmod +x build.sh
./build.sh
```

### Development

```bash
# Run backend
go run main.go -config ./test-config

# Run frontend dev server (separate terminal)
cd frontend
npm run dev
```

---

## Docker Setup

### Features

- **Nginx**: Web server and reverse proxy
- **Web UI**: Management interface on port 8080
- **Fail2ban**: Automatic IP banning for security
- **Lego**: Let's Encrypt ACME client (170+ DNS providers)
- **Cron**: Automatic certificate renewal (daily at 2 AM)
- **Supervisor**: Process management

### Quick Start

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Environment Variables

```yaml
environment:
  - TZ=UTC
  - NGINX_MANAGER_PORT=8080
  - CERT_RENEWAL_DAYS=5  # Days before expiry to renew
```

### Volumes

```yaml
volumes:
  - /DATA/docker/nginx-manager/config:/etc/nginx    # Nginx config
  - /DATA/docker/nginx-manager/logs:/var/log/nginx  # Nginx logs
  - /DATA/docker/nginx-manager/fail2ban:/var/log/fail2ban
  - /DATA/docker/nginx-manager/certs:/etc/nginx/ssl  # SSL certificates
```

### Docker Commands

```bash
# Enter container
docker exec -it nginx-manager sh

# Check services
supervisorctl status

# Check fail2ban
fail2ban-client status

# Manually trigger certificate renewal
docker exec nginx-manager /usr/local/bin/renew-certs.sh

# View renewal log
docker exec nginx-manager tail -f /var/log/cert-renewal.log
```

### Makefile Commands

A Makefile is provided for convenience:

```bash
make up          # Start containers
make down        # Stop containers
make logs        # View logs
make restart     # Restart containers
make shell       # Open shell in container
make status      # Show supervisor status
make nginx-test  # Test nginx config
make nginx-reload # Reload nginx
make fail2ban-status # Check fail2ban status
make build       # Build Docker image
make rebuild     # Clean rebuild
make help        # Show all available commands
```

---

## Certificate Management

### Install Lego (for standalone)

```bash
# Linux
wget https://github.com/go-acme/lego/releases/download/v4.15.0/lego_v4.15.0_linux_amd64.tar.gz
tar xf lego_v4.15.0_linux_amd64.tar.gz
sudo mv lego /usr/local/bin/
sudo chmod +x /usr/local/bin/lego

# macOS
brew install lego

# Verify
lego --version
```

### Challenge Types

#### HTTP-01 Challenge
- **Port**: 80 must be accessible
- **Use case**: Single domain certificates
- **Limitations**: No wildcard support

#### TLS-ALPN-01 Challenge
- **Port**: 443 must be accessible
- **Use case**: Alternative to HTTP-01
- **Limitations**: No wildcard support

#### DNS-01 Challenge
- **Requirements**: DNS provider API credentials
- **Use case**: Wildcard certificates, firewalled servers
- **Providers**: 170+ supported (Cloudflare, Route53, etc.)

### Quick Examples

#### HTTP-01 (Simple Domain)
1. Navigate to 🔐 Certificates tab
2. Enter domain: `example.com`
3. Enter email: `admin@example.com`
4. Select challenge: `HTTP-01`
5. Click "Obtain Certificate"

#### DNS-01 (Wildcard)
1. Navigate to 🔐 Certificates tab
2. Enter domain: `example.com`
3. Check: "Include wildcard certificate"
4. Select challenge: `DNS-01`
5. Enter provider code: `cloudflare`
6. Add environment variable: `CLOUDFLARE_DNS_API_TOKEN` = `your_token`
7. Click "Obtain Certificate"

This generates a certificate for **both** `example.com` and `*.example.com`

### Common DNS Providers

| Provider | Code | Environment Variables |
|----------|------|----------------------|
| Cloudflare | `cloudflare` | `CLOUDFLARE_DNS_API_TOKEN` |
| AWS Route53 | `route53` | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` |
| Google Cloud | `gcloud` | `GCE_PROJECT`, `GCE_SERVICE_ACCOUNT_FILE` |
| DigitalOcean | `digitalocean` | `DO_AUTH_TOKEN` |
| NameSilo | `namesilo` | `NAMESILO_API_KEY` |
| Namecheap | `namecheap` | `NAMECHEAP_API_USER`, `NAMECHEAP_API_KEY` |
| DuckDNS | `duckdns` | `DUCKDNS_TOKEN` |
| GoDaddy | `godaddy` | `GODADDY_API_KEY`, `GODADDY_API_SECRET` |

[View all 170+ providers →](https://go-acme.github.io/lego/dns/)

### Certificate Storage

Certificates are stored in:
```
<config-dir>/ssl/              # Certificate directory
├── example.com.crt            # Certificate (644)
├── example.com.key            # Private key (600)
└── .lego/                     # Lego working directory
    └── certificates/          # Original certificates
        ├── example.com.crt
        ├── example.com.key
        └── example.com.json   # Metadata
```

### Nginx Configuration

```nginx
server {
    listen 443 ssl http2;
    server_name example.com;

    ssl_certificate /etc/nginx/ssl/example.com.crt;
    ssl_certificate_key /etc/nginx/ssl/example.com.key;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Your site configuration...
}
```

### Automatic Renewal

#### Docker
Automatic renewal is built-in:
- Runs daily at 2 AM
- Renews certificates expiring in < 5 days (configurable)
- Logs to `/var/log/cert-renewal.log`

#### Standalone
Create a cron job:

```bash
# Create renewal script
cat > ~/renew-certs.sh << 'EOF'
#!/bin/bash
CONFIG_DIR="/etc/nginx"
LEGO_DIR="$CONFIG_DIR/.lego"
SSL_DIR="$CONFIG_DIR/ssl"

# Renew (adjust provider and credentials)
CLOUDFLARE_DNS_API_TOKEN="your_token" \
  lego --path "$LEGO_DIR" renew --days 5

# Copy certificates
cp "$LEGO_DIR/certificates/"*.crt "$SSL_DIR/"
cp "$LEGO_DIR/certificates/"*.key "$SSL_DIR/"
chmod 644 "$SSL_DIR/"*.crt
chmod 600 "$SSL_DIR/"*.key

# Reload nginx
nginx -s reload
EOF

chmod +x ~/renew-certs.sh

# Add to crontab (daily at 2 AM)
(crontab -l 2>/dev/null; echo "0 2 * * * ~/renew-certs.sh >> ~/cert-renewal.log 2>&1") | crontab -
```

---

## API Endpoints

### File Operations
- `GET /api/files?path=/` - List files in directory
- `GET /api/file/read?path=/file.conf` - Read file content
- `POST /api/file/write` - Write file content
- `POST /api/file/create` - Create file or directory
- `POST /api/file/delete` - Delete file or directory
- `POST /api/file/rename` - Rename file or directory
- `POST /api/file/move` - Move file or directory
- `POST /api/file/symlink` - Create symlink

### Nginx Operations
- `POST /api/nginx/test` - Test nginx configuration
- `POST /api/nginx/reload` - Reload nginx

### Logs
- `GET /api/logs/access?lines=100` - Get access log
- `GET /api/logs/error?lines=100` - Get error log

### Certificates
- `GET /api/certificates` - List SSL certificates
- `POST /api/certificates/obtain` - Obtain new certificate
- `POST /api/certificates/delete` - Delete certificate

### Additional Logs
- `GET /api/logs/cert-obtain` - Get certificate obtain log

---

## Security Considerations

### Docker
1. **Capabilities**: Requires `NET_ADMIN` and `NET_RAW` for fail2ban
2. **Port Restrictions**: Restrict port 8080 access with firewall
3. **SSL/TLS**: Always use HTTPS in production
4. **Authentication**: Add reverse proxy authentication for the web UI

### Fail2ban Jails
Pre-configured jails:
- `nginx-http-auth` - Failed HTTP auth
- `nginx-noscript` - Script kiddie attempts
- `nginx-badbots` - Bad bot detection
- `nginx-noproxy` - Proxy attempt blocking
- `nginx-limit-req` - Rate limit violations
- `nginx-botsearch` - Bot search patterns

### Certificate Security
- Keep private keys secure (`.key` files should be 600)
- Never commit certificates to version control
- Use strong SSL configuration
- Enable HSTS headers
- Implement OCSP stapling

---

## Troubleshooting

### Certificate Issues

**Viewing Certificate Obtain Logs**

The system now includes comprehensive logging for all certificate operations:
- View logs in the web UI: Logs → 🔐 Certificate Obtain tab
- Check log file: `/var/log/cert-obtain.log`
- Includes full lego output, timestamps, and error details
- DNS-01 challenges have a 10-minute timeout
- HTTP-01 challenges have a 2-minute timeout

**HTTP-01 fails with "connection refused"**
```bash
# Check port 80 is open
sudo netstat -tlnp | grep :80

# Check firewall
sudo ufw status
sudo ufw allow 80/tcp
```

**DNS-01 fails with "DNS validation timeout"**
- Verify API credentials are correct
- Check DNS provider API is enabled
- Wait for DNS propagation (5-10 minutes for most providers, up to 15 minutes for DuckDNS)
- Check provider API status
- View detailed logs in the Certificate Obtain log tab
- DNS timeout is set to 600 seconds (10 minutes) for propagation
- Overall operation timeout is 10 minutes for DNS-01 challenges
- Uses public DNS resolvers (8.8.8.8, 1.1.1.1) to avoid Docker DNS issues

**DuckDNS Specific Issues**
- DuckDNS can be slow to propagate DNS changes (10-15 minutes)
- Ensure your DUCKDNS_TOKEN is correct and valid
- Test your token: `curl "https://www.duckdns.org/update?domains=yourdomain&token=yourtoken&ip="`
- Check certificate obtain logs for detailed error messages
- The system uses public DNS resolvers (Google DNS, Cloudflare DNS) to avoid internal Docker DNS issues
- If you see "no such host" errors for DuckDNS nameservers, this is now automatically resolved

**Certificate not appearing**
```bash
# Check directory exists
ls -la /etc/nginx/ssl/

# Check permissions
sudo chmod 644 /etc/nginx/ssl/*.crt
sudo chmod 600 /etc/nginx/ssl/*.key

# View certificate obtain logs
docker exec nginx-manager tail -100 /var/log/cert-obtain.log

# Refresh browser
```

### Docker Issues

**Container won't start**
```bash
# Check logs
docker-compose logs nginx-manager

# Check nginx config
docker exec nginx-manager nginx -t
```

**Services not running**
```bash
# Check supervisor
docker exec nginx-manager supervisorctl status

# Restart service
docker exec nginx-manager supervisorctl restart nginx
```

---

## Requirements

- Go 1.21+
- Node.js 18+
- nginx installed (for test/reload functionality)
- lego (for SSL certificate management)
- Docker & Docker Compose (for containerized deployment)

---

## License

MIT

## Links

- [Lego Documentation](https://go-acme.github.io/lego/)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [Nginx Documentation](https://nginx.org/en/docs/)
- [Monaco Editor](https://microsoft.github.io/monaco-editor/)
- [Svelte](https://svelte.dev/)
