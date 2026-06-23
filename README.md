# Complete Project Guide: Containerizing 5 Different Application Types (Monorepo Setup)

This repository serves as a comprehensive monorepo containing production-ready Docker configurations for five structurally distinct application architectures: **Node.js (Express) API**, **Python (Flask) API**, **React SPA (Nginx)**, **Go (Golang) Binary**, and a **Custom PostgreSQL Database**.

Each subdirectory demonstrates specific containerization mechanics, layer caching strategies, multi-stage compilation workflows, and specialized security principles designed to minimize final image footprint and attack surface.

---

## Global Build Optimization: The `.dockerignore` File

Before analyzing individual applications, a global `.dockerignore` file must be placed at the root of the monorepo (`docker-examples/`). This file strictly limits the **Docker Build Context**, ensuring that massive local directories (like `node_modules`), virtual environments, sensitive credentials, and operating system artifacts are not sent to the Docker daemon. This dramatically speeds up compilation and prevents security leaks.

Create `.dockerignore` in the root folder:

```text
# Git
.git
.gitignore
README.md

# Node.js
**/node_modules
**/npm-debug.log

# Python
**/__pycache__
**/*.pyc
**/*.pyo
**/*.pyd
**/.venv
**/venv

# Go
**/bin
**/dist

# OS files
**/.DS_Store
**/Thumbs.db

```

---

## 1. Node.js (Express) API Containerization

### Objective

Containerize a runtime-interpreted Node.js application utilizing **multi-stage builds** to isolate dependency installation from production runtime execution, and enforce the principle of least privilege using a non-root user.

### Project Files

`node-api/package.json`

```json
{
  "name": "node-api",
  "version": "1.0.0",
  "description": "Simple Node.js API for Docker assignment",
  "main": "server.js",
  "scripts": {
    "start": "node server.js"
  },
  "dependencies": {
    "express": "^4.19.2"
  }
}

```

`node-api/server.js`

```javascript
const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

app.get('/', (req, res) => {
  res.json({
    status: "success",
    message: "Halo dari Node.js API di dalam Docker!",
    environment: process.env.NODE_ENV || "development"
  });
});

app.listen(PORT, () => {
  console.log(`Server berjalan di port ${PORT}`);
});

```

### Deeply Annotated Dockerfile

`node-api/Dockerfile`

```dockerfile
# ==========================================
# STAGE 1: Build & Dependency Resolution
# ==========================================
# Using LTS version on Alpine Linux for a minimal base footprint and reduced vulnerabilities.
FROM node:20-alpine AS builder

# Establishes an isolated working directory within the container filesystem.
WORKDIR /app

# Copying package management metadata first to harness Docker's Layer Caching mechanism.
# If package.json remains unchanged, Docker bypasses 'npm install' on subsequent runs.
COPY package.json ./

# Installs all node modules defined in the manifests.
RUN npm install

# Copies the rest of the application source code into the builder stage environment.
COPY . .

# ==========================================
# STAGE 2: Production Runtime Environment
# ==========================================
# Starting fresh with a pristine alpine runtime layer to drop development overhead and npm cache.
FROM node:20-alpine AS runner

WORKDIR /app

# Injecting standard production environment configurations.
ENV NODE_ENV=production
ENV PORT=3000

# Selectively copying pre-installed modules from the builder layer.
COPY --from=builder /app/node_modules ./node_modules

# Copying the verified clean source code artifacts from the builder layer.
COPY --from=builder /app/ .

# SECURITY: Enforcing Least Privilege Access Control.
# Docker defaults to the 'root' user. The node-alpine base provides a built-in unprivileged user 
# named 'node'. Switching to it mitigates severe container-breakout exploitation risks.
USER node

# Metadata exposing the designated communication port to network abstractions.
EXPOSE 3000

# CMD vs ENTRYPOINT Distinction:
# ENTRYPOINT establishes the absolute command executable bin that runs upon startup.
# CMD serves as mutable, appendable default arguments passed directly into the Entrypoint.
ENTRYPOINT ["npm"]
CMD ["start"]

```

