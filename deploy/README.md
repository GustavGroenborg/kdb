# Oracle Cloud Deployment Guide for KDB

This guide walks you through deploying the **KDB Befordringsfradrag Service** to an **Oracle Cloud Infrastructure (OCI)** Compute instance (either Always-Free ARM Ampere A1 or AMD Compute).

---

## Architecture Summary
- **Single Self-Contained Binary**: The Go backend embeds the React/TypeScript frontend assets directly, with persistent SQLite database stored on disk (e.g., `/data/database/kdb.db`).
- **Resource Footprint**: Consumes only ~15–30 MB of RAM, making it optimal for free-tier cloud instances.

---

## Step 1: Provision Compute Instance on OCI
1. Log into your **Oracle Cloud Console**.
2. Go to **Compute** ➔ **Instances** ➔ **Create Instance**.
3. Choose:
   - **Image**: Ubuntu 22.04 / 24.04 or Oracle Linux 9.
   - **Shape**: `VM.Standard.A1.Flex` (Always-Free 4 OCPUs, 24GB RAM) or `VM.Standard.E2.1.Micro`.
4. Under **Networking**, assign a Public IPv4 address.
5. Add your SSH public key and click **Create**.

---

## Step 2: Open Ingress Ports in OCI Security List
1. In the OCI Console, navigate to your instance's **Virtual Cloud Network (VCN)** ➔ **Security Lists** ➔ **Default Security List**.
2. Add Ingress Rules:
   - **Port 80 (HTTP)**: Source `0.0.0.0/0`, IP Protocol `TCP`, Destination Port `80`.
   - **Port 443 (HTTPS)**: Source `0.0.0.0/0`, IP Protocol `TCP`, Destination Port `443`.
   - (Optional) **Port 8080**: If accessing directly without reverse proxy.
3. On the instance itself, open firewall rules:
   ```bash
   sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 80 -j ACCEPT
   sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 443 -j ACCEPT
   sudo netfilter-persistent save
   ```

---

## Step 3: Deployment Option A — Docker Compose (Recommended)

1. Install Docker & Docker Compose on your instance:
   ```bash
   curl -fsSL https://get.docker.com | sh
   sudo usermod -aG docker $USER
   ```
2. Clone your repository:
   ```bash
   git clone <your-kdb-repo-url> /opt/kdb
   cd /opt/kdb
   ```
3. Set your production secrets (JWT and optional Datafordeleren DAR API Token):
   ```bash
   export JWT_SECRET=$(openssl rand -hex 32)
   export DAR_API_TOKEN="your-datafordeler-api-token"
   ```
4. Start the service:
   ```bash
   docker compose up -d --build
   ```
5. Check status:
   ```bash
   docker compose ps
   docker compose logs -f
   ```

The service is now running on port `8080`, with database persisted in `./database/kdb.db`.

---

## Step 4: Deployment Option B — Native Systemd Service

If you prefer running without Docker:
1. Copy the precompiled `bin/kdb-service` binary to `/opt/kdb/bin/kdb-service`:
   ```bash
   sudo mkdir -p /opt/kdb/bin /opt/kdb/database
   sudo cp bin/kdb-service /opt/kdb/bin/
   sudo chmod +x /opt/kdb/bin/kdb-service
   ```
2. Provision the DAR API token using systemd credentials (isolated in-memory tmpfs):
   ```bash
   sudo mkdir -p /etc/kdb
   echo "your-datafordeler-api-token" | sudo tee /etc/kdb/dar_token > /dev/null
   sudo chmod 600 /etc/kdb/dar_token
   ```
3. Copy the systemd unit file:
   ```bash
   sudo cp deploy/kdb.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now kdb.service
   ```
4. Check status:
   ```bash
   sudo systemctl status kdb.service
   ```

---

## Step 5: HTTPS Reverse Proxy with Caddy

For free automatic HTTPS / SSL:
1. Install Caddy:
   ```bash
   sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
   curl -1sLF 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
   curl -1sLF 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
   sudo apt update
   sudo apt install caddy
   ```
2. Edit `/etc/caddy/Caddyfile`:
   ```caddy
   your-domain.com {
       reverse_proxy localhost:8080
       encode gzip zstd
   }
   ```
3. Restart Caddy:
   ```bash
   sudo systemctl reload caddy
   ```

Caddy automatically obtains and renews Let's Encrypt SSL certificates for your domain!
