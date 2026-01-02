#!/bin/bash

# Certificate Renewal Script for Docker
# Runs daily at 2 AM to check and renew expiring certificates

CONFIG_DIR="/etc/nginx"
LEGO_DIR="$CONFIG_DIR/.lego"
SSL_DIR="$CONFIG_DIR/ssl"
LOG_FILE="/var/log/cert-renewal.log"

# Configurable renewal threshold (default: 5 days)
RENEWAL_DAYS="${CERT_RENEWAL_DAYS:-5}"

echo "========================================" >> "$LOG_FILE"
echo "$(date): Starting certificate renewal check (renewal threshold: $RENEWAL_DAYS days)" >> "$LOG_FILE"

# Check if lego directory exists
if [ ! -d "$LEGO_DIR" ]; then
    echo "$(date): No lego directory found. No certificates to renew." >> "$LOG_FILE"
    exit 0
fi

# Check if there are any accounts/certificates
if [ ! -d "$LEGO_DIR/accounts" ]; then
    echo "$(date): No lego accounts found. No certificates to renew." >> "$LOG_FILE"
    exit 0
fi

# Create SSL directory if it doesn't exist
mkdir -p "$SSL_DIR"

# Track renewal status
RENEWAL_NEEDED=false
RENEWAL_SUCCESS=false

# Find all certificate directories
for cert_dir in "$LEGO_DIR/certificates"/*; do
    if [ -d "$cert_dir" ] || [ -f "$LEGO_DIR/certificates/"*.crt ]; then

        # Get the domain from certificate file
        for cert_file in "$LEGO_DIR/certificates/"*.crt; do
            if [ -f "$cert_file" ]; then
                domain=$(basename "$cert_file" .crt)

                # Check certificate expiry
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

                        # Try to renew
                        # Note: This requires the original certificate request parameters
                        # We'll attempt renewal for all certificates found in lego directory
                    fi
                fi
            fi
        done

        break
    fi
done

# If renewal is needed, try to renew all certificates
if [ "$RENEWAL_NEEDED" = true ]; then
    echo "$(date): Attempting certificate renewal..." >> "$LOG_FILE"

    # Get list of domains from certificate files
    domains=""
    for cert_file in "$LEGO_DIR/certificates/"*.crt; do
        if [ -f "$cert_file" ]; then
            domain=$(basename "$cert_file" .crt)
            # Get the actual domain from certificate subject
            actual_domain=$(openssl x509 -in "$cert_file" -noout -subject 2>/dev/null | sed -n 's/.*CN = \([^,]*\).*/\1/p' | head -1)
            if [ -n "$actual_domain" ]; then
                echo "$(date): Found certificate for domain: $actual_domain" >> "$LOG_FILE"
            fi
        fi
    done

    # Try to renew certificates (this will only work if certificates were obtained via lego)
    # The renewal will use stored account information
    if [ -d "$LEGO_DIR/accounts" ]; then
        echo "$(date): Running lego renew command..." >> "$LOG_FILE"

        # Attempt renewal - lego will only renew if necessary
        /usr/local/bin/lego --path "$LEGO_DIR" renew --days $RENEWAL_DAYS >> "$LOG_FILE" 2>&1

        if [ $? -eq 0 ]; then
            echo "$(date): Certificate renewal successful" >> "$LOG_FILE"
            RENEWAL_SUCCESS=true

            # Copy renewed certificates to SSL directory
            echo "$(date): Copying renewed certificates to SSL directory" >> "$LOG_FILE"
            cp "$LEGO_DIR/certificates/"*.crt "$SSL_DIR/" 2>/dev/null
            cp "$LEGO_DIR/certificates/"*.key "$SSL_DIR/" 2>/dev/null

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
        echo "$(date): No lego accounts found. Certificates may need manual renewal." >> "$LOG_FILE"
    fi
else
    echo "$(date): No certificates need renewal at this time" >> "$LOG_FILE"
fi

echo "$(date): Certificate renewal check completed" >> "$LOG_FILE"
echo "========================================" >> "$LOG_FILE"

exit 0
