# Server-side implementation notes

## Completed
- Created a Gin-based server scaffold under the Gin project folder.
- Added a basic router with public endpoints: `/ping`, `/login`, `/forgot-password`, and `/reset-password`.
- Added environment-based configuration loading with `.env.example`.
- Added initial unit tests for the router and a Dockerfile for containerized builds.
- Verified the server starts and `/ping` returns `{"message":"pong"}`.
- Added `ChangePassword` method to `AuthService` and registered `/change-password` and `/v1/change-password` endpoints in Gin router.
- Updated `KelasService` `Create` and `Update` methods to follow the `domain.Kelas` GORM model (`T_TRX_KELAS`) and added lookup validations for `KodeKelas` (`B_KELASKHUSUS`) and `KodeFotang` (`B_FOTHANG`).
- Updated `router.go` error handling (`respondError`) to return 404 Not Found for missing records instead of 500.
- Added `GET /v1/kelas/:id/peserta` endpoint with stored procedure `SP_TRX_KELAS_PESERTA_SEARCH_DATA` and pagination (`page`, `limit`).
- Updated Vue frontend `KelasForm`, `KelasEdit`, `KelasView`, and `KelasPesertaTable` with pagination controls and preloaded lookup support.
- Fixed `IntBool` custom type scanning in `legacy_models.go` to support all database column data types (`bool`, `uint8`, `int16`, `int8`, `uint`, `string`, `float`, `[]byte`) and added JSON marshal/unmarshal handling, preventing row scan errors and empty response arrays.
- Updated `KelasService.Peserta` stored procedure invocation to use explicit named parameters (`@PageSize`, `@CurrentPage`, `@SortDirection`, `@TrxId`, `@FotangId`) with numeric `TrxId` parsing for `SP_TRX_KELAS_PESERTA_SEARCH_DATA`.

## Next recommended steps
- Replace the placeholder authentication service with SQL Server-backed authentication and stored procedure integration.
- Add JWT/refresh-token handling and proper middleware.
- Add structured logging, validation, and more comprehensive API tests.
- Add Docker Compose integration for the Gin service alongside the existing MSSQL container.
