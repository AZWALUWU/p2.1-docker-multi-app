# Project 2.1 — Monorepo: Dockerize 5 Different Application Types

This repository serves as a comprehensive monorepo containing five distinct application types, each containerized using unique Dockerfile strategies tailored to their specific technology stack, runtime requirements, and security profiles. 

## Objectives
* Understand instruction tuning and structural choices within a `Dockerfile`.
* Optimize build context size and security using `.dockerignore`.
* Leverage multi-stage builds to dramatically reduce final production image sizes.
* Distinguish between `CMD` and `ENTRYPOINT` in practice.
* Implement port mapping, environment variable injections, and initialization automation.

---

## Global Build Context Configuration

To prevent sensitive files, operating system artifacts, and massive dependency directories (like `node_modules`) from bloating the Docker build context, a global `.dockerignore` is placed at the root of the project.

### Root `.dockerignore`
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
````

## 1. Node.js Express API (`/node-api`)

### Strategy: Multi-Stage Dependency Isolation

Node.js applications often accumulate heavy development tools during package installations. This setup isolates the build tools in a `builder` layer and ports only the necessary dependencies and server files to an ultra-lean `runner` layer.

### Source Files

**`package.json`**

JSON

```
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

**`server.js`**

JavaScript

```
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

### Dockerfile

Dockerfile

```
# ==========================================
# STAGE 1: Build & Dependency Resolution
# ==========================================
FROM node:20-alpine AS builder
WORKDIR /app

# Layer Caching Optimization: copy dependencies first
COPY package.json ./
RUN npm install
COPY . .

# ==========================================
# STAGE 2: Production Runtime Environment
# ==========================================
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000

# Copy node_modules and pre-filtered files from the builder stage
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/ .

# Principle of Least Privilege: Switch away from root user
USER node

EXPOSE 3000

# ENTRYPOINT acts as the core command, CMD provides default arguments
ENTRYPOINT ["npm"]
CMD ["start"]
```

### CLI Commands

Bash

```
# Build the image
docker build -t node-api:v1 ./node-api

# Run the container with port forwarding
docker run -d -p 3000:3000 --name my-node-app node-api:v1
```

## 2. Python Flask API (`/python-flask`)

### Strategy: Offline Wheel Compilation

Python apps require dependencies that sometimes require compilation hooks (`gcc`, `musl-dev`). This Dockerfile compiles the dependencies into wheel binaries (`.whl`) within the builder phase, and installs them natively inside an offline, clean production image.

### Source Files

**`requirements.txt`**

Plaintext

```
Flask==3.0.3
werkzeug==3.0.3
```

**`app.py`**

Python

```
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

### Dockerfile

Dockerfile

```
# ==========================================
# STAGE 1: Build & Dependency Wheel Generation
# ==========================================
FROM python:3.11-alpine AS builder
WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Install heavy compiler tools required for building Python libraries
RUN apk add --no-cache gcc musl-dev linux-headers
COPY requirements.txt .

# Compile and store full dependency packages into a local workspace directory
RUN pip wheel --no-cache-dir --wheel-dir /app/wheels -r requirements.txt

# ==========================================
# STAGE 2: Production Runtime Environment
# ==========================================
FROM python:3.11-alpine AS runner
WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PORT=5000

COPY --from=builder /app/wheels /workspace/wheels
COPY --from=builder /app/requirements.txt .

# Install packages offline from the local directory to completely avoid internet dependencies
RUN pip install --no-cache-dir --no-index --find-links=/workspace/wheels -r requirements.txt \
    && rm -rf /workspace/wheels

COPY . .

# Security: Create a non-privileged system user
RUN adduser -D appuser && chown -R appuser:appuser /app
USER appuser

EXPOSE 5000

ENTRYPOINT ["python"]
CMD ["app.py"]
```

### CLI Commands

Bash

```
# Build the image
docker build -t python-flask:v1 ./python-flask

# Run the container
docker run -d -p 5000:5000 --name my-flask-app python-flask:v1
```

## 3. React SPA Hosted via Nginx (`/react-nginx`)

### Strategy: Static Asset Separation

A frontend single-page application (SPA) only needs Node.js for source compilation (bundling JSX/TSX into assets). In production, the runtime requires nothing but a high-performance web server like Nginx to serve static elements (`html`, `css`, `js`).

### Source Files

**`package.json`**

JSON

```
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

**`nginx.conf`**

Nginx

```
server {
    listen 80;
    server_name localhost;

    location / {
        root /usr/share/nginx/html;
        index index.html index.htm;
        try_files $uri $uri/ /index.html;
    }

    error_page 500 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
```

### Dockerfile

Dockerfile

