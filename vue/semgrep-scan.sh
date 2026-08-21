#!/bin/bash

TARGET_DIR="node_modules"
REPORT_FILE="semgrep-security-report.json"

# Validasi keberadaan node_modules
if [ ! -d "$TARGET_DIR" ]; then
    echo "[!] Error: Direktori '$TARGET_DIR' tidak ditemukan. Jalankan 'npm install' terlebih dahulu."
    exit 1
fi

echo "[+] Memulai pemindaian keamanan pada seluruh folder $TARGET_DIR..."
echo "[+] Menggunakan rule: p/security-audit"

# Menjalankan Semgrep dengan format JSON untuk dokumentasi/analisis lanjutan
semgrep --config "p/security-audit" "$TARGET_DIR" --json --output "$REPORT_FILE"

if [ $? -eq 0 ]; then
    echo "[✓] Pemindaian selesai. Laporan tersimpan di $REPORT_FILE"
else
    echo "[!] Pemindaian selesai dengan temuan kerentanan atau peringatan."
    echo "[✓] Laporan tersimpan di $REPORT_FILE"
fi
