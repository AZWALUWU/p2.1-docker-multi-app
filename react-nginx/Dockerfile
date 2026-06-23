# ==========================================
# STAGE 1: Compilation & Build Environment
# ==========================================
FROM node:20-alpine AS builder

WORKDIR /app

# Memanfaatkan Layer Cache untuk dependensi node_modules
COPY package.json ./
RUN npm install

# Copy seluruh source code
COPY . .

# Menjalankan proses build untuk menghasilkan file statis di folder /app/dist
RUN npm run build

# ==========================================
# STAGE 2: Production Web Server Environment
# ==========================================
# Kita beralih ke Nginx berbasis Alpine yang sangat ringan (~5MB core size)
FROM nginx:1.25-alpine AS runner

# Menyalin file konfigurasi Nginx kustom kita ke dalam folder konfigurasi Nginx container
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Mengcopy HANYA artifact html/css/js statis hasil kompilasi Stage 1
# Tempat tujuan (/usr/share/nginx/html) adalah folder default Nginx mencari file statis
COPY --from=builder /app/dist /usr/share/nginx/html

# Nginx secara default berjalan di Port 80
EXPOSE 80

# CMD vs ENTRYPOINT:
# Image nginx bawaan sudah mengatur ENTRYPOINT untuk menjalankan web server di background.
# CMD di sini digunakan untuk memastikan nginx berjalan di foreground (daemon off)
# agar container tidak langsung mati setelah dijalankan.
CMD ["nginx", "-g", "daemon off;"]
