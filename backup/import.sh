#!/bin/bash

#docker exec -it <container_name> /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "YourStrong@Passw0rd123" -C -Q "RESTORE DATABASE [guangji] FROM DISK = '/var/opt/mssql/backup/db_20260907' WITH MOVE 'LogikaNamaMDF' TO '/var/opt/mssql/data/guangji.mdf', MOVE 'LogikaNamaLDF' TO '/var/opt/mssql/data/guangji_log.ldf', REPLACE;"

#/opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "YourStrong@Passw0rd123" -C

CONTAINER_NAME="mssql_server"
SA_PASSWORD="YourStrong@Passw0rd123"
DB_NAME="guangji"
BACKUP_FILE="/var/opt/mssql/backup/db_20260907"

echo "=== 1. Cek Logical File Names dari .bak ==="
/opt/mssql-tools18/bin/sqlcmd \
  -S localhost -U sa -P "$SA_PASSWORD" -C \
  -Q "RESTORE FILELISTONLY FROM DISK = N'$BACKUP_FILE';"

echo ""
echo "=== 2. Jalankan perintah RESTORE (Sesuaikan nama logical file jika berbeda) ==="
# Catatan: Ganti 'LogicalDataName' dan 'LogicalLogName' sesuai output dari perintah di atas
/opt/mssql-tools18/bin/sqlcmd \
  -S localhost -U sa -P "$SA_PASSWORD" -C \
  -Q "RESTORE DATABASE [$DB_NAME] FROM DISK = N'$BACKUP_FILE' WITH MOVE 'ForisaEntApp'        TO '/var/opt/mssql/data/guangji.mdf', MOVE 'ForisaEntApp_Index'    TO '/var/opt/mssql/data/guangji_1.ndf', MOVE 'ForisaEntApp_Master'   TO '/var/opt/mssql/data/guangji_2.ndf', MOVE 'ForisaEntApp_Trx'      TO '/var/opt/mssql/data/guangji_3.ndf', MOVE 'ForisaEntApp_log'      TO '/var/opt/mssql/data/guangji_4.ldf', REPLACE;"

echo "Proses import selesai!"