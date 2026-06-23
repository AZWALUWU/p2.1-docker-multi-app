# ==========================================
# STAGE 1: Build & Dependency Wheel Generation
# ==========================================
# Menggunakan image python alpine untuk meminimalkan ukuran awal
FROM python:3.11-alpine AS builder

WORKDIR /app

# Memasang environment variable agar Python tidak menulis file .pyc ke disk
# dan tidak menggunakan buffering untuk output log (log langsung keluar di terminal)
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Memasang tools yang dibutuhkan untuk compile beberapa library python (jika ada)
RUN apk add --no-cache gcc musl-dev linux-headers

# Copy requirements.txt secara terpisah demi pemanfaatan Docker Layer Cache
COPY requirements.txt .

# Mengunduh dan membuat format 'wheels' (package siap install) ke dalam folder lokal.
# Ini mencegah kita melakukan kompilasi ulang di stage production.
RUN pip wheel --no-cache-dir --wheel-dir /app/wheels -r requirements.txt

# ==========================================
# STAGE 2: Production Runtime Environment
# ==========================================
FROM python:3.11-alpine AS runner

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PORT=5000

# Menyalin hasil kompilasi package (wheels) dari stage builder
COPY --from=builder /app/wheels /workspace/wheels
COPY --from=builder /app/requirements.txt .

# Menginstall python packages dari lokal wheels tanpa mendownload ulang dari internet
RUN pip install --no-cache-dir --no-index --find-links=/workspace/wheels -r requirements.txt \
    && rm -rf /workspace/wheels

# Menyalin seluruh kode aplikasi ke working directory container
COPY . .

# KEAMANAN: Membuat user non-root bernama 'appuser' demi keamanan container
RUN adduser -D appuser && chown -R appuser:appuser /app
USER appuser

# Dokumentasi Port aplikasi Flask
EXPOSE 5000

# CMD vs ENTRYPOINT:
# Menggunakan python langsung sebagai entrypoint utama, dan nama file sebagai argument default (CMD)
ENTRYPOINT ["python"]
CMD ["app.py"]
