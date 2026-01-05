FROM node:18-alpine AS frontend-builder

WORKDIR /build
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.21-alpine AS backend-builder

WORKDIR /build
COPY go.mod ./
COPY main.go ./
COPY --from=frontend-builder /build/dist ./frontend/dist
RUN go build -o nginx-manager main.go

FROM alpine:latest

# Install nginx, fail2ban, acme.sh, cron, docker-cli and other dependencies
RUN apk add --no-cache \
    nginx \
    fail2ban \
    iptables \
    bash \
    tzdata \
    ca-certificates \
    curl \
    supervisor \
    dcron \
    openssl \
    wget \
    logrotate \
    docker-cli \
    && curl https://get.acme.sh | sh \
    && ln -s /root/.acme.sh/acme.sh /usr/local/bin/acme.sh

# Install Incus CLI binary
RUN curl -fsSL https://github.com/lxc/incus/releases/download/v6.20.0/bin.linux.incus.x86_64 -o /usr/local/bin/incus \
    && chmod +x /usr/local/bin/incus

# Create necessary directories
RUN mkdir -p /var/log/nginx \
    /var/log/fail2ban \
    /var/log/supervisor \
    /var/run/fail2ban \
    /var/lib/fail2ban \
    /etc/nginx/sites-available \
    /etc/nginx/sites-enabled \
    /etc/nginx/conf.d \
    /etc/fail2ban/filter.d \
    /etc/fail2ban/action.d \
    /etc/fail2ban/jail.d

# Create logrotate configuration for nginx
RUN mkdir -p /etc/logrotate.d
RUN cat <<EOF > /etc/logrotate.d/nginx
/var/log/nginx/*.log /var/log/fail2ban/*.log /var/log/supervisor/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 0644 root root
    postrotate
        /usr/sbin/nginx -s reload
    endscript
}
EOF

# Copy nginx-manager binary
COPY --from=backend-builder /build/nginx-manager /usr/local/bin/nginx-manager
RUN chmod +x /usr/local/bin/nginx-manager

# Copy fail2ban configurations
COPY docker/fail2ban/jail.local /etc/fail2ban/jail.local
COPY docker/fail2ban/filter.d/ /etc/fail2ban/filter.d/

# Copy nginx default configuration
COPY docker/nginx/nginx.conf /etc/nginx/nginx.conf
COPY docker/nginx/default.conf /etc/nginx/sites-available/default.conf
RUN ln -s /etc/nginx/sites-available/default.conf /etc/nginx/sites-enabled/default.conf

# Copy supervisor configuration
COPY docker/supervisor/supervisord.conf /etc/supervisord.conf

# Copy certificate renewal script
COPY docker/renew-certs.sh /usr/local/bin/renew-certs.sh
RUN chmod +x /usr/local/bin/renew-certs.sh

# Setup cron job for certificate renewal (daily at 2 AM) and logrotate (daily at midnight)
RUN echo -e '0 0 * * * /usr/sbin/logrotate /etc/logrotate.d/nginx\n0 2 * * * /usr/local/bin/renew-certs.sh >> /var/log/cert-renewal.log 2>&1' > /etc/crontabs/root \
    && touch /var/log/cert-renewal.log

# Copy entrypoint script
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose ports
EXPOSE 80 443 8080

# Set working directory
WORKDIR /etc/nginx

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/ || exit 1

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
