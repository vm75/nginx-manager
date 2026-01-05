# Reverse Proxy Guide

This guide shows how to set up nginx as a reverse proxy for server-manager, either as a subdomain or subfolder.

## Subdomain Setup

Add this to your nginx configuration:

```nginx
server {
    listen 80;
    server_name manager.example.com;

    location / {
        proxy_pass http://your-server-ip:8080;
        include snippets/proxy.conf;
    }
}
```

Replace `manager.example.com` with your subdomain and `your-server-ip` with the IP where server-manager is running.

## Subfolder Setup

Add this to your existing server block:

```nginx
server {
    listen 80;
    server_name example.com;

    # Your existing site config here
    location / {
        # ... your main site
    }

    # Server manager at /manager/
    location /manager {
        return 301 $scheme://$host/manager/;
    }

    location ^~ /manager/ {
        proxy_pass http://your-server-ip:8080/;
        include snippets/proxy.conf;
        rewrite ^/manager/(.*) /$1 break;
    }
}
```

Access server-manager at `http://example.com/manager/`.

## SSL/HTTPS

For HTTPS, add SSL certificates to the server blocks above and redirect HTTP to HTTPS:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # ... rest of your config
}
```