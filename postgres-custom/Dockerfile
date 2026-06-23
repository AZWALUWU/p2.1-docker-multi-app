# Menggunakan official image PostgreSQL berbasis Alpine Linux agar aman dan ringan
FROM postgres:16-alpine

# MENGENAL ENVIRONMENT VARIABLES (Default value untuk konfigurasi awal DB)
# Catatan: Di dunia nyata/production, jangan hardcode password di Dockerfile! 
# Gunakan .env atau secret manager. Ini hanya sebagai nilai default instruksional.
ENV POSTGRES_DB=assignment_db
ENV POSTGRES_USER=admin
ENV POSTGRES_PASSWORD=supersecret

# MENGENAL DOCKER EXTENSION / INIT SCRIPTS
# Jalur '/docker-entrypoint-initdb.d/' adalah folder khusus pada image Postgres resmi.
# Setiap file .sql, .sql.gz, atau .sh di folder ini akan dieksekusi secara alfabetis 
# HANYA saat container pertama kali diinisialisasi (ketika direktori data masih kosong).
COPY init.sql /docker-entrypoint-initdb.d/

# VOLUME MOUNTING DOCUMENTATION
# Menandakan bahwa folder data postgres harus di-mount ke volume eksternal 
# agar data tidak hilang (persistent) saat container dimatikan atau dihapus.
VOLUME /var/lib/postgresql/data

# Port default PostgreSQL
EXPOSE 5432

# CMD vs ENTRYPOINT:
# Image postgres resmi menggunakan ENTRYPOINT untuk menjalankan script setup internal (docker-entrypoint.sh).
# CMD di sini diteruskan sebagai argumen ke entrypoint tersebut untuk menjalankan server database utama ("postgres").
CMD ["postgres"]