### Verification & Lifecycle Commands

```bash
# Execute compilation from the monorepo root directory
docker build -t node-api:v1 ./node-api

# Spin up container instance binding container port 3000 to localhost port 3000
docker run -d -p 3000:3000 --name my-node-app node-api:v1

# Query API via terminal to verify health check status
curl http://localhost:3000

# Remove container instance and release bound network ports
docker rm -f my-node-app

```

---

## 2. Python (Flask) API Containerization

### Objective

Dockerize an interpreted Python application while addressing Python's dependency compilation layer requirements. We extract pre-compiled `.whl` (wheels) binaries within an isolated build phase and perform an entirely offline installation during the production phase.

### Project Files

`python-flask/requirements.txt`

```text
Flask==3.0.3
werkzeug==3.0.3

```

`python-flask/app.py`

```python
import os
from flask import Flask, jsonify

app = Flask(__name__)

@app.route('/')
def hello():
    return jsonify({
        "status": "success",
        "message": "Halo dari Python Flask di dalam Docker!",
        "framework": "Flask",
        "debug_mode": os.environ.get("FLASK_DEBUG", "False")
    })

if __name__ == '__main__':
    port = int(os.environ.get("PORT", 5000))
    app.run(host='0.0.0.0', port=port)

```

### Deeply Annotated Dockerfile

`python-flask/Dockerfile`

```dockerfile
# ==========================================
# STAGE 1: Build & Dependency Wheel Generation
# ==========================================
FROM python:3.11-alpine AS builder

WORKDIR /app

# Optimization: Inhibits python compilation of bytecode (.pyc) onto virtual disk 
# and enforces immediate unbuffered pipeline streaming of logs directly to standard out.
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Installs essential C-compilation tools needed to natively compile complex extension modules.
RUN apk add --no-cache gcc musl-dev linux-headers

COPY requirements.txt .

# Resolves dependencies and bundles them into highly optimized local platform-specific wheel files.
# This stage isolates compiler bloat from ever reaching the production runtime platform.
RUN pip wheel --no-cache-dir --wheel-dir /app/wheels -r requirements.txt

# ==========================================
# STAGE 2: Production Runtime Environment
# ==========================================
FROM python:3.11-alpine AS runner

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PORT=5000

# Imports wheels cache directory from the decoupled compilation sandbox.
COPY --from=builder /app/wheels /workspace/wheels
COPY --from=builder /app/requirements.txt .

# Executes fully offline setup by pulling wheels locally without attempting network indexing (PyPI).
RUN pip install --no-cache-dir --no-index --find-links=/workspace/wheels -r requirements.txt \
    && rm -rf /workspace/wheels

COPY . .

# SECURITY: Spawns an isolated unprivileged user profile and restricts execution space permissions.
RUN adduser -D appuser && chown -R appuser:appuser /app
USER appuser

EXPOSE 5000

ENTRYPOINT ["python"]
CMD ["app.py"]

```

### Verification & Lifecycle Commands

```bash
docker build -t python-flask:v1 ./python-flask
docker run -d -p 5000:5000 --name my-flask-app python-flask:v1
curl http://localhost:5000
docker rm -f my-flask-app

```

---

## 3. React SPA (Served with Nginx)

### Objective

Containerize a static frontend client architecture. Node.js engine tooling is required exclusively at the early build stage to transpile, bundle, and compile raw JSX asset directories down into vanilla assets. Production environments drop Node.js entirely and inherit a high-performance web-server engine architecture via **Nginx**.

### Project Files

`react-nginx/package.json`