```
# ==========================================
# STAGE 1: Compilation & Build Environment
# ==========================================
FROM node:20-alpine AS builder
WORKDIR /app

COPY package.json ./
RUN npm install
COPY . .
RUN npm run build

# ==========================================
# STAGE 2: Production Web Server Environment
# ==========================================
FROM nginx:1.25-alpine AS runner

# Inject custom configuration to resolve React-Router 404 navigation errors
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Extract ONLY static html/css files from the node build agent
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

# Keep the Nginx service processing in the foreground to prevent immediate container lifecycle death
CMD ["nginx", "-g", "daemon off;"]
```

### CLI Commands

Bash

```
# Build the image
docker build -t react-nginx:v1 ./react-nginx

# Run the container
docker run -d -p 8080:80 --name my-react-app react-nginx:v1
```

## 4. Go Native Binary (`/go-binary`)

### Strategy: Zero-Byte Distroless Environment (`scratch`)

Go compiles applications down to runtime-agnostic, self-sufficient binary executables. By eliminating CGO dependencies, the app can run on top of an entirely blank slate base image named `scratch`. This drops image sizes to merely the size of the compiled program (~10-15MB) with zero shell dependencies.

### Source Files

**`main.go`**

Go

```
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

### Dockerfile

Dockerfile

```
# ==========================================
# STAGE 1: Compilation Environment
# ==========================================
FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY main.go .

# Compilation optimization:
# CGO_ENABLED=0 disables C dependencies, rendering the binary truly standalone
# -ldflags="-s -w" strips debugging symbols, shrinking executable weight drastically
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o myserver main.go

# ==========================================
# STAGE 2: Ultra-minimal Runtime Environment
# ==========================================
FROM scratch AS runner
WORKDIR /

# Bring in only the final standalone runtime execution file
COPY --from=builder /app/myserver /myserver

ENV PORT=8080
EXPOSE 8080

# Because `scratch` has no underlying operating system shell, 
# you MUST use the bracketed Exec configuration format to run the engine directly.
ENTRYPOINT ["/myserver"]
```

### CLI Commands

Bash

```
# Build the image
docker build -t go-binary:v1 ./go-binary

# Run the container
docker run -d -p 8081:8080 --name my-go-app go-binary:v1
```

## 5. PostgreSQL with Automatic Seed Scripts (`/postgres-custom`)

### Strategy: Stateful Automation Extensions

Databases are stateful layers requiring volume mounts to make data persistent. Official database images recognize the entry path `/docker-entrypoint-initdb.d/`. Throwing SQL execution blueprints inside this destination forces PostgreSQL to spin up schema objects automatically during initial startup.

### Source Files

**`init.sql`**

SQL

```
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

### Dockerfile

Dockerfile

```
FROM postgres:16-alpine

# Administrative Configuration via Environment Variables
ENV POSTGRES_DB=assignment_db
ENV POSTGRES_USER=admin
ENV POSTGRES_PASSWORD=supersecret

# Automated Initialization Schema Mounting
# File executions are evaluated strictly in alphabetical order on database instantiation
COPY init.sql /docker-entrypoint-initdb.d/

# Declare persistent mapping zone to ensure safety of stored files outside ephemeral storage
VOLUME /var/lib/postgresql/data

EXPOSE 5432

CMD ["postgres"]
```

### CLI Commands

Bash

```
# Build the image
docker build -t postgres-custom:v1 ./postgres-custom

# Run the container
docker run -d -p 5432:5432 --name my-postgres-db postgres-custom:v1

# Interactive data assertion to verify seed data ingestion inside the running engine
docker exec -it my-postgres-db psql -U admin -d assignment_db -c "SELECT * FROM projects;"
```

## Summary Comparison Matrix

|**Application Folder**|**Strategy Paradigm**|**Base Target Image**|**Highlight Feature**|**Approximate Size Profile**|
|---|---|---|---|---|
|`node-api`|Multi-Stage Build|`node:20-alpine`|Non-root configuration (`USER node`)|Minimal Runtime Node Layer (~180MB)|
|`python-flask`|Build wheels artifact|`python:3.11-alpine`|Offline isolated pip configurations|Lean Engine Footprint (~60MB)|
|`react-nginx`|Multi-Stage Compilation|`nginx:1.25-alpine`|Tailored SPA internal route handling|Super Lightweight Static Assets (~30MB)|
|`go-binary`|Distroless compilation|`scratch`|Complete reduction of vulnerability scope|Ultra Tiny Execution Shell (~15MB)|
|`postgres-custom`|Stateful Extension Hooks|`postgres:16-alpine`|Initialization migration parsing|Core Persistent State Layer (~250MB)|

