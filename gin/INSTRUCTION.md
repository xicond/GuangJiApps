# Gobal
- this project is mean to migrate dotnet + mssql 2022 project GuangJiAppsNew to gin gonic + vue3
- database still ue mssql
- target deploy IIS windows
- read from ./backup/db_export_structure.sql for existing database
- create Dockerfile for compilation target windows and test target linux alpine

# Golang (server side)
- do in ./gin folder
- use gin gonic as restful framework
- use go.work, go workspace as microservice to seperate binary each endpoint as different app in IIS to reduce size
- minimize GC pressure di Golang
- use best practice in gin gonic
- optimize for speed and concurrent connection
- Use sync.Pool for Reusable Object
- Utilize Pre-Allocation (make with Capacity)
- Use GOGC and GOMEMLIMIT (Tuning Runtime)
- optimize separation to reduce size
- minimize memory footprint
- minimize latency
- use Concurrent GORM
- optimize query using concurrency with Connection Pool
- always use docker-compose for compilation and runtime
- use ms sql server
- create unit test and functional test
- use go-playground/validator/v10 for form validator
- use joho/godotenv: to define all configuration, and put in docker compose env
- install excelize to read .rdl from ssrs
- do proper validation and error messages
- create docker tools service on demand to generate compiled exe target deploy IIS windows to ./gin/dist

# HTTP
- configurable baseUrl in env
- Auth using bearier: secret token
- JWT_SECRET & REFRESH_SECRET using pem file of openssl
- create GET /ping return {"message":"pong"}
- able to generate and login by refresh token too
- the restful contain basic interface: login, ping as public url

# Menu
Use Vue3 typescript and use ElementFE, with menu:
- Admin Management: Admin, Admin Group, Group Menu Mapping, Admin Sub Warehouse
- Master Data: Umat, Topik, Activity, Tim Kerja, Tahun Ciu Tao, Penggalang Dana, Sxy Donatur
- Transaction: Kelas, Donasi Sxy

# Database
- call existing SP (Stored Procedure) for user login by password
- use legacy password encryption
- if necesary, can. create table for secret token and refresh token

# ORM
- still using ORM
- use lazy load and eager load


## documentation
 - task that done, immediately put in DOCUMENTATION.md ( done only )
 - create postman json, and always update it

## Mapping
 - admin (T_Login_Mst), 
 - Admin Group (T_Login_Group), 
 - Group Menu Mapping (T_Login_Menu), 
 - Admin Sub Warehouse(T_WH_SUBWH_MST)
 
 - umat(T_BUS_UMAT), Topic(T_BUS_TOPIC), Activity(T_BUS_EVENT),
 - Tim Kerja(T_APP_LOOKUP=>B_POSISI),
 - Tahun Ciu Tao(T_BUS_TAHUN_CIUTAO),
 - Penggalang Dana(T_SXY_MST_PENGGALANG),
 - Sxy Donatur(T_SXY_MST_DONATUR)
 
 - Kelas(T_APP_LOOKUP=>B_KELASKHUSUS)
 - Donasi Sxy(T_SXY_TRANSAKSI)
