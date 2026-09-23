# GuangJiApps

Modern enterprise web application for the GuangJi organization, providing congregant management, educational classes, administrative workflows, and donation tracking.

This project is a complete modernization and migration from a legacy **.NET Framework + SSRS + Microsoft SQL Server** monolith (`GuangJiAppsNew`) to a decoupled, high-performance architecture powered by **Gin-Gonic (Go)** and **Vue 3 + TypeScript + Element Plus + PWA**.

---

## Table of Contents

- [GuangJiApps](#guangjiapps)
  - [Table of Contents](#table-of-contents)
  - [Project Overview](#project-overview)
  - [Technology Stack](#technology-stack)
  - [Repository Structure](#repository-structure)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Environment Configuration](#environment-configuration)
    - [Starting Services with Docker Compose](#starting-services-with-docker-compose)
    - [Running the Backend Locally (Go)](#running-the-backend-locally-go)
    - [Running the Frontend Locally (Vue 3)](#running-the-frontend-locally-vue-3)
  - [Testing \& Quality Assurance](#testing--quality-assurance)
  - [Production Windows IIS Deployment](#production-windows-iis-deployment)
  - [Coding Style \& Development Guidelines](#coding-style--development-guidelines)
    - [Backend (Golang / Gin)](#backend-golang--gin)
    - [Frontend (Vue 3 / TypeScript)](#frontend-vue-3--typescript)
    - [Documentation \& Workflow](#documentation--workflow)
  - [Postman API Collection \& Testing](#postman-api-collection--testing)
  - [License](#license)

---

## Project Overview

GuangJiApps manages organizational data, liturgical events, classes, and contributions:

- **Auth & Security**: JWT-based authentication with RSA/PEM secret signing, Redis token caching, client IP-aware rate limiting (`ulule/limiter`), and brute-force protection (`Fail2Ban` middleware).
- **Admin Management**: User management (`T_Login_Mst`), permission groups (`T_Login_Group`), group menu mappings (`T_Login_Menu`), and warehouse sub-locations (`T_WH_SUBWH_MST`).
- **Master Data**:
  - **Umat (Congregants)** (`T_BUS_UMAT`): Comprehensive profiles, profile photo upload & cropping, dynamic JWT QR code generation & verification, OCR identity card scanning, and SSRS paginated reporting.
  - **Topics & Activities**: Categorized syllabus topics (`T_BUS_TOPIC`) and events (`T_BUS_EVENT`).
  - **Lookups & Classifications**: Dedicated lookup services for positions, education, gender, blood types, and organization divisions.
  - **Sxy Donaturs & Fundraisers**: Donor records and donation solicitors.
- **Transactions**:
  - **Kelas (Classes)**: Class schedules, participant attendance, previous cohort pre-loading, volunteers/instructors, logistics, and class expenses.
  - **Donasi Sxy**: Financial contributions and SSRS Excel reporting.
- **Progressive Web App (PWA)**:
  - Custom Workbox service worker caching strategies (`DynamicNetworkCacheStrategy` and `DynamicLowNetworkCacheStrategy`).
  - Request deduplication to avoid redundant in-flight network requests on rapid pagination.
  - Persistent offline indicators and automatic reload notifications on new deployments.

---

## Technology Stack

| Layer | Technologies |
|---|---|
| **Backend API** | [Go 1.25+](https://go.dev/), [Gin-Gonic](https://gin-gonic.com/), [GORM](https://gorm.io/) (SQL Server driver), [Redis](https://redis.io/) (v9 client), [go-playground/validator/v10](https://github.com/go-playground/validator), [Excelize](https://qax-os.github.io/excelize/) |
| **Frontend Web** | [Vue 3](https://vuejs.org/) (Composition API, `<script setup>`), [TypeScript 5](https://www.typescriptlang.org/), [Element Plus](https://element-plus.org/), [Pinia](https://pinia.vuejs.org/), [Vue Router](https://router.vuejs.org/), [Vite 6](https://vite.dev/), [Day.js](https://day.js.org/) |
| **PWA / Offline** | [Vite-Plugin-PWA](https://vite-pwa-org.netlify.app/), [Workbox](https://developer.chrome.com/docs/workbox), Custom Service Worker Caching Strategies |
| **Database & Cache** | [Microsoft SQL Server 2022](https://www.microsoft.com/en-us/sql-server/sql-server-2022) (`mssql_server`), [Redis 7 Alpine](https://hub.docker.com/_/redis) (`redis_app`) |
| **Tooling & Auditing** | [Docker & Compose](https://docs.docker.com/), [Grafana k6](https://k6.io/) (Performance testing), [Semgrep](https://semgrep.dev/) (Security audit), [Postman](https://www.postman.com/) |
| **Target Deployment** | Windows Server **IIS** with [HttpPlatformHandler](https://learn.microsoft.com/en-us/iis/extensions/httpplatformhandler/httpplatformhandler-configuration-reference) |

---

## Repository Structure

```text
GuangJiApps/
├── docker-compose.yml              # Multi-container orchestration (MSSQL, Redis, Gin, Vue, k6, Semgrep)
├── fix-docker.sh                   # Helper script to clean stale Docker Desktop sockets and restart services
├── DOCUMENTATION.md                # Living documentation of all completed tasks and system enhancements
├── README.md                       # Main project documentation (this file)
│
├── gin/                            # Golang RESTful API Backend
│   ├── cmd/server/main.go          # Application entrypoint & HTTP server bootstrap
│   ├── internal/
│   │   ├── api/                    # Gin router, route groups, and middlewares (auth, ratelimit, fail2ban)
│   │   ├── config/                 # Environment configuration loader (godotenv)
│   │   ├── database/               # Database connection pool (MSSQL & Redis)
│   │   ├── domain/                 # Entity models, legacy schemas, and DTOs
│   │   └── service/                # Business logic, validation, stored procedure execution
│   ├── tests/
│   │   └── e2e/                    # Comprehensive acceptance and integration test suites
│   ├── certs/                      # RSA/PEM private/public keys for JWT signing
│   ├── dist/                       # Output directory for Windows cross-compiled binary (server.exe) & web.config
│   ├── postman/                    # Postman collection & environment (v1 hierarchy, UAT tests, sample responses)
│   ├── Dockerfile                  # Production & staging container build definition
│   └── go.mod                      # Go module dependencies
│
├── vue/                            # Vue 3 + TypeScript Frontend Single Page Application
│   ├── src/
│   │   ├── api/                    # Axios REST client endpoints organized by feature module
│   │   ├── components/             # Reusable Vue components (forms, tables, cropper, popups)
│   │   ├── layouts/                # MainLayout (header, sidebar, offline notification, speed test)
│   │   ├── router/                 # Vue Router configurations and authentication guards
│   │   ├── stores/                 # Pinia state stores (auth, theme)
│   │   ├── strategies/             # Custom Service Worker caching (DynamicNetworkCache, LowNetworkCache)
│   │   ├── types/                  # TypeScript interfaces and domain types
│   │   ├── views/                  # View pages (Admin, Master Data, Transactions, Reports, Login)
│   │   ├── sw.ts                   # Service worker registration and lifecycle hooks
│   │   └── main.ts                 # Vue application mount and global styles
│   ├── pentest.js                  # Grafana k6 load test script for parallel batch pagination testing
│   ├── package.json                # Frontend dependencies and npm scripts
│   └── vite.config.ts              # Vite bundler, PWA configuration, compression, and auto-imports
│
└── backup/                         # Database schema definition, export scripts, and seed files
    ├── db_export_structure.sql     # Reference MSSQL database structure
    └── export_to_sql.py            # Automated SQL schema exporter
```

---

## Getting Started

### Prerequisites

- **Docker & Docker Compose**: Recommended for local development to spin up MSSQL Server, Redis, and hot-reload containers.
- **Go 1.25+**: Required if running the backend directly on your host machine.
- **Node.js 20+ & npm**: Required if running the frontend directly on your host machine.

### Environment Configuration

1. **Gin Backend** (`gin/.env`):
   ```env
   PORT=8080
   GIN_MODE=debug
   BASE_URL=/
   JWT_SECRET=/app/certs/dev-private-key.pem
   REFRESH_SECRET=/app/certs/dev-private-key.pem
   DATABASE_DSN="server=mssql;port=1433;user id=sa;password=YourStrong@Passw0rd123;database=master;encrypt=disable;"
   REDIS_ADDR=redis:6379
   ```

2. **Vue Frontend** (`vue/.env`):
   ```env
   VITE_API_BASE_URL=http://localhost:8080
   VITE_DNS_PREFETCH=https://speed.cloudflare.com
   ```

### Starting Services with Docker Compose

To launch the complete development environment (MSSQL Server, Redis, Gin Backend, and Vue Dev Server):

```bash
# Start primary development stack
docker compose up -d

# Verify all containers are running and healthy
docker compose ps
```

The services will be accessible at:
- **Vue Frontend**: [http://localhost:5173](http://localhost:5173)
- **Gin REST API**: [http://localhost:8080](http://localhost:8080)
- **MSSQL Server**: `localhost:1433`
- **Redis Cache**: `localhost:6379`

### Running the Backend Locally (Go)

If running outside Docker:

```bash
cd gin
go mod download
go run ./cmd/server
```

### Running the Frontend Locally (Vue 3)

If running outside Docker:

```bash
cd vue
npm install
npm run dev
```

---

## Testing & Quality Assurance

### Go Backend Tests

Run unit and integration tests across the Gin codebase:

```bash
cd gin

# Run all unit and service tests
go test -v ./internal/...

# Run specific E2E test suites
go test -v ./tests/e2e/ -run TestMasterDataE2E
```

### Frontend TypeScript Verification & Build

Verify type integrity and compile the production bundle:

```bash
cd vue

# Type-check without emitting files
npm run type-check

# Compile production bundle with Service Worker
npm run build
```

### Performance & Security Auditing (On-Demand)

```bash
# Execute k6 concurrent load testing
docker compose --profile tools run --rm k6

# Run Semgrep static security audit on node_modules
docker compose --profile tools run --rm semgrep
```

---

## Production Windows IIS Deployment

GuangJiApps is engineered to run in a Windows Server IIS environment via the **HttpPlatformHandler** module.

1. **Cross-Compile Windows Executable**:
   Execute the dedicated Docker build container to generate `server.exe`:
   ```bash
   docker compose --profile build run --rm gin-build
   ```
   This outputs the optimized Windows binary directly to [gin/dist/server.exe](./gin/dist/).

2. **IIS Configuration (`gin/dist/web.config`)**:
   The provided [web.config](./gin/dist/web.config) orchestrates `server.exe` under IIS:
   - Dynamic port binding via `%HTTP_PLATFORM_PORT%`.
   - Tuned runtime settings (`GOGC=100`, `GOMEMLIMIT=512MiB`).
   - Logging redirected to `.\logs\gin.log`.
   - Reverse proxy client IP preservation and HTTP/3 `Alt-Svc` headers.

3. **Frontend Production Build**:
   ```bash
   cd vue
   npm run build:client
   ```
   Deploy the generated `dist/` directory to the IIS web application root or virtual directory.

---

## Coding Style & Development Guidelines

These rules and conventions must be strictly followed when contributing to the codebase:

### Backend (Golang / Gin)

1. **Framework & Architecture**:
   - Use **Gin-Gonic** for all RESTful endpoints following standard HTTP verbs.
   - Maintain modular separation in `internal/service`, `internal/api`, and `internal/domain`.
2. **Memory & Garbage Collection Optimization**:
   - **Minimize GC pressure**: Utilize `sync.Pool` for reusable, short-lived structures.
   - **Pre-allocation**: Always pre-allocate slices and maps when capacity is known (`make([]T, 0, capacity)`).
   - **Runtime Tuning**: Respect `GOGC` and `GOMEMLIMIT` configurations in production deployment.
3. **Database & Concurrency**:
   - Leverage **Concurrent GORM** with connection pooling (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`).
   - Use `wg.Go` instead of raw `go func` unless the task is explicitly fire-and-forget.
   - Handle stored procedures cleanly using row scanning helpers (`nullFieldScanner`) that tolerate NULL columns safely.
4. **Validation & Errors**:
   - Use `go-playground/validator/v10` for request validation.
   - Return structured validation error payloads: `{"error": "Validation failed", "details": {"field": ["message"]}}`.
   - Never expose raw database errors or stack traces to end users.
5. **SSRS Reporting**:
   - Use `excelize` for generating and reading `.rdl` reports from SQL Server Reporting Services.

### Frontend (Vue 3 / TypeScript)

1. **Form Submission State**:
   - On `submitForm` or `handleSubmit` (Create/Update operations), **always disable the page or submit button** while in-flight.
   - Always clear loading/submitting states inside a `finally` block:
     ```typescript
     try {
       submitting.value = true;
       await api.save(formData);
     } finally {
       submitting.value = false;
     }
     ```
2. **Network & Request Optimization**:
   - Eliminate redundant network requests; utilize cached stores or Service Worker strategies where applicable.
   - Cancel orphaned in-flight requests using `AbortController` during rapid navigation or pagination.
3. **Strict TypeScript**:
   - **Avoid using `any`**: Always define precise interfaces in `src/types/` for all API payloads and view states.
4. **Component Design**:
   - Follow Element Plus guidelines with responsive breakpoints (`el-row`, `el-col`).
   - Use theme-aware styles and dynamic CSS variables for autofill and dark mode support.

### Documentation & Workflow

- **Document Completed Work**: As soon as a task is finished and verified, immediately record a summary of the changes in [DOCUMENTATION.md](./DOCUMENTATION.md). Only document completed and tested tasks.
- **Maintain Postman Sync**: Whenever endpoints, request payloads, or parameters are added or modified, immediately update the Postman collection at [gin/postman/apps-gin.postman_collection.json](./gin/postman/apps-gin.postman_collection.json).

---

## Postman API Collection & Testing

A complete Postman test suite and environment configuration is maintained under `gin/postman/`:

- **Collection**: [`apps-gin.postman_collection.json`](./gin/postman/apps-gin.postman_collection.json).
- **Environment**: [`apps-gin.postman_environment.json`](./gin/postman/apps-gin.postman_environment.json).
- **Features**:
  - **Version `v1`**: Collection metadata versioned at `v1` with `apiVersion` variable.
  - **Nested CRUD Folders**: All entities in Admin Management, Master Data, and Transactions are categorized into dedicated sub-folders.
  - **Collection-Wide Headers**: Automatic injection of `Content-Type: application/json` and `Accept: application/json`.
  - **Auto-Login Pre-Request Script**: Automatically detects missing tokens, executes login with default credentials, and sets the `Authorization` header.
  - **HTTP 401 Auto-Retry**: Automatically re-authenticates and retries the original request if an authorization token expires.
  - **Full UAT Coverage**: 100% of endpoints include automated test assertions and mock response examples.

---

## License

This software is governed by a **Custom Proprietary & Restricted License**, available in 3 language editions:
- [English (en)](./LICENSE.md)
- [Bahasa Indonesia (id)](./LICENSE.id.md)
- [简体中文 (zh)](./LICENSE.zh.md)

Key Terms:
- Specifically designated for authorized non-profit foundations with prior written permission from the Developer.
- Testing, integration evaluations, and security audits within authorized environments are permitted.
- Duplication, public cloning, redistribution, commercialization, or derivative distribution of this repository is strictly prohibited without prior written consent from the Developer.