```json
{
  "name": "react-nginx-app",
  "version": "1.0.0",
  "private": true,
  "dependencies": {
    "react": "^18.3.1",
    "react-dom": "^18.3.1"
  },
  "scripts": {
    "build": "echo '<html><head><title>React Docker App</title></head><body style=\"background:#282c34;color:white;font-family:sans-serif;text-align:center;padding-top:50px;\"><h1>Halo dari React SPA + Nginx di Docker!</h1><p>Aplikasi statis ini di-serve dengan sangat efisien oleh Nginx.</p></body></html>' > index.html && mkdir -p dist && mv index.html dist/"
  }
}

```

`react-nginx/nginx.conf`

```nginx
server {
    listen 80;
    server_name localhost;

    location / {
        root /usr/share/nginx/html;
        index index.html index.htm;
        # Crucial fallback mechanism for Single Page Applications (SPA).
        # Re-routes deep browser URL entries back into internal application bundle entry points.
        try_files $uri $uri/ /index.html;
    }

    error_page 500 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}

```

### Deeply Annotated Dockerfile

`react-nginx/Dockerfile`

```dockerfile
# ==========================================
# STAGE 1: Compilation & Build Environment
# ==========================================
FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .

# Compiles JSX/TypeScript structure patterns into standard cross-browser files inside /app/dist
RUN npm run build

# ==========================================
# STAGE 2: Production Web Server Environment
# ==========================================
# Drops heavy dev tools and transitions directly to ultra-lean alpine Nginx images.
FROM nginx:1.25-alpine AS runner

# Injects custom virtual hosts block configuration over the internal default configuration template.
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copies over ONLY the compiled assets from the compilation container stage.
# Places code directly into Nginx's targeted root filesystem endpoint location.
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

# Nginx image provides standard internal Entrypoint execution loops. 
# Explicitly forcing daemon flag off ensures the task loop attaches forever inside foreground logs.
CMD ["nginx", "-g", "daemon off;"]

```

### Verification & Lifecycle Commands

```bash
docker build -t react-nginx:v1 ./react-nginx
docker run -d -p 8080:80 --name my-react-app react-nginx:v1
# Visit http://localhost:8080 in a browser engine
docker rm -f my-react-app

```

---

## 4. Go (Golang) Binary Containerization

### Objective

Containerize a natively compiled static binary. Go generates completely autonomous binary formats, rendering system OS components, packages, runtime dependencies, and command shells utterly unnecessary during execution. We isolate and bundle the binary directly into a zero-byte filesystem abstraction (**`scratch`**), yielding an image with maximum security and an incredibly small size.

### Project Files

`go-binary/main.go`

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Lang    string `json:"language"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res := Response{
			Status:  "success",
			Message: "Halo dari Go Binary di dalam container SCRATCH (Kosong)!",
			Lang:    "Go (Golang)",
		}
		json.NewEncoder(w).Encode(res)
	})

	fmt.Printf("Server Go berjalan di port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Gagal menjalankan server: %v\n", err)
	}
}

```

### Deeply Annotated Dockerfile

`go-binary/Dockerfile`

```dockerfile
# ==========================================
# STAGE 1: Compilation Environment
# ==========================================
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY main.go .

# Compiling the Go Binary with optimized environment tags:
# CGO_ENABLED=0   -> Disables C bindings, stripping glibc links to create a fully self-contained static binary.
# GOOS=linux      -> Forces target compilation pattern matching execution configurations for the Linux kernel.
# -ldflags="-s -w" -> Strips debug symbols, symbol tables, and DWARF tracking headers to reduce binary footprint.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o myserver main.go

# ==========================================
# STAGE 2: Ultra-minimal Runtime Environment
# ==========================================
# 'scratch' is Docker's empty base layer (0 bytes). No bash, no vulnerabilities, pure binary execution.
FROM scratch AS runner

WORKDIR /

# Pulls over exclusively the naked independent binary file from the build phase.
COPY --from=builder /app/myserver /myserver

ENV PORT=8080
EXPOSE 8080

# Because scratch lacks a shell environment (/bin/sh), we MUST use the array exec format 
# to trigger the binary directly without shell parsing.
ENTRYPOINT ["/myserver"]

```

### Verification & Lifecycle Commands

