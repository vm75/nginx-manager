#!/bin/bash

# Certificate Renewal Script for Docker
# Runs daily at 2 AM to check and renew expiring certificates

CONFIG_DIR="/etc/nginx"
ACME_DIR="/root/.acme.sh"
SSL_DIR="$CONFIG_DIR/ssl"
LOG_FILE="/var/log/cert-renewal.log"

# Configurable renewal threshold (default: 5 days)
RENEWAL_DAYS="${CERT_RENEWAL_DAYS:-5}"

echo "========================================" >> "$LOG_FILE"
echo "$(date): Starting certificate renewal check (renewal threshold: $RENEWAL_DAYS days)" >> "$LOG_FILE"

# Check if acme.sh directory exists
if [ ! -d "$ACME_DIR" ]; then
    echo "$(date): No acme.sh directory found. No certificates to renew." >> "$LOG_FILE"
    exit 0
fi

# Check if there are any accounts/certificates
if [ ! -d "$ACME_DIR/ca" ]; then
    echo "$(date): No acme.sh accounts found. No certificates to renew." >> "$LOG_FILE"
    exit 0
fi

# Create SSL directory if it doesn't exist
mkdir -p "$SSL_DIR"

# Track renewal status
RENEWAL_NEEDED=false
RENEWAL_SUCCESS=false

# Find all certificate directories
for cert_dir in "$ACME_DIR"/*; do
    if [ -d "$cert_dir" ]; then
        # Check for both ECC and RSA directories
        if [ -f "$cert_dir/fullchain.cer" ]; then
            # Get the domain from directory name (remove _ecc suffix if present)
            domain=$(basename "$cert_dir" | sed 's/_ecc$//')

            # Check certificate expiry
            cert_file="$cert_dir/fullchain.cer"
            if [ -f "$cert_file" ]; then
                expiry_date=$(openssl x509 -in "$cert_file" -noout -enddate 2>/dev/null | cut -d= -f2)

                if [ -n "$expiry_date" ]; then
                    expiry_epoch=$(date -d "$expiry_date" +%s 2>/dev/null || date -j -f "%b %d %T %Y %Z" "$expiry_date" +%s 2>/dev/null)
                    current_epoch=$(date +%s)
                    days_left=$(( ($expiry_epoch - $current_epoch) / 86400 ))

                    echo "$(date): Certificate for $domain expires in $days_left days" >> "$LOG_FILE"

                    # Renew if less than RENEWAL_DAYS remaining
                    if [ $days_left -lt $RENEWAL_DAYS ]; then
                        echo "$(date): Certificate for $domain needs renewal (expires in $days_left days, threshold: $RENEWAL_DAYS days)" >> "$LOG_FILE"
                        RENEWAL_NEEDED=true
                    fi
                fi
            fi
        fi
    fi
done

# If renewal is needed, try to renew all certificates
if [ "$RENEWAL_NEEDED" = true ]; then
    echo "$(date): Attempting certificate renewal..." >> "$LOG_FILE"

    # Try to renew certificates using acme.sh
    echo "$(date): Running acme.sh renew-all command..." >> "$LOG_FILE"

    # Attempt renewal - acme.sh will only renew if necessary
    /usr/local/bin/acme.sh --renew-all >> "$LOG_FILE" 2>&1

    if [ $? -eq 0 ]; then
        echo "$(date): Certificate renewal successful" >> "$LOG_FILE"
        RENEWAL_SUCCESS=true

        # Copy renewed certificates to SSL directory
        echo "$(date): Copying renewed certificates to SSL directory" >> "$LOG_FILE"
        for cert_dir in "$ACME_DIR"/*; do
            if [ -d "$cert_dir" ] && [ -f "$cert_dir/fullchain.cer" ]; then
                # Get the domain from directory name (remove _ecc suffix if present)
                domain=$(basename "$cert_dir" | sed 's/_ecc$//')
                cp "$cert_dir/fullchain.cer" "$SSL_DIR/$domain.crt" 2>/dev/null
                cp "$cert_dir/$domain.key" "$SSL_DIR/$domain.key" 2>/dev/null
            fi
        done

        # Set proper permissions
        chmod 644 "$SSL_DIR/"*.crt 2>/dev/null
        chmod 600 "$SSL_DIR/"*.key 2>/dev/null

        # Reload nginx to use new certificates
        echo "$(date): Reloading nginx..." >> "$LOG_FILE"
        nginx -s reload >> "$LOG_FILE" 2>&1

        if [ $? -eq 0 ]; then
            echo "$(date): Nginx reloaded successfully" >> "$LOG_FILE"
        else
            echo "$(date): WARNING: Failed to reload nginx" >> "$LOG_FILE"
        fi
    else
        echo "$(date): Certificate renewal failed or no renewal was needed" >> "$LOG_FILE"
    fi
else
    echo "$(date): No certificates need renewal at this time" >> "$LOG_FILE"
fi

echo "$(date): Certificate renewal check completed" >> "$LOG_FILE"
echo "========================================" >> "$LOG_FILE"

exit 0
