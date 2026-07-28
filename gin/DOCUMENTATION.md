# Server-side implementation notes

## Completed
- Created a Gin-based server scaffold under the Gin project folder.
- Added a basic router with public endpoints: `/ping`, `/login`, `/forgot-password`, and `/reset-password`.
- Added environment-based configuration loading with `.env.example`.
- Added initial unit tests for the router and a Dockerfile for containerized builds.
- Verified the server starts and `/ping` returns `{"message":"pong"}`.
- Added `ChangePassword` method to `AuthService` and registered `/change-password` and `/v1/change-password` endpoints in Gin router.
- Updated `KelasService` `Create` and `Update` methods to follow the `domain.Kelas` GORM model (`T_TRX_KELAS`) and added lookup validations for `KodeKelas` (`B_KELASKHUSUS`) and `KodeFotang` (`B_FOTHANG`).

## Next recommended steps
- Replace the placeholder authentication service with SQL Server-backed authentication and stored procedure integration.
- Add JWT/refresh-token handling and proper middleware.
- Add structured logging, validation, and more comprehensive API tests.
- Add Docker Compose integration for the Gin service alongside the existing MSSQL container.
