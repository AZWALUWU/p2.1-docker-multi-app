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