```bash
docker build -t go-binary:v1 ./go-binary
docker run -d -p 8081:8080 --name my-go-app go-binary:v1
curl http://localhost:8081

# Optimization check: Verify the incredibly low storage usage of the scratch build
docker images | grep go-binary

docker rm -f my-go-app

```

---

## 5. PostgreSQL Database with Custom Initial Scripts

### Objective

Containerize a stateful data layer. Databases store persistent data, meaning they rely on **Volume Mounting** rather than internal application compilation. This setup uses custom lifecycle initialization points inside official image configurations to automate database schema generation and initial data seed routines during initial system bootstrap loops.

### Project Files

`postgres-custom/init.sql`

```sql
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    project_name VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO projects (project_name, status) VALUES
('Dockerize Node.js API', 'Completed'),
('Dockerize Python Flask', 'Completed'),
('Dockerize React SPA', 'Completed'),
('Dockerize Go Binary', 'Completed'),
('Dockerize PostgreSQL Custom', 'In Progress');

```

### Deeply Annotated Dockerfile

`postgres-custom/Dockerfile`

```dockerfile
# Using official validated PostgreSQL image wrapped on top of minimal Alpine secure footprints.
FROM postgres:16-alpine

# ENVIRONMENT VARIABLES CONFIGURATION:
# Injects standard base cluster definition metadata values to configure database startup parameters.
# WARNING: Avoid hardcoding secret credentials directly into a production Dockerfile. 
# Pass them dynamically via runtime engines or secrets managers instead.
ENV POSTGRES_DB=assignment_db
ENV POSTGRES_USER=admin
ENV POSTGRES_PASSWORD=supersecret

# HOOK INITIALIZATION AUTOMATION:
# The official entrypoint script scans '/docker-entrypoint-initdb.d/' upon first cluster boot.
# All nested .sql scripts found here execute sequentially to instantiate schemas and seed default entries.
COPY init.sql /docker-entrypoint-initdb.d/

# PERSISTENT DATA LAYER SPECIFICATION:
# Declares a directory point linking structural data blocks directly into out-of-container engine volumes.
# This ensures application data persists when the transient container instance lifecycle restarts or tears down.
VOLUME /var/lib/postgresql/data

EXPOSE 5432

# The parent upstream base image sets up a complex entrypoint initialization bash script wrapper. 
# CMD acts as parameter strings instructing that wrapper execution stack to spawn the server listener daemon.
CMD ["postgres"]

```

### Verification & Lifecycle Commands

```bash
docker build -t postgres-custom:v1 ./postgres-custom

# Run container and expose engine standard ports externally
docker run -d -p 5432:5432 --name my-postgres-db postgres-custom:v1

# Verification: Interrogate internal database layer via client binary tool executed internally
# This confirms the automatic execution of the initialization schema seeding logic
docker exec -it my-postgres-db psql -U admin -d assignment_db -c "SELECT * FROM projects;"

docker rm -f my-postgres-db

```

---

## Technical Summary Matrix

| Framework/Type | Base Target Image | Structural Strategy | Key Security Enhancement | Build Speed Drivers |
| --- | --- | --- | --- | --- |
| **Node.js Express** | `node:20-alpine` | Multi-Stage Build | Explicit Non-Root `USER node` | Layer caching of `package.json` |
| **Python Flask** | `python:3.11-alpine` | Multi-Stage (Wheels Isolation) | Non-Root Workspace `adduser` | Isolated dependency compilation |
| **React Frontend** | `nginx:1.25-alpine` | Asset Build Transpilation Split | Minimal Server Footprint Only | Discarding entire compilation tree |
| **Go Native** | `scratch` | Pure Autonomous Monolithic Binary | `scratch` (Zero Shell Vulnerability) | `-ldflags="-s -w"` optimization flags |
| **PostgreSQL DB** | `postgres:16-alpine` | Ephemeral Initialization Hooks | External Volume Persistence | `/docker-entrypoint-initdb.d/` |
