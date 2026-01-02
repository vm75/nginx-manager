#!/bin/bash
set -e

echo "🚀 Starting Nginx Editor Container..."

# Initialize fail2ban directories and files
echo "📦 Initializing fail2ban..."
mkdir -p /var/run/fail2ban /var/lib/fail2ban
touch /var/log/fail2ban/fail2ban.log
chmod 755 /var/run/fail2ban
chmod 755 /var/lib/fail2ban

# Create nginx log files if they don't exist
touch /var/log/nginx/access.log
touch /var/log/nginx/error.log

# Set proper permissions
chown -R nginx:nginx /var/log/nginx
chmod -R 755 /var/log/nginx

# Create certificate renewal log
touch /var/log/cert-renewal.log
chmod 644 /var/log/cert-renewal.log

# Create certificate obtain log
touch /var/log/cert-obtain.log
chmod 644 /var/log/cert-obtain.log

# Ensure cron is ready
touch /var/log/cron.log

# Test nginx configuration
echo "🔍 Testing nginx configuration..."
nginx -t || {
    echo "❌ Nginx configuration test failed"
    exit 1
}

echo "✅ All checks passed, starting services..."

# Execute the main command (supervisord)
exec "$@"
