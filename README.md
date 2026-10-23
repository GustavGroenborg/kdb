# KDB — Befordringsfradrag Service

KDB is a self-contained web service to register and calculate Danish transport tax deductions (*befordringsfradrag*) for commuting between your residence, workplaces, and university.

Built with a **Go backend** embedding a **persistent SQLite database** (zero external database daemon required) and a modern **React + TypeScript + Tailwind CSS** frontend in `ui/`.

---

## Features

- **Smart Onboarding Flow**:
  - Clean login and signup screens with persistent session authentication.
  - If no routes exist, guides you to register your home and first workplace/university destination.
  - If routes exist, defaults directly to commute registration for fast logging.
- **Single Home & Multiple Workplaces/Universities**:
  - Enforces a single registered primary residence.
  - Supports multiple workplaces and educational institutions.
  - **Route Visibility Toggle**: Hide legacy or inactive routes from the registration dropdown without deleting historical drive records or losing deduction statistics.
- **Round-Trip Commute Calculations ("A drive counts for both back and forth")**:
  - Automatically calculates round-trip kilometers (2x one-way, with optional override).
  - Built-in Danish SKAT deduction rules across tax years:
    - **0 – 24 km/day**: 0 DKK
    - **25 – 120 km/day**: 2.23 DKK/km (2024–2025) / **3.17 DKK/km** (2026)
    - **> 120 km/day**: 1.12 DKK/km (2024–2025) / **1.59 DKK/km** (2026)
    - **Yderkommuner & Småøer (Ligningslovens § 9 C, stk. 3)**:
      - 2024–2025: 2.23 DKK/km for all km > 24
      - 2026: **3.51 DKK/km** for all km > 24 regardless of distance
      - Extensible for 2027+ annual rate updates via `tax_rates` database table.
- **Automatic Municipality & Yderkommune Resolution via DAR**:
  - Resolves home addresses directly against the Datafordeleren Danish Address Register (DAR) GraphQL API.
  - Automatically identifies residences in 25 authorized yderkommuner / eligible small island municipalities.
  - Displays municipality and elevated rate badges in Settings and live deduction preview in Register.
  - Displays a warning banner if an address cannot be resolved in DAR.
- **Interactive Overview & Route Breakdown**:
  - Total deductions (DKK), kilometers logged, and monthly breakdown charts.
  - Segmented monthly progress bars showing each route's commute contribution in high-contrast distinct colors.
  - Interactive inspection: hover over sub-segments on desktop or tap on mobile to inspect trip counts, kilometers, and deductions per route.
  - **CSV export for SKAT** tax reporting.
- **Account Access Sharing & Collaborator Delegation**:
  - Grant other registered users **Read Only** or **Read & Write** access to your account.
  - **Integrated Top Menu User Card**: Seamlessly switch between your personal account and shared accounts right from the sticky top navigation on both desktop and mobile.
- **Mobile First & iOS Optimized**:
  - Clean responsive slide-in drawer for app navigation.
  - Fully optimized for iOS Safari and WebKit viewports (no horizontal overflow).
- **Hermetic Nix Flakes**:
  - Declarative build targets for native Linux, cross-compiled ARM64 (OCI), and static frontend assets.
  - Zero-configuration development shells (`nix develop`).

---

## Development Setup with Nix

Start the unified development shell providing both Go and Node.js:

```bash
nix develop
```

### 1. Run the Frontend (with Hot-Module Reloading)
```bash
cd ui
npm install
npm run dev
# Vite dev server runs on http://localhost:5173
```

### 2. Run the Backend Service
```bash
cd service
go run ./cmd/server
# Backend API runs on http://localhost:8080
```

---

## Building with Nix (`nix build`)

KDB uses hermetic Nix Flake derivations to package the application with zero ambient host dependencies.

### 1. Build Native Standalone Server
Compiles the React frontend, embeds the static assets into Go, and produces a single self-contained binary:
```bash
nix build
# Or explicitly:
nix build .#kdb
```
The output binary will be symlinked to **`./result/bin/kdb-service`** (~12 MB).

### 2. Cross-Compile ARM64 Binary for Oracle Cloud (OCI)
Cross-compiles directly for ARM64 without needing QEMU or cross-gcc downloads:
```bash
nix build .#kdb-arm64
```
The output binary will be symlinked to **`./result/bin/kdb-service-linux-arm64`**.

### 3. Build Only Frontend Static Assets
```bash
nix build .#ui
```
Outputs the compiled production HTML/JS/CSS assets to **`./result/`**.

---

## Alternative Build Script (Non-Nix)

If building outside of Nix, you can use the bundled build script:

```bash
./build.sh
```

This will run `npm run build` in `ui/`, copy built assets into `service/cmd/server/dist`, and compile the binary to `bin/kdb-service`.

---

## Deployment to Oracle Cloud Infrastructure (OCI)

Refer to the complete deployment guide in [`deploy/README.md`](deploy/README.md) for:
- One-click Docker Compose deployment
- Systemd service unit setup
- Automatic HTTPS certificates with Caddy reverse proxy

### Quick 1-Second Zero-Downtime Server Update
To update an already running OCI instance:

```bash
# 1. Build ARM64 binary locally
nix build .#kdb-arm64

# 2. Upload to temporary file on the server (zero downtime during transfer)
scp ./result/bin/kdb-service-linux-arm64 ubuntu@kdb.hjortbonnerup.dk:/home/ubuntu/kdb-service.new

# 3. SSH into instance and swap binary (downtime < 1s)
ssh ubuntu@kdb.hjortbonnerup.dk
chmod +x /home/ubuntu/kdb-service.new
sudo systemctl stop kdb
sqlite3 /home/ubuntu/database/kdb.db "VACUUM INTO '/home/ubuntu/database/kdb_pre_update.db';"
mv /home/ubuntu/kdb-service.new /home/ubuntu/kdb-service
sudo systemctl start kdb
```
*(Database migrations apply automatically and idempotently on startup with zero data loss).*
