# ==========================================
# STAGE 1: Build & Dependency Resolution
# ==========================================
# Menggunakan versi LTS (Long Term Support) berbasis Alpine Linux agar ukuran dasar image kecil dan aman.
FROM node:20-alpine AS builder

# Menentukan working directory di dalam container untuk proses isolasi.
WORKDIR /app

# Menyalin package.json terlebih dahulu (Urutan Instruksi & Layer Caching).
# Docker merekam layer ini. Jika package.json tidak berubah, Docker akan menggunakan cache
# dan melewati proses 'npm install' pada build berikutnya, sehingga build menjadi sangat cepat.
COPY package.json ./

# Menginstall dependensi. Menggunakan --only=production jika hanya butuh runtime dependencies,
# namun di sini kita install standar untuk memastikan semua modul terpasang di stage builder.
RUN npm install

# Menyalin sisa source code aplikasi setelah dependensi selesai diinstall.
# Jika ada perubahan pada server.js, layer ini saja yang di-rebuild, layer npm install tetap aman.
COPY . .

# ==========================================
# STAGE 2: Production Runtime Environment
# ==========================================
FROM node:20-alpine AS runner

WORKDIR /app

# Menentukan environment variable default
ENV NODE_ENV=production
ENV PORT=3000

# Perubahan di sini: Kita copy node_modules terlebih dahulu
COPY --from=builder /app/node_modules ./node_modules

# Daripada satu-satu, kita copy seluruh source code yang sudah di-filter .dockerignore dari builder
COPY --from=builder /app/ .

# KEAMANAN: Menggunakan user non-root
USER node

# Dokumentasi Port
EXPOSE 3000

ENTRYPOINT ["npm"]
CMD ["start"]
