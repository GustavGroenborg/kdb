# KDB — Befordringsfradrag Service

KDB is a self-contained web service to register and calculate Danish transport tax deductions (*befordringsfradrag*) for commuting between your residence, workplaces, and university.

Built with a **Go backend** embedding a **persistent SQLite database** (zero external database daemon required) and a modern **React + TypeScript + Tailwind CSS** frontend in `ui/`.

---

## Features

- **Smart Onboarding Flow**:
  - Greeted by a clean login and signup screen.
  - If no routes are configured yet, the app prompts you to register your home and first workplace/university route.
  - If routes exist, the default landing view prompts you directly to log a drive on that route.
- **Single Home & Multiple Workplaces/Universities**:
  - Enforces a single registered home residence.
  - Supports multiple workplaces and educational institutions.
- **Round-Trip Commute Calculations ("A drive counts for both back and forth")**:
  - Automatically calculates round-trip kilometers (2x one-way, with optional override).
  - Built-in Danish SKAT deduction rules:
    - 0 – 24 km/day: 0 DKK
    - 25 – 120 km/day: 2.23 DKK/km
    - > 120 km/day: 1.12 DKK/km
- **Three-Tab Top Menu**:
  1. **Register**: Fast 1-click drive logging, route picker, live deduction calculation, and today's drive list.
  2. **Overview**: Total DKK deductions, total km, monthly breakdown charts, workplace vs. university distribution, history log, and **CSV export for SKAT**.
  3. **Settings**: Manage home address, workplaces/universities, routes, default route, SKAT rates, and **Account Access Sharing**.
- **Account Access Sharing**:
  - Grant other registered users **Read Only** or **Read & Write** access to your account.
  - In-app Account Switcher lets authorized users log drives on shared accounts.
- **Nix Flakes**:
  - `service/flake.nix` for Go development.
  - `ui/flake.nix` for Node.js / TypeScript frontend development.
  - `flake.nix` in root for unified workspace development.
- **Oracle Cloud Ready**:
  - Single static binary deployment (17 MB total), Dockerfile, docker-compose, systemd service, and Caddy reverse-proxy config.

---

## Development Setup with Nix Flakes

### 1. Backend Shell
```bash
nix develop ./service
cd service
go test -v ./...
go run ./cmd/server
```

### 2. Frontend Shell
```bash
nix develop ./ui
cd ui
npm install
npm run dev
```

### 3. Unified Root Shell
```bash
nix develop
# Provides both go and node/npm in the same environment
```

---

## Building the Single Standalone Binary

To compile the entire application (embedded SQLite engine + backend + embedded React frontend) into one self-contained binary:

```bash
# 1. Build the UI
cd ui && npm run build && cd ..

# 2. Copy UI assets into Go embed folder
mkdir -p service/cmd/server/dist
cp -r ui/dist/* service/cmd/server/dist/

# 3. Build Go static binary
cd service
go build -o ../bin/kdb-service ./cmd/server
cd ..

# 4. Run!
./bin/kdb-service
# Visit http://localhost:8080
```

---

## Deployment to Oracle Cloud

Refer to the complete deployment guide in [`deploy/README.md`](deploy/README.md) for:
- Docker Compose one-click deployment
- Systemd service unit configuration
- Caddy reverse-proxy for free automatic HTTPS certificates
