#!/bin/bash

CONTAINER_NAME="mssql_server"
SA_PASSWORD="YourStrong@Passw0rd123"
DB_NAME="api"
BACKUP_FILE="/var/opt/mssql/backup/db_20260621"

echo "=== 1. Cek Logical File Names dari .bak ==="
/opt/mssql-tools18/bin/sqlcmd \
  -S localhost -U sa -P "$SA_PASSWORD" -C \
  -Q "RESTORE FILELISTONLY FROM DISK = N'$BACKUP_FILE';"

echo ""
echo "=== 2. Jalankan perintah RESTORE (Sesuaikan nama logical file jika berbeda) ==="
# Catatan: Ganti 'LogicalDataName' dan 'LogicalLogName' sesuai output dari perintah di atas
/opt/mssql-tools18/bin/sqlcmd \
  -S localhost -U sa -P "$SA_PASSWORD" -C \
  -Q "RESTORE DATABASE [$DB_NAME] FROM DISK = N'$BACKUP_FILE' WITH MOVE N'LogicalDataName' TO N'/var/opt/mssql/data/${DB_NAME}.mdf', MOVE N'LogicalLogName' TO N'/var/opt/mssql/data/${DB_NAME}_log.ldf', REPLACE, STATS = 5;"

echo "Proses import selesai!"