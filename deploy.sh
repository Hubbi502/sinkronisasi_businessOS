#!/bin/bash

# Script Deploy - Sinkronisasi DB untuk server Ubuntu
# Pastikan sudah berada di directory project sebelum menjalankan script ini

set -e

echo "====================================="
echo "Mulai Proses Deployment Sinkronisasi DB..."
echo "====================================="

# 1. Cek apakah Docker terinstall
if ! command -v docker &> /dev/null; then
    echo "[-] Docker belum terinstall. Menginstall Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    echo "[+] Docker berhasil diinstall."
else
    echo "[+] Docker sudah terinstall."
fi

# 2. Cek apakah Docker Compose tersedia (v2 menggunakan 'docker compose')
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
elif docker-compose --version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
else
    echo "[-] Docker Compose tidak ditemukan. Harap update Docker ke versi yang mendukung plugin 'docker compose'."
    exit 1
fi

# 3. Pull source code terbaru dari git (jika diperlukan)
echo "[ ] Pulling revisi terbaru dari git branch main... (Uncomment jika menggunakan git)"
# git pull origin main

# 4. Build dan restart services menggunakan Docker Compose
echo "[ ] Membangun (Build) image Docker dan merestart containers..."
sudo $DOCKER_COMPOSE_CMD up -d --build

echo "====================================="
echo "[+] Deployment berhasil selesai!"
echo "[+] Service API dan Worker seharusnya sudah berjalan."
echo "[+] Untuk melihat log, jalankan: sudo $DOCKER_COMPOSE_CMD logs -f"
echo "====================================="
