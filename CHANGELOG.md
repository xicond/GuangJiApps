# Documentation - Completed Tasks

## \[Completed\] Frontend Dependencies Update: Axios 1.20.0 & Package-Lock Flattening / Deduplication

### Summary & Changes Made:
- **Axios Upgrade**:
  - In [`vue/package.json`](file:///Users/xicond/Workspace/www/GuangJiApps/vue/package.json): Updated `axios` to `1.20.0`.
- **Lockfile Prune & Deduplication**:
  - Executed `npm dedupe` and `npm prune` on [`vue/package-lock.json`](file:///Users/xicond/Workspace/www/GuangJiApps/vue/package-lock.json):
    - Flattened and deduplicated `brace-expansion` across the dependency tree:
      - `minimatch@5.1.9` (under `filelist`) and `minimatch@9.0.9` (under `vue-tsc`) now cleanly share `brace-expansion@2.1.7`.
      - `minimatch@10.2.6` (under `glob@11.1.0` in `workbox-build`) uses `brace-expansion@5.0.12`.
- **Verification**:
  - `npm ls brace-expansion`: All subdependencies resolved with zero conflicts.
  - `npm run type-check` (`vue-tsc --noEmit`): **PASS** (0 errors).
  - `npm run build` (`vue-tsc --noEmit && vite build`): **PASS** (all chunks and PWA service worker generated cleanly).

---

## \[Completed\] Context Cancellation & Unit Test Coverage for Umat and Donasi SXY Report Methods

### Summary & Changes Made:
- **Client Cancellation Handling in Report Services (`.WithContext(ctx)` & `errors.Is(err, context.Canceled)`)**:
  - In [`gin/internal/service/umat_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/umat_service.go) (`UmatService.Report`):
    - Added context cancellation check immediately upon raw SQL row retrieval (`errors.Is(err, context.Canceled)`), returning `Query dibatalkan karena klien memutus koneksi.`.
  - In [`gin/internal/service/donasi_sxy_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/donasi_sxy_service.go) (`DonasiSxyService.Report`):
    - Added context cancellation check on column inspection (`rows.Columns()`), cleanly handling client connection aborts.
- **Unit Test Coverage & Resiliency Updates**:
  - In [`gin/internal/service/umat_service_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/umat_service_test.go):
    - Updated `TestUmatService` to use dynamic ID `idStr := fmt.Sprintf("%d", created.ID)` (preventing hardcoded ID 1 mismatch on MSSQL identity sequences).
    - Added `svc.Report(c, 1, reportFilters, 10)` invocation to test the Report function in the standard workflow.
    - Added dedicated test `TestUmatService_Report_ContextCanceled` validating that a canceled context triggers an error properly.
  - In [`gin/internal/service/donasi_sxy_service_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/donasi_sxy_service_test.go):
    - Added `svc.Report(c, 1, reportFilters, 10)` invocation to `TestDonasiSxyService`.
    - Added dedicated test `TestDonasiSxyService_Report_ContextCanceled` testing context cancellation on `Report`.
- **Verification**:
  - Ran `go test -v -run "TestUmatService|TestDonasiSxyService" ./internal/service/...` with 100% PASS across all 7 test cases.
  - Verified compilation via `docker compose run --rm gin-build` exiting with code 0.

---

## \[Completed\] Context Cancellation & Concurrency Propagation Across All List Services

### Summary & Changes Made:
- **Client Disconnect Cancellation & Context Propagation (`.WithContext(ctx)` & `errors.Is(err, context.Canceled)`)**:
  - Implemented request context extraction with safe fallback across all 23 backend services:
    - Extracted request context using `reqCtx := context.Background()` and `if c != nil && c.Request != nil { reqCtx = c.Request.Context() }`.
    - Applied `.WithContext(reqCtx)` to all Count and Find queries via `query.Session(&gorm.Session{})`.
    - Added immediate client cancellation detection using `if errors.Is(err, context.Canceled) { return nil, 0, fmt.Errorf("Query dibatalkan karena klien memutus koneksi.") }` to abort processing immediately when clients disconnect or cancel requests.
- **Concurrent GORM Execution via `sync.WaitGroup` / Goroutines**:
  - Refactored Count and Find operations to run concurrently in parallel goroutines using isolated sessions (`query.Session(&gorm.Session{})`), dramatically reducing query response latency and database hold time.
- **Services Updated**:
  1. `UmatService.List` ([`gin/internal/service/umat_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/umat_service.go))
  2. `ActivityService.List` ([`gin/internal/service/activity_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/activity_service.go))
  3. `AdminService.List` ([`gin/internal/service/admin_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/admin_service.go))
  4. `AdminGroupService.List` ([`gin/internal/service/admin_group_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/admin_group_service.go))
  5. `AdminSubWarehouseService.List` ([`gin/internal/service/admin_sub_warehouse_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/admin_sub_warehouse_service.go))
  6. `GroupMenuService.List` ([`gin/internal/service/group_menu_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/group_menu_service.go))
  7. `KelasMasterService.List` ([`gin/internal/service/kelas_master_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_master_service.go))
  8. `TimKerjaService.List` ([`gin/internal/service/tim_kerja_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/tim_kerja_service.go))
  9. `TopicService.List` ([`gin/internal/service/topic_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/topic_service.go))
  10. `KelasKendaraanService.List` ([`gin/internal/service/kelas_kendaraan_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_kendaraan_service.go))
  11. `KelasAbsensiService.List` ([`gin/internal/service/kelas_absensi_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_absensi_service.go))
  12. `KelasDonasiBarangService.List` ([`gin/internal/service/kelas_donasi_barang_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_donasi_barang_service.go))
  13. `KelasDonasiService.List` ([`gin/internal/service/kelas_donasi_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_donasi_service.go))
  14. `KelasMusikService.List` ([`gin/internal/service/kelas_musik_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_musik_service.go))
  15. `KelasPengeluaranService.List` ([`gin/internal/service/kelas_pengeluaran_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_pengeluaran_service.go))
  16. `KelasTopikService.List` ([`gin/internal/service/kelas_topik_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_topik_service.go))
  17. `KelasPengabdiService.List` ([`gin/internal/service/kelas_pengabdi_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_pengabdi_service.go))
  18. `KelasService.List` ([`gin/internal/service/kelas_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_service.go))
  19. `KelasPesertaService.List` ([`gin/internal/service/kelas_peserta_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_peserta_service.go))
  20. `TahunCiuTaoService.List` ([`gin/internal/service/tahun_ciu_tao_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/tahun_ciu_tao_service.go))
  21. `PenggalangDanaService.List` ([`gin/internal/service/penggalang_dana_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/penggalang_dana_service.go))
  22. `SxyDonaturService.List` ([`gin/internal/service/sxy_donatur_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/sxy_donatur_service.go))
  23. `DonasiSxyService.List` ([`gin/internal/service/donasi_sxy_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/donasi_sxy_service.go))
- **Routing & Unit Test Updates ([`gin/internal/api/router.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/router.go), `gin/internal/service/*_service_test.go`)**:
  - Updated all route handlers in `router.go` to pass `c *gin.Context` as the first argument to every `List` method call.
  - Updated unit test suites across all 12 test files to pass context `c` to `svc.List(...)`.
- **Verification**:
  - Verified Docker compilation via `docker compose run --rm gin-build` exiting cleanly with code 0.
  - Verified unit test executions.

---

## \[Completed\] Passkey Model Refactor (Credential, Nullable RefreshToken, IP & UserAgent Logging) & POST /login-passkey Endpoint

### Summary & Changes Made:
- **Database Schema & Docker MSSQL Migration ([`backup/db_export_structure.sql`](file:///Users/xicond/Workspace/www/GuangJiApps/backup/db_export_structure.sql), [`docker mssql_server`](file:///Users/xicond/Workspace/www/GuangJiApps))**:
  - Renamed column `PublicKey` to `Credential` (`VARCHAR(MAX) NOT NULL`) in `T_Auth_Passkey`.
  - Altered `RefreshToken` to be nullable (`VARCHAR(MAX) NULL`) in `T_Auth_Passkey`.
  - Added `Ip VARCHAR(45) NULL` to support IPv4 and IPv6 addresses.
  - Added `UserAgent VARCHAR(500) NULL`.
  - Executed migration commands directly on `mssql_server` docker container and updated `AutoMigrate` helper in [`gin/internal/database/db.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/database/db.go).
- **Domain Model Updates ([`gin/internal/domain/legacy_models.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go))**:
  - In `AuthPasskey` struct, renamed `PublicKey` to `Credential string gorm:"column:Credential;not null"`.
  - Updated `RefreshToken` to pointer `*string gorm:"column:RefreshToken"`.
  - Added `IP *string gorm:"column:Ip;type:varchar(45)"` and `UserAgent *string gorm:"column:UserAgent;type:varchar(500)"`.
- **IP & UserAgent Logging & Renewal Validation ([`gin/internal/service/auth_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service.go), [`gin/internal/api/middleware/auth.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/middleware/auth.go))**:
  - Updated `AddPasskey` and `LoginRenew` to accept and persist client IP and User-Agent, with structured logging output (`[Passkey] AddPasskey: ...`, `[Passkey] LoginRenew: ...`).
  - In `RefreshTokenAuthMiddleware` and `LoginRenew`, added validation ensuring `RefreshToken` is not null/empty; if `RefreshToken` is null in DB, renewal is rejected with 401 Unauthorized.
  - In `GetLoginOptions`, filtered out passkeys where `CredentialId IS NULL OR CredentialId = ''`.
  - In `getWebAuthn`, resolved host dynamically from `Referer` header: if `localhost`, set `rpID = host`, `rpOrigin = http://host:env(PORT)` (defaulting to 8080); if not localhost, set `rpID = host`, `rpOrigin = https://host`. Included ceremony origin in allowed RP origins for full multi-origin compatibility.
- **New Rate-Limited Endpoint: POST /login-passkey ([`gin/internal/api/router.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/router.go), [`gin/internal/service/auth_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service.go))**:
  - Added route `POST /login-passkey` protected by rate limiter `loginPasskeyLimiter` (10 requests/minute).
  - Handler receives `{username, credential: {id, response: {clientDataJSON, authenticatorData, signature}}, rememberme}` from `navigator.credentials.get()`.
  - Looks up user from `T_Login_Mst` and matches passkey by `CredentialId`.
  - Validates WebAuthn assertion signature against stored `Credential`.
  - Updates `LastUsed` on matched passkey and `LastLogin` on user.
  - If `rememberme == true`: generates a new refresh token and creates a new active row in `T_Auth_Passkey` (logging IP & UserAgent) without invalidating existing passkey records, returning `token`, `access_token`, `user`, `main_menu`, `refresh_token`, `public_key`, `credential`.
  - If `rememberme == false` (or omitted): returns `token`, `access_token`, `user`, `main_menu` like standard login.
- **Postman Collection Update ([`gin/postman/apps-gin.postman_collection.json`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/postman/apps-gin.postman_collection.json))**:
  - Added request `POST {{baseUrl}}/login-passkey` under Authentication folder.
  - Included 200 OK sample response (with `user`, `main_menu`, `token`, `access_token`, `refresh_token`, `public_key`, `credential`).
  - Included 400 Bad Request sample response for missing/invalid credential or signature.
- **Test Coverage with Direct MSSQL**:
  - `TestRefreshTokenAuthMiddleware` in [`gin/internal/api/middleware/auth_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/middleware/auth_test.go): verified rejection of null refresh tokens and acceptance of active tokens.
  - `TestAuthService_Passkey` in [`gin/internal/service/auth_service_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service_test.go): tested `GetLoginOptions_IgnoreNullCredentialId`, `LoginPasskey_WithoutRememberMe`, and `LoginPasskey_WithRememberMe_IPv6`.
  - `TestPasskey_Acceptance` and `TestPasskey_FullCycleAndTimeTravel` in [`gin/tests/passkey_acceptance_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/passkey_acceptance_test.go): 100% PASSED directly on MSSQL.

---

## \[Completed\] Passkey Schema Refactor (LoginId & CredentialId) & POST /login-options Endpoint

### Summary & Changes Made:
- **Database Structure & Domain Model Refactor ([`backup/db_export_structure.sql`](file:///Users/xicond/Workspace/www/GuangJiApps/backup/db_export_structure.sql), [`gin/internal/domain/legacy_models.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go))**:
  - Renamed `UserId` to `LoginId` (`INT NOT NULL`) in `T_Auth_Passkey` and added non-clustered index `IX_T_Auth_Passkey_LoginId`.
  - Added column `CredentialId VARCHAR(255) NULL` with non-clustered index `IX_T_Auth_Passkey_CredentialId` on `T_Auth_Passkey`.
  - In `AuthPasskey` struct, updated fields: `LoginID int32` mapped to `LoginId`, `CredentialID string` mapped to `CredentialId`.
  - Removed deprecated `type:text` tag from `PublicKey` and `RefreshToken` strings, ensuring GORM cross-compatibility (`VARCHAR(MAX)` on MSSQL without `=` operator incompatibility, and `text` on SQLite without syntax errors).
  - Updated auto-migration checks in [`gin/internal/database/db.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/database/db.go) to safely rename `UserId` to `LoginId` and add `CredentialId` on MSSQL instances.
- **Passkey Persistence & CredentialId Storage ([`gin/internal/service/auth_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service.go))**:
  - In `AddPasskey`, saved WebAuthn `Credential.ID` into `CredentialId` as unpadded Base64URL (`base64.RawURLEncoding.EncodeToString(credential.ID)`).
  - In `AddPasskey`, mapped `LoginID: userID` and captured authenticator transports from attestation response.
  - In `LoginRenew`, preserved `CredentialID: passkey.CredentialID` and `LoginID: passkey.LoginID` across refresh token rotations.
  - In `RefreshTokenAuthMiddleware` ([`gin/internal/api/middleware/auth.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/middleware/auth.go)), set both `loginID` and `userID` in the context using `passkey.LoginID`.
- **New Rate-Limited Endpoint: POST /login-options ([`gin/internal/api/router.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/router.go), [`gin/internal/service/auth_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service.go))**:
  - Implemented `GetLoginOptions(username, origin string) (*LoginOptionsResponse, error)`:
    - Queries user from `T_Login_Mst` by `Username`.
    - Queries all active passkeys (`Status = 1 OR Status IS NULL`) for user's `LoginId`, supporting 1-to-many relationship (`T_Login_Mst` to `T_Auth_Passkey`).
    - Generates 32-byte cryptographically secure WebAuthn challenge via `protocol.CreateChallenge()`.
    - Dynamically derives `rpId` from `Origin` header (or fallback host / domain).
    - Populates `allowCredentials` list with each credential's `type: "public-key"`, Base64URL `id`, and transports (defaulting to `["hybrid", "internal"]`).
    - Sets `userVerification: "required"`.
  - Added route `POST /login-options` guarded by rate limiter `loginOptionsLimiter` (10 requests/minute).
- **Postman Collection Update ([`gin/postman/apps-gin.postman_collection.json`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/postman/apps-gin.postman_collection.json))**:
  - Added request `POST {{baseUrl}}/login-options` under the `Authentication` folder with payload `{"username": "admin"}`.
  - Added 200 OK sample response demonstrating multiple allowCredentials and WebAuthn options.
  - Added 400 Bad Request sample response for missing/empty username.
- **Verification & Test Coverage**:
  - Unit tests in [`gin/internal/service/auth_service_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service_test.go): `GetLoginOptions_SinglePasskey`, `GetLoginOptions_MultiplePasskeys_1ToMany`, `GetLoginOptions_UserNotFound`, `GetLoginOptions_EmptyUsername` (all passed 100%).
  - Live MSSQL Acceptance & E2E tests in [`gin/tests/passkey_acceptance_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/passkey_acceptance_test.go): `TestPasskey_Acceptance` (including single & multiple credentials assertion options) and `TestPasskey_FullCycleAndTimeTravel` (all passed 100%).

---

## \[Completed\] Fix "refresh key not found" on Login Renew & Race Condition in Axios Client

### Summary & Changes Made:
- **Root Cause Resolution for "refresh key not found" / "refresh token not found or revoked"**:
  - Identified scan error in `RefreshTokenAuthMiddleware`: MSSQL column `[Status]` in `T_Auth_Passkey` is defined as `BIT` (returning driver `bool`). In [`gin/internal/domain/legacy_models.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), `Status` was typed `*int`, causing `sql: Scan error on column index 7, name "Status": converting driver.Value type bool ("true") to a int: invalid syntax`. GORM returned this scan error during `First(&passkey)`, causing middleware to mistakenly reject valid tokens with `401 {"error": "refresh token not found or revoked"}`.
  - Updated `AuthPasskey.Status` to `*bool` with `type:bit` across models, services, and test suites.
  - Added missing `c.Set("passkey", passkey)` in `RefreshTokenAuthMiddleware` so downstream `/login-renew` route handler receives the validated passkey.
- **VARCHAR(MAX) & MD5 Index Acceleration**:
  - Reverted `PublicKey` and `RefreshToken` columns in `T_Auth_Passkey` to `VARCHAR(MAX)`.
  - Added `RefreshTokenHash VARCHAR(32) NULL` column with non-clustered index `IX_T_Auth_Passkey_RefreshTokenHash` in MSSQL and [`backup/db_export_structure.sql`](file:///Users/xicond/Workspace/www/GuangJiApps/backup/db_export_structure.sql).
  - Updated [`gin/internal/service/auth_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service.go) to compute and store MD5 hex hash of refresh tokens.
  - Updated [`gin/internal/api/middleware/auth.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/middleware/auth.go) to utilize indexed `(RefreshTokenHash = ? OR RefreshToken = ?)` for O(1) B-Tree lookup speed while preserving backward compatibility.
- **Full Lifecycle Acceptance & E2E Testing with Live MSSQL**:
  - Tested directly against MSSQL (`admin`/`password`).
  - Added [`TestPasskey_FullCycleAndTimeTravel`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/passkey_acceptance_test.go) asserting the entire 9-step flow:
    1. Login credentials (`admin`/`password`).
    2. Protected resource access (`/v1/umats`) with login token.
    3. Biometric passkey registration with Remember Me flow.
    4. Time-travel access token expiration (401 invalid token).
    5. Biometric session recovery via `POST /login-renew` returning renewed access and refresh tokens.
    6. Access restored: protected endpoint succeeds with renewed access token.
    7. Old refresh token replay rejection (401).
    8. Time-travel refresh token expiration (401).
    9. Continuous chain rotation renewal (200 OK).
  - Verified 100% pass on all acceptance test suites.
- **Race Condition & Empty Bearer Elimination ([`vue/src/api/client.ts`](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/api/client.ts))**:
  - Request interceptor: Awaits in-flight `renewPromise` before dispatching so outgoing requests use the fresh token instead of sending the known-expired token.
  - Request interceptor: Trims token and removes `Authorization` header (`delete config.headers.Authorization`) if token is empty, guaranteeing no empty `Bearer ` is ever sent.
  - Response interceptor: Parallel 401 requests during user biometric prompt join and wait on `renewPromise` rather than triggering redundant prompts or cascading into premature logout.
  - Response interceptor: Fallback logout is strictly guarded by `if (!isRenewing && !renewPromise)`, preventing concurrent 401s from clearing session data or redirecting during an active renewal.
  - Verified with `npm run build` (`vue-tsc --noEmit && vite build`), building cleanly with 0 errors.

---

## \[Completed\] Biometric Passkey Remember Me & Auto-Renewal on Vue Frontend

### Summary & Changes Made:
- **Passkey Utility Service ([`vue/src/services/passkey.ts`](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/services/passkey.ts))**:
  - Implemented `bufferToBase64URL` for RFC 4648 §5 WebAuthn payload encoding without third-party dependencies.
  - Implemented `isPlatformAuthenticatorAvailable` evaluating WebAuthn support and platform authenticators (FaceID/TouchID/Windows Hello).
  - Implemented `registerBiometricPasskey` using `navigator.credentials.create` with ES256 (`alg: -7`), challenge as Uint8Array of the access token, and sending attestation to `POST /v1/add-passkey`.
  - Implemented `renewBiometricSession` using `navigator.credentials.get` with challenge as Uint8Array of the refresh token, `userVerification: 'required'`, and sending assertion to `POST /login-renew`.
- **Login Component ([`vue/src/views/Login.vue`](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/views/Login.vue))**:
  - Added checkbox "Remember Me (Biometrik)" conditioned on platform authenticator availability (`(!window.PublicKeyCredential || !PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable) == true` or browser platform authenticator check).
  - On successful login with remember me checked, invokes `registerBiometricPasskey` and saves the received `refresh_token` in `localStorage`.
  - Catches passkey registration errors non-fatally with `ElMessage.warning` without forcing logout.
- **Axios HTTP Client Interceptor ([`vue/src/api/client.ts`](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/api/client.ts))**:
  - Intercepts HTTP 401 Unauthorized responses.
  - If a saved `refresh_token` exists and the request was not to an auth endpoint, triggers biometric session renewal via `renewBiometricSession`.
  - Implemented concurrency queue so simultaneous 401 requests share a single in-flight renewal without premature logout if the biometric process takes longer than 3 seconds.
  - Checks if token has already been renewed by another request (`currentActiveToken !== requestToken`) and retries immediately with the new token.
  - Applied a 3-second throttle (`RENEW_THROTTLE_MS = 3000`) to prevent excessive calls to `/login-renew`, only logging out if the biometric verification actually fails or is canceled.
  - Automatically updates `authStore.token` and `localStorage` with renewed access token and refresh token, then retries the original failed request.
  - If renewal fails or is canceled by user, performs logout and redirects to `/login`.
- **Auth Store ([`vue/src/stores/auth.ts`](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/stores/auth.ts))**:
  - Updated `logout()` to remove `refresh_token` from `localStorage` in addition to `token`, `user`, and `main_menu`.
- **Verification**:
  - Validated with `npm run build` (`vue-tsc --noEmit && vite build`), compiling cleanly with 0 TypeScript/build errors.

---

## \[Completed\] Add Renewed Access Token to `/login-renew` Endpoint

### Summary & Changes Made:
- **Service Layer ([`gin/internal/service/auth_service.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service.go))**:
  - Updated `AuthService.LoginRenew` signature and logic to generate and return both `newRefreshToken` and `newAccessToken` (via `s.GenerateToken(passkey.UserID)`).
- **HTTP Routing ([`gin/internal/api/router.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/router.go))**:
  - Updated `POST /login-renew` handler to include `"token"` and `"access_token"` in the 200 OK JSON response:
    ```json
    {
      "token": "...",
      "access_token": "...",
      "refresh_token": "..."
    }
    ```
- **Postman Collection ([`gin/postman/apps-gin.postman_collection.json`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/postman/apps-gin.postman_collection.json))**:
  - Updated `Login Renew` sample 200 response to `200 OK - Access and Refresh Token Renewed` with realistic `token`, `access_token`, and `refresh_token` payload.
- **Tests & Verification**:
  - Updated [`gin/internal/service/auth_service_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service_test.go) asserting non-empty renewed access token.
  - Updated [`gin/tests/passkey_acceptance_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/passkey_acceptance_test.go) and [`gin/tests/e2e/passkey_e2e_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/passkey_e2e_test.go) validating presence of `token` and `access_token` in response body.

---

## \[Completed\] Add Sample 200 and 400 Responses in Postman Collection

### Summary & Changes Made:
- **Postman Collection Update ([`gin/postman/apps-gin.postman_collection.json`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/postman/apps-gin.postman_collection.json))**:
  - Added sample responses to **`Add Passkey`** (`POST {{baseUrl}}/v1/add-passkey`):
    - `200 OK - Passkey Added Successfully`: Returns `{"message": "passkey added successfully", "refresh_token": "..."}`.
    - `400 Bad Request - Challenge Mismatch`: Returns `{"error": "challenge mismatch: provided challenge does not match expected authentication token"}`.
  - Added sample responses to **`Login Renew`** (`POST {{baseUrl}}/login-renew`):
    - `200 OK - Refresh Token Renewed`: Returns `{"refresh_token": "..."}`.
    - `400 Bad Request - Invalid Payload`: Returns `{"error": "Format payload JSON tidak valid atau rusak.", "details": {}}`.
  - Validated Postman Collection JSON schema syntax and structure via `python3 json.load`.

---

## \[Completed\] Refactor Passkey Authentication to `github.com/go-webauthn/webauthn/webauthn`

### Summary & Changes Made:
- **Library Integration & Removal of Custom Implementation**:
  - Replaced handwritten WebAuthn/COSE/CBOR parsing in [`gin/internal/service/passkey.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/passkey.go) and scratch tests with the official, maintained [`github.com/go-webauthn/webauthn/webauthn`](https://github.com/go-webauthn/webauthn) package (v0.18.2).
  - Deleted obsolete files: `gin/internal/service/passkey.go`, `gin/internal/service/passkey_test.go`, and `gin/internal/service/webauthn_test.go`.
- **Service Layer Refactoring (`gin/internal/service/auth_service.go`)**:
  - Implemented [`PasskeyUser`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service.go) adapting administrative users to the `webauthn.User` interface (`WebAuthnID`, `WebAuthnName`, `WebAuthnDisplayName`, `WebAuthnCredentials`).
  - Added concurrent `webauthnCache` (`sync.Map`) on `AuthService` caching Relying Party `webauthn.WebAuthn` instances per origin/domain to minimize GC pressure and eliminate allocation overhead during ceremonies.
  - Refactored `AddPasskey`:
    - Validates WebAuthn creation ceremony type and challenge against the caller's JWT Bearer token and refresh tokens.
    - Parses attestation responses using `protocol.ParseCredentialCreationResponseBytes` and performs full standard verification via `w.CreateCredential`.
    - Supports synthesis of standard `none` attestationObject CBOR from direct public key PEM/DER for testing and direct browser clients.
    - Persists the serialized `webauthn.Credential` record as JSON in `T_Auth_Passkey.PublicKey`.
  - Refactored `LoginRenew`:
    - Validates assertion ceremony type and challenge against the active refresh token.
    - Deserializes `webauthn.Credential` from `passkey.PublicKey` (with backward-compatible decoding fallback for legacy PEM keys).
    - Verifies biometric assertions using `w.ValidateLogin(user, loginSession, parsedAssertion)`.
    - Revokes the consumed refresh token (`Status = 0`, `LastUsed = now`), generates and persists a new 30-day refresh token with the identical credential, and returns only the new refresh token (`{"refresh_token": "..."}`).
- **Database Model Compatibility (`gin/internal/domain/legacy_models.go`)**:
  - Updated GORM column tags for `PublicKey` and `RefreshToken` on [`AuthPasskey`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go) to `type:text;not null`, ensuring multi-database compatibility across both SQLite test environments and MSSQL production environments.
- **Test Suite Updates & Verification**:
  - Updated [`gin/internal/service/auth_service_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service_test.go), [`gin/tests/passkey_acceptance_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/passkey_acceptance_test.go), and [`gin/tests/e2e/passkey_e2e_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/passkey_e2e_test.go) with standard WebAuthn authenticator data formatting (including Relying Party ID SHA-256 hashes).
  - Executed `go test -count=1 -v ./internal/service -run TestAuthService_Passkey` (**PASS** in 0.00s).
  - Executed `go test -count=1 -v ./tests -run TestPasskey_Acceptance` (**PASS** all 9 subtests in 0.03s).

---

## \[Completed\] Biometric Passkey ("Remember Me") Phase 1 Implementation

### Summary & Changes Made:
- **Database Schema**:
  - Added [`T_Auth_Passkey`](file:///Users/xicond/Workspace/www/GuangJiApps/backup/db_export_structure.sql) in [`backup/db_export_structure.sql`](file:///Users/xicond/Workspace/www/GuangJiApps/backup/db_export_structure.sql) with columns: `PasskeyId` (IDENTITY primary key), `UserId`, `PublicKey` (VARCHAR(MAX)), `RefreshToken` (VARCHAR(MAX)), `CreatedAt`, `LastUsed`, `ExpiredAt`, and `Status` (BIT DEFAULT 1 for soft-delete).
  - Defined [`AuthPasskey`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go) domain model and mapped to `T_Auth_Passkey`.
  - Registered `&domain.AuthPasskey{}` in [`database.AutoMigrate`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/database/db.go).
- **WebAuthn / Passkey Cryptographic Support**:
  - Created [`gin/internal/service/passkey.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/passkey.go) supporting:
    - Base64/Base64URL decoding with padding tolerance.
    - Public key extraction from direct PEM, base64 SPKI DER, or CBOR `attestationObject` (parsing `authData` and COSE key ES256 / RS256).
    - ClientData JSON challenge validation against Bearer token and Refresh tokens.
    - W3C WebAuthn assertion signature verification (`authenticatorData || sha256(clientDataJSON)`) against stored public key PEM using standard Go crypto (`crypto/ecdsa`, `crypto/rsa`, `crypto/x509`).
    - Allocation minimization using `sync.Pool` for byte buffers.
- **Service Layer**:
  - Added [`GenerateRefreshToken`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service.go) to `AuthService` issuing 30-day signed JWT tokens.
  - Added [`AddPasskey`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service.go) to `AuthService` validating registration attestation, verifying challenge, decoding public key, ensuring unique refresh token, persisting in `T_Auth_Passkey`, and returning initial refresh token.
  - Added [`LoginRenew`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service.go) to `AuthService` validating assertion signature, invalidating previous refresh token (`Status = 0`), issuing and storing a new refresh token with the same public key, and returning only the new refresh token.
- **Middleware Layer**:
  - Implemented [`RefreshTokenAuthMiddleware`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/middleware/auth.go) in `gin/internal/api/middleware/auth.go` validating refresh token presence, JWT signature & expiration claims, and active status (`Status = 1`) in `T_Auth_Passkey`.
- **Router Layer**:
  - Added `POST /v1/add-passkey` under the authenticated `/v1` group in [`gin/internal/api/router.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/router.go).
  - Added `POST /login-renew` with dedicated user rate limiting (`login_renew_limiter:`) and `RefreshTokenAuthMiddleware`.
- **Postman Collection & Verification**:
  - Updated [`gin/postman/apps-gin.postman_collection.json`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/postman/apps-gin.postman_collection.json) with `Add Passkey` and `Login Renew` endpoints under the Auth group.
  - Authored comprehensive unit tests in [`gin/internal/service/passkey_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/passkey_test.go), [`gin/internal/service/auth_service_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/auth_service_test.go), and [`gin/internal/api/middleware/auth_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/middleware/auth_test.go).
  - Created acceptance test suite [`gin/tests/passkey_acceptance_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/passkey_acceptance_test.go) and end-to-end integration test [`gin/tests/e2e/passkey_e2e_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/passkey_e2e_test.go) verifying registration, cryptographic biometric renewal assertion, rate limiting, and old refresh token replay prevention.

---

### Summary & Changes Made:
- **Exclusion of `*.py` Files**:
  - Configured [`.codacy.yml`](file:///Users/xicond/Workspace/www/GuangJiApps/.codacy.yml) and [`.codacy.yaml`](file:///Users/xicond/Workspace/www/GuangJiApps/.codacy.yaml) with explicit file and glob patterns (`*.py`, `**/*.py`, `backup/**`, `gin/postman/**`, `gin/*.py`) under both root `exclude_paths` and `languages.python.exclude_paths`.
  - Disabled Python analysis engines (`bandit`, `prospector`, `pylint`, `flake8`, `ruff`) and excluded Python files from `semgrep`, `duplication`, and `revive`.
  - Added temporary stashing & immediate restoration of `*.py` files around `Run Codacy Analysis CLI` in [`.github/workflows/codacy.yml`](file:///Users/xicond/Workspace/www/GuangJiApps/.github/workflows/codacy.yml) so Codacy CLI physically cannot discover or scan Python files regardless of cloud UI settings.
- **Disabling `Revive_var-naming` and `expr` Rules**:
  - Updated [`revive.toml`](file:///Users/xicond/Workspace/www/GuangJiApps/revive.toml) with `[rule.var-naming] disabled = true` and `[rule.bool-literal-in-expr] disabled = true`.
- **Workflow-Level SARIF Sanitization**:
  - Updated [`.github/workflows/codacy.yml`](file:///Users/xicond/Workspace/www/GuangJiApps/.github/workflows/codacy.yml) with permission fix (`sudo chown`) and filtering step before uploading to GitHub Security, stripping any alerts for files ending in `.py`, rules matching `Revive_var-naming` / `var-naming`, or rules matching `expr` (`expr`, `Revive_expr`, `bool-literal-in-expr`, etc.).

---

## \[Completed\] Configure OSV-Scanner False Positive Suppression (`GO-2026-5932`)

### Summary & Changes Made:
- **Suppression of Unused Subpackage Advisory**:
  - Addressed OSV vulnerability warning `GO-2026-5932` on `golang.org/x/crypto/openpgp`.
  - Confirmed via `go list -deps ./...` that `golang.org/x/crypto/openpgp` is never imported, called, or reachable in the codebase.
  - Configured [`osv-scanner.toml`](file:///Users/xicond/Workspace/www/GuangJiApps/osv-scanner.toml) at repository root and in [`gin/osv-scanner.toml`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/osv-scanner.toml) with `[[IgnoredVulns]]` to prevent CI scanner failures.

---

## \[Completed\] Function and Method Docblock Comments in `gin/internal/domain/legacy_models.go`

### Summary & Changes Made:
- **Comprehensive Method & Function Docblock Comments**:
  - Authored standard Go doc comments (`// MethodName ...`, with detailed `Parameters:` and `Returns:` sections) across all methods and functions in [`gin/internal/domain/legacy_models.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go):
    - Table name resolvers (`TableName() string`) implementing GORM `schema.Tabler`: [`Admin`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L154), [`AdminMatrix`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L165), [`AdminGroup`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L200), [`DepartmentMst`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L214), [`GroupMenuMapping`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L231), [`AdminSubWarehouse`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L248), [`Umat`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L404), [`UmatFoto`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L421), [`AppLookup`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L440), [`AppLookupCategory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L457), [`Topic`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L475), [`Activity`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L491), [`TimKerja`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L505), [`TahunCiuTao`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1004), [`PenggalangDana`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1040), [`SxyDonatur`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1098), [`Kelas`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1126), [`KelasPeserta`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1170), [`KelasPengabdi`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1473), [`KelasTopik`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1502), [`KelasKendaraan`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1525), [`KelasDonasi`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1543), [`KelasDonasiBarang`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1561), [`KelasPengeluaran`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1580), [`KelasMusik`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1600), [`KelasAbsensi`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1618), [`DonasiSxy`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1644), [`WorkMapping`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L1674).
    - JSON unmarshaling and string parsing methods: [`(u *Umat) UnmarshalJSON`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L367), [`(d *DateOnly) parseString`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L532), [`(dt *DateTime) parseString`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L810).
    - Inner/nested auxiliary structs: documented `aux` within [`(u *Umat) UnmarshalJSON`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go#L367).
- **Validation**:
  - Ran `go vet ./...` across the entire codebase with 0 errors.

---

## \[Completed\] Struct and Function Docblock Comments in E2E Testing Suite (`gin/tests/e2e/`)

### Summary & Changes Made:
- **Comprehensive E2E Factory Docblock Comments**:
  - Authored standard Go doc comments (`// StructName ...`, `// MethodName ...`) across all test factory utilities and data payload generators in [`gin/tests/e2e/factory/`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory):
    - Random generation helpers ([`factory.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/factory.go)): [`RandString`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/factory.go#L18), [`RandDigits`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/factory.go#L33), [`RandCode`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/factory.go#L46), [`RandPhone`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/factory.go#L54), [`RandEmail`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/factory.go#L64), [`UniqueSuffix`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/factory.go#L72).
    - Admin management factories ([`admin_factory.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/admin_factory.go)): [`AdminFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/admin_factory.go#L7), [`AdminGroupFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/admin_factory.go#L59), [`GroupMenuMappingFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/admin_factory.go#L99), [`AdminSubWarehouseFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/admin_factory.go#L141) with all `ValidPayload`, `InvalidPayload`, `ValidUpdatePayload`, and `InvalidUpdatePayload` generator methods.
    - Master data factories ([`master_factory.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/master_factory.go)): [`UmatFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/master_factory.go#L5), [`TopicFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/master_factory.go#L49), [`KelasMasterFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/master_factory.go#L93), [`ActivityFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/master_factory.go#L154), [`TimKerjaFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/master_factory.go#L198), [`TahunCiuTaoFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/master_factory.go#L242), [`PenggalangDanaFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/master_factory.go#L286), [`SxyDonaturFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/master_factory.go#L330).
    - Transaction factories ([`transaction_factory.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/transaction_factory.go)): [`KelasFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/transaction_factory.go#L5), [`KelasPesertaFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/transaction_factory.go#L47), [`KelasPengabdiFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/transaction_factory.go#L91), [`KelasTopikFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/transaction_factory.go#L125), [`DonasiSxyFactory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/factory/transaction_factory.go#L153).
- **Comprehensive E2E Test Suite Docblock Comments**:
  - Authored standard Go doc comments for all test fixtures and suites in [`gin/tests/e2e/`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e):
    - Fixture configuration ([`setup_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/setup_test.go)): [`getBaseURL`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/setup_test.go#L22), [`newExpect`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/setup_test.go#L34), [`getAuthToken`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/setup_test.go#L49), [`newAuthExpect`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/setup_test.go#L81), [`getAuthenticatedExpect`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/setup_test.go#L93).
    - Authentication suite ([`auth_e2e_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/auth_e2e_test.go)): [`TestE2E_Ping`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/auth_e2e_test.go#L12), [`TestE2E_Login`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/auth_e2e_test.go#L25), [`TestE2E_ChangePassword`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/auth_e2e_test.go#L75).
    - Lookups suite ([`lookup_e2e_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/lookup_e2e_test.go)): [`TestE2E_LookupService`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/lookup_e2e_test.go#L14).
    - Admin management suite ([`admin_mgmt_e2e_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/admin_mgmt_e2e_test.go)): [`TestE2E_AdminManagement`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/admin_mgmt_e2e_test.go#L17).
    - Master data suite ([`master_data_e2e_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/master_data_e2e_test.go)): [`TestE2E_MasterData`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/master_data_e2e_test.go#L17).
    - Transactions suite ([`transactions_e2e_test.go`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/transactions_e2e_test.go)): [`TestKelas_EndToEnd`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/transactions_e2e_test.go#L20), [`TestDonasiSxy_EndToEnd`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/tests/e2e/transactions_e2e_test.go#L424).
- **Validation**:
  - Ran `go vet ./...` and `go test -run=^$ ./tests/e2e/...` with 0 errors and clean compilation.

---

## \[Completed\] Struct and Function Docblock Comments in Golang Backend (`gin/`)

### Summary & Changes Made:
- **Comprehensive Struct Docblock Comments**:
  - Authored standard Go doc comments (`// StructName ...`) for all domain entities, request payloads, response DTOs, and reporting models in [`gin/internal/domain/legacy_models.go`](./gin/internal/domain/legacy_models.go), including:
    - Administrative & authorization structures: [`Admin`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`AdminMatrix`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`SubMenuItem`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`MainMenuItem`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`AdminGroup`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`DepartmentMst`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`GroupMenuMapping`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`AdminSubWarehouse`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go).
    - Congregant & membership structures: [`Umat`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`VerifyQRRequest`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`VerifyQRResponse`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`UmatFoto`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`UmatReport`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`UmatPopUpResponse`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go).
    - Master data & lookups: [`AppLookup`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`AppLookupCategory`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`Topic`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`Activity`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`TimKerja`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`TahunCiuTao`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`WorkMapping`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go).
    - Class management & event logistics: [`Kelas`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasResponse`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasPeserta`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasPesertaBulkRequest`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasPesertaList`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasPesertaResponse`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasPesertaPrevious`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasPengabdi`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasTopik`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasKendaraan`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasDonasi`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasDonasiBarang`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasPengeluaran`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasMusik`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`KelasAbsensi`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go).
    - Charity transactions: [`PenggalangDana`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`PenggalangDanaResponse`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`SxyDonatur`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`SxyDonaturResponse`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`SxyDonasiReport`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`DonasiSxy`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`DonasiSxyResponse`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go).
    - Custom scalar types: [`DateOnly`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`DateTime`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`FlexString`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go), [`IntBool`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/domain/legacy_models.go).
    - Query filter rules & service models: [`FilterRule`](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/kelas_master_service.go).
- **Validation**:
  - Executed `go vet ./...` and `go build ./...` across all packages in `gin/` with 0 errors and 0 warnings.

---

## \[Completed\] Function Docstrings in Postman UAT Generator (`generate_uat.py`)

### Summary & Changes Made:
- **Standardized Docstrings Across All Functions**:
  - Authored comprehensive Google-style / PEP 257 docstrings for all functions in [`gin/postman/generate_uat.py`](./gin/postman/generate_uat.py):
    - `build_collection()`: Overall collection processing lifecycle.
    - `process_items()`: Recursive traversal of collection items/subfolders with `Args`.
    - `make_test()`: Postman test event script structure generation with `Args` and `Returns`.
    - `make_sample_response()`: Postman mock example response structure generation with `Args` and `Returns`.
    - `configure_item()`: Endpoint-level test script and mock response attachment with `Args`.
    - `handle_crud()`: Standard CRUD test assertion generator with `Args`.
    - `make_get_list_test()`, `make_get_detail_test()`, `make_get_lookup_test()`, `make_get_excel_test()`: Dedicated GET assertion builders with `Args` and `Returns`.
    - `make_post_create_test()`, `make_patch_update_test()`, `make_delete_test()`: Mutation test assertion builders with `Args` and `Returns`.
- **Removed Unused Parameter**:
  - Removed unused `item` argument from [`handle_crud`](./gin/postman/generate_uat.py) definition and all 15 invocation sites, updating the docstring accordingly.
- **Validation**:
  - Verified syntax with `py_compile` to guarantee zero syntax or linting regressions.

---

## \[Completed\] Custom Proprietary & Restricted License (`LICENSE.md`)

### Summary & Changes Made:
- **Trilingual Custom Proprietary License**:
  - Authored formal license agreements across 3 dedicated files:
    - **English (en)**: [`LICENSE.md`](./LICENSE.md)
    - **Bahasa Indonesia (id)**: [`LICENSE.id.md`](./LICENSE.id.md)
    - **简体中文 (zh-Hans)**: [`LICENSE.zh.md`](./LICENSE.zh.md)
  - All editions feature cross-navigation language headers and standardized legal clauses.
- **Permitted Use for Authorized Non-Profit Foundations**:
  - Explicitly restricts application deployment and operation solely to authorized non-profit foundations (Yayasan) that have obtained express written permission from the Developer.
  - Prohibits any commercial, for-profit, or unauthorized third-party entity usage.
- **Testing & Quality Assurance Allowance**:
  - Explicitly permits functional testing, integration testing, E2E validation, load testing (e.g. k6), and security audits within authorized testing environments.
- **Strict Prohibitions**:
  - **No Repository Duplication**: Strictly prohibits cloning, copying, mirroring, public re-uploading, or redistributing the repository or source code in any form.
  - **Mandatory Developer Consent**: Mandates prior written authorization from the Developer for any other use, modification outside the foundation, or distribution.
- **TOC & README Integration**: Linked all 3 license files in [`README.md`](./README.md) Table of Contents and License section, and added the official GuangJiApps logo ([`favicon.png`](./vue/public/favicon.png)) centered at the very top of [`README.md`](./README.md).

---

## \[Completed\] Project README Documentation (`README.md`)

### Summary & Changes Made:
- **Comprehensive Project Overview**: Created top-level [`README.md`](./README.md) introducing the GuangJiApps enterprise platform, explaining the migration from the legacy .NET Framework / SSRS monolith (`GuangJiAppsNew`) to Gin-Gonic (Go) and Vue 3 + TypeScript.
- **Architecture & Tech Stack Breakdown**: Documented technology choices across backend (Go 1.25+, Gin-Gonic, GORM, Redis, Validator v10, Excelize), frontend (Vue 3, TypeScript 5, Element Plus, Pinia, Vite, Workbox PWA), database (MSSQL Server 2022, Redis), and quality assurance tools (k6, Semgrep, Postman).
- **Directory Structure & Component Mapping**: Outlined repository tree mapping `gin/`, `vue/`, `backup/`, and Docker services.
- **Step-by-Step Quickstart**: Provided local running instructions via Docker Compose (`docker compose up -d`), independent host commands for Go backend and Vue frontend, environment configurations (`.env`), and Windows IIS cross-compilation (`docker compose --profile build run --rm gin-build`).
- **Coding Style & Development Standards**: Embedded core engineering guidelines covering GC minimization (`sync.Pool`, slice pre-allocation, runtime tuning with `GOGC`/`GOMEMLIMIT`), concurrent GORM connection pooling, form submission disabling with `finally` blocks, strict TypeScript usage, Postman collection synchronization, and documentation rules.

---

## \[Completed\] Postman Documentation Re-arrangement: API v1 Hierarchy, CRUD Sub-folders, and Sample Environment

### Summary & Changes Made:
- **Collection Version (v1) & Collection Headers**:
  - Updated collection metadata in [`apps-gin.postman_collection.json`](./gin/postman/apps-gin.postman_collection.json) to version `"v1"`.
  - Added `apiVersion: "v1"` to collection variables alongside `baseUrl`, `authToken`, and `resourceId`.
  - Reverted outer `v1` folder to keep top-level navigation direct (`Public`, `Admin Management`, `Master Data`, `Lookup Service`, `Transactions`).
  - Added collection-level header definition for `Content-Type: application/json` and `Accept: application/json` (both via collection root `header` definition and collection pre-request script handling automatic injection while preserving multipart/formdata).
- **Automated Authentication Lifecycle (Pre-request & 401 Retry Test Scripts)**:
  - **Pre-request Script (`prerequest`)**:
    - Automatically checks if `authToken` is present in environment or collection variables.
    - If `authToken` already exists, skips login and attaches `Bearer <token>` to request headers.
    - If `authToken` is empty or missing (and current request is not `Login` or `Ping`), automatically sends an asynchronous POST to `{{baseUrl}}/login` using default environment/collection credentials (`root` / `password123`), stores the received token in both environment and collection variables, and attaches the `Authorization` header to the outgoing request.
  - **Post-response Script (`test`)**:
    - Listens for HTTP 401 Unauthorized responses on protected requests.
    - Automatically executes a re-login flow to acquire a fresh JWT token, updates `authToken` in environment and collection variables, and retries the original request with the renewed Bearer token.
- **Nested Folder Classification for CRUD Operations**:
  - **Admin Management**: Grouped 21 requests into 5 dedicated entity folders:
    - `Admins` (List, Create, Get By ID, Update, Delete)
    - `Department` (List)
    - `Admin Groups` (List, Create, Get By ID, Update, Delete)
    - `Group Menu Mappings` (List, Create, Get By ID, Update, Delete)
    - `Admin Sub Warehouses` (List, Create, Get By ID, Update, Delete)
  - **Master Data**: Grouped 52 requests into 9 dedicated entity folders:
    - `Umat` (12 endpoints: Report, Report Excel, OCR, List, PopUp, Create, Create Multipart, Get By ID, Verify QR, Update, Update Multipart, Delete)
    - `Topics` (List, Create, Get By ID, Update, Delete)
    - `Kelas Master` (List, Create, Get By ID, Update, Delete)
    - `Activities` (List, Create, Get By ID, Update, Delete)
    - `Tim Kerja` (List, Create, Get By ID, Update, Delete, Lookup, Lookup Sub, Lookup Report)
    - `Tahun Ciu Tao` (List, Create, Get By ID, Update, Delete)
    - `Penggalang Dana` (List, Create, Get By ID, Update, Delete)
    - `Sxy Donatur` (List, Create, Get By ID, Update, Delete)
    - `Fotang` (Lookup, Lookup Sxy)
  - **Transactions**: Grouped 32 requests into 12 dedicated entity folders:
    - `Kelas` (List, Lookup, Report, Peserta List, Peserta Load Previous, Create, Get By ID, Update, Delete)
    - `Kelas Peserta` (Get, Get By IdPeserta, Create, Create Bulk, Update, Delete)
    - `Kelas Pengabdi` (List, Create)
    - `Kelas Topik` (List, Create)
    - `Kelas Kendaraan` (List)
    - `Kelas Donasi` (List)
    - `Kelas Donasi Barang` (List)
    - `Kelas Pengeluaran` (List)
    - `Kelas Musik` (List)
    - `Kelas Absensi` (List)
    - `Donasi Sxy` (List, Create, Get By ID, Update, Delete)
    - `Report` (SXY List, SXY Excel)
- **Sample Environment Configuration**:
  - Created [`apps-gin.postman_environment.json`](./gin/postman/apps-gin.postman_environment.json) (and symlink `app-gin.postman_environment.json`) containing standard Postman environment format with pre-configured variables: `baseUrl`, `apiVersion`, `authToken`, `resourceId`, `username`, and `password`.
- **Verification**:
  - Validated that all 122 endpoints with their complete request bodies, query parameters, path variables, UAT test scripts, and mock response examples are 100% preserved without any drops or modifications.

---

## \[Completed\] Gin: Endpoint `GET /v1/umats/popup` via Stored Procedure `SP_BUS_UMAT_SEARCH_POPUP`

### Summary & Changes Made:
- **Domain Response Model ([legacy_models.go](./gin/internal/domain/legacy_models.go))**:
  - Added `UmatPopUpResponse` matching the columns returned by `dbo.SP_BUS_UMAT_SEARCH_POPUP`: `id`, `kode`, `marga`, `alias`, `nama_indonesia`, `nama_mandarin`, `alamat`, `fotang_aktif_desc`, `fotang_ciu_tao_desc`, `pengajak`, and `penanggung`.
- **Service Implementation ([umat_service.go](./gin/internal/service/umat_service.go))**:
  - Added `func (s *UmatService) PopUp(c *gin.Context, page int, filters map[string]string, limit int) ([]domain.UmatPopUpResponse, int64, error)`.
  - Executes stored procedure `EXEC dbo.SP_BUS_UMAT_SEARCH_POPUP` without filtering by `SUBWHID`.
  - Supports filters: `nama_indonesia`, `nama_mandarin`, `alias`, `fotang_aktif`, and `fotang_ciu_tao`.
  - Dynamically scans rows using `nullFieldScanner` and extracts `totalRow` for pagination metadata.
- **Route Registration ([router.go](./gin/internal/api/router.go))**:
  - Registered `protected.GET("/umats/popup", ...)` with `middleware.StatusNotModifiedHeader` cache headers.
- **Unit & Acceptance Testing**:
  - Added `TestUmatService_PopUp` in `gin/internal/service/umat_service_test.go` covering parameter passing, pagination, and filter execution.
  - Added acceptance tests under `umat_crud_and_reports` in `gin/tests/e2e/master_data_e2e_test.go` verifying status `200 OK`, `meta.total >= 0`, and filtering capabilities.
- **Postman Documentation ([apps-gin.postman_collection.json](./gin/postman/apps-gin.postman_collection.json))**:
  - Added `Umat - PopUp` item under Master Data with complete query parameter definitions and sample `200 OK` JSON mock response.
- **Frontend Integration ([UmatForm.vue](./vue/src/components/umat/UmatForm.vue), [umat.ts](./vue/src/api/umat.ts), [umat.ts](./vue/src/types/umat.ts))**:
  - Added TypeScript interfaces `UmatPopUpItem`, `UmatPopUpQueryParams`, and `UmatPopUpResponse`.
  - Added `umatApi.getUmatsPopup` endpoint consumer in `api/umat.ts`.
  - Updated `UmatPopupSelector` in `UmatForm.vue` to fetch from `umatApi.getUmatsPopup` and aligned filter parameter keys (`nama_indonesia`, `nama_mandarin`, `alias`, `fotang_ciu_tao`, `fotang_aktif`).
- **Verification**:
  - Recompiled Windows distribution binary via `docker compose run --rm gin-build` (**PASS**).
  - Validated frontend TypeScript types via `npm run type-check` (`vue-tsc --noEmit`) (**PASS** with 0 errors).

---

## \[Completed\] Service Worker Abort Propagation & Rapid Pagination Loader Fix

### Summary & Changes Made:
- **Service Worker Abort Propagation (`vue/src/strategies/dynamicnetworkcache.ts` & `dynamiclownetworkcache.ts`)**:
  - Attached dedicated `AbortController` instances to deduplicated in-flight fetches, tracking subscribing `request.signal`s.
  - When all client subscribers abort (e.g. rapid pagination), the Service Worker automatically executes `abortController.abort()` to terminate the outbound HTTP/TCP connection to the backend server.
  - Prevented aborted requests from triggering unwanted debounced background updates (`triggerDebouncedBackgroundUpdate`) or delivering stale fallback cache responses.
  - Updated `DynamicLowNetworkCacheStrategy` to cleanly handle `request.signal?.aborted` in low-network mode.
- **Client Loader & State Race Condition Fix (`vue/src/views/report/SxyReportList.vue`, `UmatReportList.vue`, `DonasiSxyList.vue`)**:
  - Fixed rapid pagination bug where aborted previous requests executed `finally { loading.value = false }`, prematurely hiding the table loading spinner while the new page request was still in-flight.
  - Tied `loading.value = false`, `dataList.value`, and pagination/totals updates strictly to `currentAbortController === controller`, ensuring only the latest active request controls UI loading states and prevents stale data overwrites.
- **Verification**:
  - Ran `npm run type-check` (`vue-tsc --noEmit`) — Passed with 0 errors.
  - Ran `npm run build` (`vite build` + Service Worker build) — Succeeded with exit code 0.

---

## \[Completed\] Vue: QR Code SVG Display & Download in `UmatForm.vue`

### Summary & Changes Made:
- **Type Definitions ([umat.ts](./vue/src/types/umat.ts))**:
  - Added `qr_token?: string` property to the `Umat` interface to support typed QR token binding from backend responses.
- **Dependency Integration**:
  - Installed `qrcode` and `@types/qrcode` for scalable in-memory SVG generation from JWT tokens.
- **Component Implementation ([UmatForm.vue](./vue/src/components/umat/UmatForm.vue))**:
  - Added reactive generation of QR Code SVG using `QRCode.toString(token, { type: 'svg', margin: 1, errorCorrectionLevel: 'M' })`.
  - Added a responsive `.qr-code-section` in the photo column displaying the rendered SVG, action to enlarge via modal dialog (`el-dialog`), and action to download the vector SVG (`downloadQRCode`).
  - Added full modal dialog previewing enlarged QR Code alongside Umat name, alias, and kode.
- **Verification & Bug Fix**:
  - Resolved `ReferenceError: Cannot access 'formData' before initialization` by moving `qrCodeSvg` and its `watch` after `formData` definition in `<script setup>` (resolving Temporal Dead Zone).
  - Executed `npm run type-check` (`vue-tsc --noEmit`) (**PASS**).
  - Executed `npm run build` (**PASS** - bundled with zero errors).

---

## \[Completed\] Gin: JWT-based Umat QR Token Generation (`Get`) & Verification (`VerifyQR`)

### Summary & Changes Made:
- **Model Enhancements ([legacy_models.go](./gin/internal/domain/legacy_models.go))**:
  - Added `QRToken *string` (`gorm:"-" json:"qr_token,omitempty"`) to `type Umat struct`.
  - Added `VerifyQRRequest` struct (`qr_token string`).
  - Added `VerifyQRResponse` struct returning `claims`, `nama_indonesia`, `nama_mandarin`, `alias`, `fotang_ciu_tao`, and `fotang_aktif`.
- **Service Implementation ([umat_service.go](./gin/internal/service/umat_service.go))**:
  - Injected `cfg config.Config` into `UmatService` and updated `NewUmatService(db *gorm.DB, cfgs ...config.Config) *UmatService`.
  - Cleaned up JWT claims in `GenerateQRToken`: removed personal data (`id`, `nama_indonesia`, `fotang_ciu_tao`, `fotang_aktif`, etc.) from token payload to avoid bloating and leaking sensitive info; token claims now only carry identity `sub: item.ID`.
  - Added `umat.` prefix to generated QR token string (`item.QRToken = &qrToken` -> `umat.eyJ...`).
  - Updated `VerifyQR(tokenString string)`: automatically strips `umat.` prefix via `strings.TrimPrefix`, verifies JWT signature, decodes `sub` identity claim, acquires fresh profile data (`nama_indonesia`, `nama_mandarin`, `alias`) directly from the database, and resolves `fotang_ciu_tao` and `fotang_aktif` to the actual fotang names looked up from `T_APP_LOOKUP` with `CategoryId = 'B_FOTHANG'`.
- **Routing ([router.go](./gin/internal/api/router.go))**:
  - Registered `POST /umats/verify-qr` on the protected route group (mapped to `/v1/umats/verify-qr`).
  - Handles JSON payload validation via `respondValidationError` and returns `{ "data": result, "resource": "Umat" }` on success (200) or `{ "error": ... }` on failure (400).
- **Testing & Verification ([gin/internal/service/umat_service_test.go](./gin/internal/service/umat_service_test.go), [gin/tests/e2e/master_data_e2e_test.go](./gin/tests/e2e/master_data_e2e_test.go))**:
  - Unit tests: Added `TestUmatService_GenerateQRToken_ClaimsAndPrefix` to verify token generation with `umat.` prefix and claims containing only `sub` (with `nama_indonesia`, `fotang_ciu_tao`, `fotang_aktif`, `id` stripped).
  - Updated `TestUmatService_QRTokenAndVerifyQR` to verify both `umat.` prefixed tokens and non-prefixed raw tokens, asserting verified data comes from DB.
  - Verified via `go test -v -count=1 ./internal/service -run TestUmatService_GenerateQRToken_ClaimsAndPrefix` (**PASS**).
- **Postman Collection ([apps-gin.postman_collection.json](./gin/postman/apps-gin.postman_collection.json))**:
  - Updated `Umat - Verify QR` endpoint (`POST {{baseUrl}}/v1/umats/verify-qr`) request and response examples with `umat.eyJ...` prefix and stripped claims.

---

## \[Completed\] Vue: Readonly Mode in `KelasView.vue` for All 7 Sub-model Tables

### Summary & Changes Made:
- **Component Prop & Conditional Actions**:
  - Added `readonly?: boolean` prop (defaults to `false`) across all 7 kelas sub-model table components:
    - [KelasPesertaTable.vue](./vue/src/components/kelas/KelasPesertaTable.vue): hides "Tambah Peserta", "Add From History", and the "Aksi" column (Edit/Delete).
    - [KelasPengabdiTable.vue](./vue/src/components/kelas/KelasPengabdiTable.vue): hides "Tambah Pengabdi" and the "Aksi" column.
    - [KelasTopikTable.vue](./vue/src/components/kelas/KelasTopikTable.vue): hides "Tambah Topik" and the "Aksi" column.
    - [KelasKendaraanTable.vue](./vue/src/components/kelas/KelasKendaraanTable.vue): hides "Tambah Kendaraan" and the "Aksi" column.
    - [KelasDonasiTable.vue](./vue/src/components/kelas/KelasDonasiTable.vue): hides "Tambah Donasi" and the "Aksi" column.
    - [KelasDonasiBarangTable.vue](./vue/src/components/kelas/KelasDonasiBarangTable.vue): hides "Tambah Donasi Barang" and the "Aksi" column.
    - [KelasPengeluaranTable.vue](./vue/src/components/kelas/KelasPengeluaranTable.vue): hides "Tambah Pengeluaran" and the "Aksi" column.
- **View Integration ([KelasView.vue](./vue/src/views/transaction/KelasView.vue))**:
  - Passed `readonly` attribute to all 7 components in `KelasView.vue`, effectively rendering a clean read-only view while retaining full create/edit/delete capabilities in [KelasEdit.vue](./vue/src/views/transaction/KelasEdit.vue).
- **Verification**:
  - `npm run type-check` (`vue-tsc --noEmit`): **Passed (0 errors)**.
  - `npm run build:client`: **Production client build succeeded cleanly (20.68s)**.

---

## \[Completed\] Vue: "Add From History" Feature in `KelasPesertaTable.vue`

### Summary & Changes Made:
- **Component UI & User Experience (`vue/src/components/kelas/KelasPesertaTable.vue`)**:
  - Added button `"Add From History"` with `:icon="Timer"` beside `"Tambah Peserta"`.
  - Added multi-select `UmatPopupSelector` dialog with `:multiple="true"`, bound to `historyPopupVisible`.
  - Integrated with `fetchPreviousPesertaApi` calling `GET /v1/kelas/:id/peserta/load-previous`.
  - In the Add/Edit/History dialog:
    - Sets title to `"Add From History"` when in history mode.
    - Displays selected participants as closable tags in a container (`selected-tags-container`), with single-click removal (`removeHistoryUmat`), count summary badge, and a button to reopen selection.
    - Hides individual participant status lulus / keterangan lulus (only editable when editing an existing participant).
    - In footer, updates submit button label dynamically to reflect selected participant count: `Simpan ({count} Peserta)`.
    - Fully disables the form and buttons during submission with `v-loading="submitting"` and `:disabled="submitting"`, strictly executing `finally { submitting.value = false }`.
    - Dispatches to `kelasApi.createKelasPesertaBulk` with array of `id_peserta` and shared logistics / ikrar data.
- **API Client & TypeScript Types (`vue/src/api/kelas.ts`, `vue/src/types/kelas.ts`)**:
  - Defined `KelasPesertaPrevious`, `KelasPesertaPreviousListResponse`, `KelasPesertaBulkPayload`, and `KelasPesertaBulkResponse`.
  - Added typed API functions:
    - `loadPreviousKelasPeserta(id: number | string, params?: Record<string, any>, signal?: AbortSignal)`
    - `createKelasPesertaBulk(payload: KelasPesertaBulkPayload)`
- **Verification**:
  - TypeScript static type-checking passed cleanly: `npm run type-check` (`vue-tsc --noEmit`, exit code 0).
  - Production build succeeded cleanly: `npm run build` (built in 15.17s, exit code 0).

---

## \[Completed\] Vue: Fix Multi-Page Selection in `UmatPopupSelector` & Full-Width Form Keterangan

### Summary & Changes Made:
- **Keterangan Input Full Width (`vue/src/components/kelas/KelasPesertaTable.vue`)**:
  - Removed `<el-row>` and `<el-col :md="12" :sm="24">` wrapper around `<el-form-item label="Keterangan">`, allowing the keterangan input to span the full 100% width of the dialog form consistently with other items.
- **Persistent Multi-Page Selection (`vue/src/components/common/UmatPopupSelector.vue`)**:
  - Added `:row-key="getRowKey"` on `<el-table>` and `:reserve-selection="true"` on `<el-table-column type="selection">`.
  - Added `getRowKey` computing consistent string identity (`row.id_peserta ?? row.id ?? row.kode`).
  - Added `selected?: any[]` prop and `rowKey?: string | ((row: any) => string | number)` prop.
  - In `handleSelectionChange`, properly preserved selected items belonging to off-page records when pagination data updates.
  - In `fetchData`, re-synchronized checkboxes for rows on the active page matching existing selections via `tableRef.value.toggleRowSelection(row, true)`.
  - In `action-toolbar`, displayed the real-time cumulative count indicator (`{{ selectedMultipleRows.length }} data terpilih`) so users have clear visibility across all pages.
  - In dialog footer, updated the confirm button label to `Pilih ({count})` reflecting the total selected across all pages.
  - Bound `:selected="selectedHistoryUmats"` from `KelasPesertaTable.vue` so reopening the selector restores previous selections.
- **Tag Ellipsis & Truncation (`vue/src/components/kelas/KelasPesertaTable.vue`)**:
  - Set `max-width: 60px` with `overflow: hidden; text-overflow: ellipsis; white-space: nowrap;` on `.selected-tags-container :deep(.el-tag)` and `.el-tag__content`.
  - Added `:title` tooltip on each `<el-tag>` to preserve full readability on hover.
- **Verification**:
  - `npm run type-check` (`vue-tsc --noEmit`): **Passed (0 errors)**.
  - `npm run build`: **Production build succeeded cleanly (15.79s)**.

---

## \[Completed\] General JSON Parse & Payload Error Handler (`ParseJSONError`)

### Summary & Changes Made:
- **Router Implementation (`gin/internal/api/router.go`)**:
  - Implemented `ParseJSONError(err error) map[string]string`:
    - Handles syntax corruption errors (`*json.SyntaxError`, `io.ErrUnexpectedEOF`) returning user-friendly `"Format payload JSON tidak valid atau rusak."` under key `"general"`.
    - Handles empty payloads (`io.EOF`, `"EOF"`) returning `"Payload request tidak boleh kosong."` under key `"general"`.
    - Handles specific field type mismatches (`*json.UnmarshalTypeError`), mapping custom messages for `idpeserta` / `id_peserta`, and dynamic fallback `"Field '{field}' memiliki tipe data yang tidak sesuai."` for any other field.
    - Gracefully handles formatted struct validation or unstructured data errors.
  - Integrated with `FormatValidationError(err)`:
    - Converts parsed JSON error map into `map[string][]string` for client contracts (`details: Record<string, string[]>`).
  - Integrated with `respondValidationError(c, err)`:
    - Sanitizes the top-level `"error"` message, completely preventing Go internal error text (like `invalid character '...'`, `unexpected EOF`, `cannot unmarshal string into Go struct field ...`) from leaking to users.
- **Unit & Integration Tests (`gin/internal/api/router_test.go`)**:
  - Added unit test `TestParseJSONError` checking `*json.SyntaxError`, `*json.UnmarshalTypeError`, and `nil`.
  - Added API route-level tests in `TestKelasPeserta_PayloadTypeValidation`:
    - Verified syntax errors (`{"id_peserta": `) return status 400 with `"Format payload JSON tidak valid atau rusak."`.
    - Verified empty payloads (`""`) return status 400 with `"Payload request tidak boleh kosong."`.
- **Verification**:
  - Unit tests passed: `go test -v -count=1 ./internal/api -run "TestParseJSONError|TestKelasPeserta_PayloadTypeValidation"`.

---

## \[Completed\] Custom JSON Unmarshaling Breakdown for `KelasPeserta` and `KelasPesertaBulkRequest`

### Summary & Changes Made:
- **Domain Models (`gin/internal/domain/legacy_models.go`)**:
  - Implemented `func (p *KelasPeserta) UnmarshalJSON(data []byte) error`:
    - Enforces that `id_peserta` (or `idpeserta`) must be a single umat value (number or string representation of integer).
    - If an array is passed (e.g., `{"id_peserta": [101, 102]}`), it explicitly fails with `idpeserta: idpeserta should be single value of umat` instead of leaking internal Go `"cannot unmarshal"` error messages.
    - Accurately parses valid integer or numeric string values into `p.IdPeserta` without error.
  - Implemented `func (p *KelasPesertaBulkRequest) UnmarshalJSON(data []byte) error`:
    - Enforces that `id_peserta` (or `idpeserta`) must be a JSON array.
    - If a single value (string, integer, object, etc.) is passed (e.g., `{"id_peserta": 101}` or `{"id_peserta": "101"}`), it explicitly fails with `idpeserta: idpeserta must array` instead of default Go unmarshal errors.
    - Flexible support for decoding `[]int32` or string arrays `[]string` of IDs into `p.IdPeserta`.
- **API Error Formatting Integration (`gin/internal/api/router.go`)**:
  - The errors returned (`idpeserta: idpeserta must array` and `idpeserta: idpeserta should be single value of umat`) seamlessly plug into `FormatValidationError(err)`.
  - Produces clean structured responses:
    - For bulk endpoint type mismatch:
      ```json
      {
        "error": "idpeserta: idpeserta must array",
        "details": {
          "idpeserta": ["idpeserta must array"]
        }
      }
      ```
    - For single participant create/update endpoint type mismatch:
      ```json
      {
        "error": "idpeserta: idpeserta should be single value of umat",
        "details": {
          "idpeserta": ["idpeserta should be single value of umat"]
        }
      }
      ```
- **Postman Collection (`gin/postman/apps-gin.postman_collection.json`)**:
  - Added `400 Bad Request - IdPeserta Should Be Single Value` under `Kelas Peserta - Create`.
  - Added `400 Bad Request - IdPeserta Must Array` under `Kelas Peserta - Create Bulk`.
- **Unit & Integration Tests (`gin/internal/service/kelas_peserta_service_test.go`, `gin/internal/api/router_test.go`)**:
  - Updated `TestKelasPeserta_PayloadTypeValidation` in `kelas_peserta_service_test.go` to assert exact error strings `idpeserta should be single value of umat` and `idpeserta must array`.
  - Updated `TestKelasPeserta_PayloadTypeValidation` in `router_test.go` to assert the parsed JSON `details` map has `details["idpeserta"]` containing the exact error messages.
- **Verification**:
  - `go test -v ./internal/service -run "TestKelasPeserta"` passed.
  - `go test -v -count=1 ./internal/api -run "TestKelasPeserta_PayloadTypeValidation"` passed.
  - Docker container built and restarted successfully.

---

## \[Completed\] Endpoint `POST /v1/kelas/:id/peserta/bulk` & `CreateBulk` Implementation

### Summary & Changes Made:
- **Domain Model (`gin/internal/domain/legacy_models.go`)**:
  - Added `KelasPesertaBulkRequest` struct supporting `IdPeserta []int32` with `validate:"required,min=1"` alongside all shared participant properties (`sumbangan`, `barang`, `tim_kerja`, `keterangan`, `status`, `lulus`, `keterangan_lulus`, `anak`, `suster`, `menginap`, `makanan_pagi`, `makanan_siang`, `makanan_malam`).
- **Service Implementation (`gin/internal/service/kelas_peserta_service.go`)**:
  - Consolidated validation into a single `validatePesertaLookups` function accepting `*domain.KelasPesertaBulkRequest` (removed redundant single validator). Callers `Create` and `Update` transform their single-item payload (`IdPeserta *int32`) into bulk request (`[]int32{*p.IdPeserta}`).
  - Validates `trx_id` in `Kelas`, batch lookups `id_peserta` in `Umat` (`id IN (?)`), and checks `tim_kerja` in `AppLookup` concurrently using `wg.Go` and separate GORM sessions.
  - Implemented `func (s *KelasPesertaService) CreateBulk(payload domain.KelasPesertaBulkRequest, c *gin.Context) ([]domain.KelasPeserta, error)`:
    - Deduplicates `IdPeserta` preserving order.
    - Uses a single database transaction (`s.db.Transaction`) to generate IDs via `SP_APP_GenerateId`, update SD Pemula records (`updateUmatSDPemula`), and insert each participant into `T_TRX_KELAS_PESERTA`.
    - Batch preloads `Umat` details for all created participants before returning.
- **API Routing (`gin/internal/api/router.go`)**:
  - Registered `protected.POST("/kelas/:id/peserta/bulk", ...)` handling path parameter `:id` binding and JSON deserialization into `KelasPesertaBulkRequest`.
- **Postman Collection (`gin/postman/apps-gin.postman_collection.json`)**:
  - Added `Kelas Peserta - Create Bulk` under `Kelas` folder with array payload, test script assertions (status 201, `KelasPeserta` resource, array return), sample success response (`201 Created`), and validation error response (`400 Bad Request`).
- **Unit & Integration Tests (`gin/internal/service/kelas_peserta_service_test.go`, `gin/internal/api/router_test.go`, `gin/tests/e2e/transactions_e2e_test.go`)**:
  - Added `TestKelasPesertaService_CreateBulk` covering validation of empty arrays, invalid foreign keys, and successful multi-record insertion with verified generated `detail_id` and rollback cleanup.
  - Added `TestKelasPeserta_PayloadTypeValidation` in `kelas_peserta_service_test.go` and `router_test.go`:
    - Tests that passing an array payload `{"id_peserta": [101, 102]}` to single `POST /v1/kelas/:id/peserta` returns `400 Bad Request` with unmarshal validation error.
    - Tests that passing a single number payload `{"id_peserta": 101}` to bulk `POST /v1/kelas/:id/peserta/bulk` returns `400 Bad Request` with unmarshal validation error.
  - Updated E2E acceptance tests in `transactions_e2e_test.go` verifying both payload type mismatch error cases.
- **Verification**:
  - Docker container built and deployed successfully (`docker compose build gin && docker compose up -d gin`).
  - Unit tests passed: `go test -v ./internal/service -run "TestKelasPeserta"` (PASS) and `go test -v ./internal/api -run "TestKelasPeserta_PayloadTypeValidation"` (PASS).
  - Verified live endpoint with cURL: 400 on empty list, 400 on non-existent `id_peserta`, 400 on array payload for single create, 400 on single payload for bulk create, and 201 Created on valid insertion.

---

## \[Completed\] Endpoint `GET /v1/kelas/:id/peserta/load-previous` & SP Integration (`SP_TRX_KELAS_GET_PESERTA_BY_CODE_AND_LEVEL`)

### Summary & Changes Made:
- **Domain Model (`gin/internal/domain/legacy_models.go`)**:
  - Created `KelasPesertaPrevious` struct mapping output fields from `SP_TRX_KELAS_GET_PESERTA_BY_CODE_AND_LEVEL`:
    `id_peserta`, `id`, `kode`, `nama_indonesia`, `nama_mandarin`, `alias`, `marga`, `alamat`, `fotang_aktif`, `fotang_aktif_desc`, `fotang_ciu_tao`, `fotang_ciu_tao_desc`, `pengajak`, `penanggung`.
- **Service Implementation (`gin/internal/service/kelas_peserta_service.go`)**:
  - Implemented `func (s *KelasPesertaService) LoadPrevious(id string, c *gin.Context, page int, limit int) ([]domain.KelasPesertaPrevious, int64, error)`.
  - Executed queries for `kodeKelas` (from `T_TRX_KELAS`) and `subWhId` (from `T_WH_USER_MATRIX_MST` / `AdminMatrix`) concurrently using `wg.Go` and separate GORM sessions (`s.db.Session(&gorm.Session{})`) with connection pooling.
  - Calls `EXEC [dbo].[SP_TRX_KELAS_GET_PESERTA_BY_CODE_AND_LEVEL] @TrxId = ?, @KodeKelas = ?, @SubWhId = ?`.
  - Removed in-memory filters (`namaindonesia`, `namamandarin`, `alias`, `fotangaktif`, `fotangciutao`) since the stored procedure does not facilitate filtering, keeping code lightweight with minimal GC overhead.
  - Applies pagination (`page`, `limit`) with slice pre-allocation.
- **API Routing (`gin/internal/api/router.go`)**:
  - Registered `protected.GET("/kelas/:id/peserta/load-previous", ...)` before detail routes to prevent routing conflicts.
- **Stored Procedure Update (`dbo.SP_TRX_KELAS_GET_PESERTA_BY_CODE_AND_LEVEL.StoredProcedure.sql`)**:
  - Updated to return `fotangaktif`, `FotangAktifDesc`, `fotangciutao`, and `FotangCiuTaoDesc`.
  - Supported `@SubWhId = 0` (`@SubWhId = 0 or b.fotangaktif = @SubWhId`) allowing global admins to query past participants across all fotang.
- **Postman Collection (`gin/postman/apps-gin.postman_collection.json`)**:
  - Added and cleaned `Kelas - Peserta Load Previous` under `Kelas` folder with pagination query parameters (`page`, `limit`), test assertions (status 200, array schema, pagination meta), and sample mock response.
- **Unit & Acceptance Testing (`gin/internal/service/kelas_peserta_service_test.go` & `gin/tests/e2e/transactions_e2e_test.go`)**:
  - Added unit test `TestKelasPesertaService_LoadPrevious` verifying invalid format, non-existent class, valid class, and pagination limits.
  - Added E2E acceptance test in `TestKelas_EndToEnd` verifying `GET /v1/kelas/:id/peserta/load-previous` with pagination.
- **Verification**:
  - Verified Docker container compilation with `docker compose build gin`.
  - Verified unit test with `go test -v ./internal/service -run "TestKelasPesertaService"` (PASS).
  - Verified acceptance tests with `go test -v -count=1 ./tests/e2e -run "TestKelas_EndToEnd"` (PASS).

---

## \[Completed\] Responsive Umat Selector & Decoupled `UmatPopupSelector.vue` Dialog

### Summary & Changes Made:
- **Decoupled Reusable Dialog Component (`vue/src/components/common/UmatPopupSelector.vue`)**:
  - Implemented an API-free, fully property-controlled dialog selector for Umat.
  - Receives `fetchApi` function prop handling pagination, page size, and dynamic query filtering.
  - Dynamically configurable text filters via `:filterName` (default: `namaindonesia`, `namamandarin`, `alias`; can be set to `false`/`null` or custom mappings).
  - Dynamically configurable Fotang dropdown filter via `:filterFotang` (supports `fetchFotangApi` and `fotangOptions` props; can be set to `false`/`null`).
  - Dynamically configurable columns via `:Columns` / `:columns` prop with customizable widths and alignments.
  - Supports both single selection (`multiple: false`) with radio button, row click, double-click instant selection, and multi-selection (`multiple: true`) with checkboxes.
  - Emits `@select` and `@onselect` feedback upon confirmation or double-click selection.
  - Includes responsive search form, "Add Content" primary button, data counter tag, empty state, and full Element Plus pagination controls.
- **Responsive Umat Selector in `KelasPengabdiTable.vue`, `KelasPesertaTable.vue`, & `UmatForm.vue`**:
  - Maintained `<el-select>` / `<LookupSelect>` remote search dropdown exclusively on mobile viewports (`isMobile: width < 768px`).
  - Created custom popup dialog trigger with readonly input and `...` ellipsis append button on desktop and tablet viewports (`!isMobile: width >= 768px`).
  - Integrated `UmatPopupSelector` with `umatApi.getUmats` and `lookupApi.getLookupFotang`.
  - In `UmatForm.vue`: Added dual Fotang filters (`fotang_chiutao` / Fotang Chiu Tao and `fotang_aktif` / Fotang Aktif) alongside the 3 name filters (`namaindonesia`, `namamandarin`, `alias`) for both Pengajak and Penanggung selectors with dynamic dialog title (`Popup Pengajak` / `Popup Penanggung`).
  - In `umat.ts`: Forwarded all filter parameters in `getUmats`.
  - In `umat_service.go`: Whitelisted `fotang_aktif`, `fotangaktif`, `fotang_chiutao`, and `fotangciutao` in query filter rules.
  - Automatically syncs selected Umat display label with `form.id_pengabdi`, `form.id_peserta`, `formData.pengajak`, `formData.penanggung`, `umatOptions`, Ikrar 1-6 updates for Suai Sing Pan, and form validation.
- **Verification**:
  - Verified with `vue-tsc --noEmit` (`npm run type-check` exited with code 0).
  - Verified with `docker compose build gin` (exited with code 0).

---

## \[Completed\] End-to-End (E2E) Acceptance Testing in Go (`gin/tests/e2e/`)

### Summary & Changes Made:
- **E2E Test Architecture & Factory Layer (`gin/tests/e2e/`)**:
  - Implemented modular, fast E2E test framework using `github.com/gavv/httpexpect/v2` and `github.com/stretchr/testify`.
  - Built automated JWT authentication workflow using `admin` / `password`, dynamic token extraction, and cached authentication helper (`getAuthenticatedExpect(t)`).
  - Built fake factory payload generators in `factory/` (`admin_factory.go`, `master_factory.go`, `transaction_factory.go`) providing valid (HTTP 200/201) and invalid (HTTP 400 validation error) models with realistic randomized data.
- **Test Suites Across All API Endpoints**:
  - `auth_e2e_test.go`: Tests Ping (`/ping`), Login (`POST /login` with 200 OK, 400 bad JSON, 401 invalid credentials), and Change Password (`PATCH /change-password` with 400).
  - `lookup_e2e_test.go`: Tests all 14 lookup endpoints for HTTP 200 OK and validates unauthenticated rejection (HTTP 401 Unauthorized).
  - `admin_mgmt_e2e_test.go`: Tests Department (`GET /v1/departments`), Admins CRUD (`/v1/admins`), Admin Groups CRUD (`/v1/admin-groups`), Group Menu Mappings CRUD (`/v1/group-menus`), and Admin Sub Warehouses CRUD (`/v1/admin-sub-warehouses`).
  - `master_data_e2e_test.go`: Tests Umat CRUD & Reports (`/v1/umats`), Topics CRUD (`/v1/topic`), Kelas Master CRUD (`/v1/kelas-master`), Activity CRUD (`/v1/activities`), Tim Kerja CRUD & Lookups (`/v1/tim-kerja`), Tahun Ciu Tao CRUD (`/v1/tahun-ciu-tao`), Penggalang Dana CRUD (`/v1/penggalang-dana`), SXY Donatur CRUD (`/v1/sxy-donatur`), and Fotang Lookups (`/v1/fotang/lookup`).
  - `transactions_e2e_test.go`: Tests Kelas CRUD (`/v1/kelas`), nested sub-resources (Peserta, Pengabdi, Topik, Kendaraan, Donasi, Donasi Barang, Pengeluaran, Musik, Absensi), SSRS Report generation, and Donasi SXY CRUD (`/v1/donasi-sxy`) with SSRS Excel export (`/v1/donasi-sxy/report/excel`).
- **Status Code & Validation Criteria**:
  - All GET and DELETE endpoints asserted for HTTP 200 OK.
  - All POST and PATCH endpoints tested with fake factory data for both HTTP 200/201 (success) and HTTP 400 (validation failure).
- **Backend Fixes & Database Adaptations**:
  - Resolved SQL Server datetime out-of-range errors on zero timestamps (`0001-01-01`) by modifying `DateOnly.Value()` and `DateTime.Value()` in `legacy_models.go` to return formatted date strings or `nil`.
  - Added association omission (`Omit("AdminGroup", "Department")`) and non-zero partial updates in `AdminService` to prevent foreign key errors.
  - Fixed parameter count discrepancies for SQL Server stored procedures `SP_BUS_RPT_UMAT` (13 parameters) and `SP_SXY_RPT_TRANSAKSI` (5 parameters).
  - Updated router handlers to support dual path & query parameter lookups for `tahun-ciu-tao`, `activity`, and `tim-kerja`.
- **Verification & Stored Procedure Pagination Updates**:
  - Validated `/v1/umats/report` (`SP_BUS_RPT_UMAT`) with 15 parameters returning paginated rows with `meta.page`, `meta.limit`, and `meta.total` (`total = 19503`).
  - Validated `/v1/donasi-sxy/report` (`SP_SXY_RPT_TRANSAKSI`) with 7 parameters returning paginated rows with `meta.page`, `meta.limit`, `meta.total` (`total = 45239`), and `meta.total_jumlah`.
  - Updated E2E test suites in `gin/tests/e2e/` (`master_data_e2e_test.go` and `transactions_e2e_test.go`) to strictly assert `meta.total > 0` on list and report responses.
  - Rebuilt Docker `gin_app` container (`docker compose up -d --build gin`) and ran full E2E test suite (`go test -v ./tests/e2e/...`), verifying all test suites passing with 100% success (`PASS` in 92.8s).

---

## \[Completed\] User Acceptance Tests (UAT) & Sample Mock Responses for All 117 Endpoints (`gin/postman/apps-gin.postman_collection.json`)

### Summary & Changes Made:
- **Login Credentials & Dynamic Token Authentication**:
  - Updated Login endpoint credentials in [`apps-gin.postman_collection.json`](./gin/postman/apps-gin.postman_collection.json) to `admin` / `password`.
  - Added UAT test script in Login endpoint to verify HTTP 200 OK, token presence, and automatically extract and set `authToken` in Postman collection variables (`pm.collectionVariables.set("authToken", jsonData.token)`).
- **Comprehensive User Acceptance Test (UAT) Scripts**:
  - Configured automated test scripts (`event` with `listen: "test"`) across all 117 endpoints covering:
    - **Status Code Validation**: Verifies HTTP 200 OK for GET/PATCH/DELETE, and HTTP 201 Created or 200 OK for POST.
    - **Latency SLA**: Ensures API response time is within SLA thresholds (< 2000ms for queries, < 3000ms for mutations).
    - **Schema & Structure**: Asserts JSON response structures, checking `data` array/object, pagination `meta` (`page`, `limit`, `total`), and message properties.
    - **Excel Exports**: Validates `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` and `Content-Disposition: attachment; filename=...`.
- **Sample Output Examples (Saved Responses)**:
  - **GET Endpoints (65 requests)**: Added realistic HTTP 200 OK response examples with domain model payloads.
  - **POST, PATCH, DELETE Endpoints (52 requests)**: Added dual response examples for each endpoint:
    - **Success (HTTP 200 OK / 201 Created)**: Domain-accurate payload structure with created/updated data or deletion confirmation.
    - **Error (HTTP 400 Bad Request)**: Validation error payload matching the backend's `{"error": "...", "details": {...}}` contract.
- **Verification**:
  - Automated verification confirmed 117 of 117 endpoints configured with test scripts.
  - Confirmed 0 missing test scripts and 0 missing sample responses.

---

## \[Completed\] Docker Daemon Recovery & Unresponsive Socket Fix


### Summary & Changes Made:
- **Diagnosed Docker Failure**:
  - Found `docker info` hanging indefinitely and throwing `permission denied while trying to connect to the docker API at unix:///Users/xicond/.docker/run/docker.sock`.
  - Investigated process table and discovered stale Docker Desktop background helper processes (`com.docker.backend`, `com.docker.virtualization`, `docker-agent`) had hung since Sep 3 while the main `Docker.app` was inactive, leaving an unresponsive domain socket.
- **Process Cleanup & Service Relaunch**:
  - Terminated hung CLI and background processes (`kill -9`, `killall -9`).
  - Relaunched Docker Desktop cleanly via `/Applications/Docker.app`.
  - Monitored startup until socket listener initialized and responded.
- **Automated Recovery Script (`fix-docker.sh`)**:
- **Verification**:
  - `docker info` responds instantly with Docker Engine version `29.7.2`.
  - `docker ps` verified all workspace containers (`gin_app`, `redis_app`, `mssql_server`, `vue_dev`) are running and healthy.

---

## \[Completed\] Master Data Kelas Service & Vue Integration (`gin/internal/service/kelas_master_service.go`, `gin/internal/api/router.go`, `vue/src/views/master-data/KelasMasterList.vue`, `vue/src/layouts/MainLayout.vue`, `vue/src/router/index.ts`)

### Summary & Changes Made:
- **Backend Service `KelasMasterService` (`gin/internal/service/kelas_master_service.go`)**:
  - Targets `T_APP_LOOKUP` with `CategoryId = 'B_KELASKHUSUS'`.
  - **Automated LookupValue and LookupId Generation**: Evaluates existing numeric `LookupValue` records in the database, determines the maximum integer, and auto-generates the next sequential 3-digit zero-padded value (e.g. `023`) and LookupId (`B_KELASKHUSUS023`) if not supplied by the client.
  - **Uniqueness Validation**: `validateDescription` ensures `LookupDescription` is non-empty, max 150 characters, and unique among active records (`Status = true` / `1`) within category `B_KELASKHUSUS`.
  - **Combined Validation**: `validateKelasMaster` runs `ValidateStruct` and `validateDescription` together, consolidating all errors into a single `*ValidationError` for simultaneous delivery to the frontend.
  - **Concurrent Queries**: `List` runs `Count` and `Find` concurrently via `sync.WaitGroup` on separate GORM sessions for high performance.
  - **CRUD Operations**: Full support for `List`, `Get`, `Create`, `Update` (with duplicate check excluding self), and `Delete` (soft delete `Status = false`, `ModAct = 'D'`).
  - **Unit Testing**: Added unit tests in `kelas_master_service_test.go` covering auto-number generation, duplicate name detection, combined struct validation, update rules, and delete behavior. Verified passing (`0.31s`).
- **Gin Router & Postman**:
  - Registered routes in `router.go`: `GET /v1/kelas-masters` (list with ETag caching), `GET /v1/kelas-master` (item by ID or list fallback), `GET /v1/kelas-masters/:id`, `POST /v1/kelas-master`, `PATCH /v1/kelas-master`, and `DELETE /v1/kelas-master`.
  - Invalidates Redis lookup cache `lookup:B_KELASKHUSUS` upon create/update/delete.
  - Added 5 requests to Postman collection `apps-gin.postman_collection.json`.
- **Vue Frontend (`KelasMasterList.vue`, `kelasMasterApi`, `types/kelasMaster`)**:
  - Created TypeScript definitions in `vue/src/types/kelasMaster.ts` and API client in `vue/src/api/kelasMaster.ts`.
  - Created full management view `KelasMasterList.vue` with filter by `lookup_value`, `lookup_description`, and `status`, pagination, and responsive layout.
  - **Auto-Suggest**: On create, typing 2 or more words in `LookupDescription` (`el-autocomplete`) queries existing kelas. Selecting an existing class switches to the edit modal for that record.
  - **Duplicate Link**: If backend rejects with duplicate name, shows `"..., update here instead"` linking directly to edit mode for the existing record.
  - **Automated Kode Kelas**: Input for `LookupValue` is optional with clear helper hints; empty values auto-generate on save.
  - **Submitting Lock**: Page & form inputs are disabled during submission (`:disabled="submitting"`), wrapped with `try ... finally { submitting.value = false }`.
- **Route & Menu Isolation & Permission Guard**:
  - Mounted `/master-data/kelas` in `vue/src/router/index.ts` with `name: 'master-kelas'` and explicit `activeMenu: '/master-data/kelas'`.
  - Added menu item under Master Data in `vue/src/layouts/MainLayout.vue` (`v-if="hasSubMenu('Master Data', 'Kelas')" index="/master-data/kelas"`).
  - Explicit index paths prevent any active menu CSS conflicts with `/transaction/kelas`.
  - Added fallback in `authStore.hasSubMenu` (`vue/src/stores/auth.ts`) so that if `Master Data -> Kelas` is not yet populated in the database table `T_Login_Menu`, it automatically permits navigation as long as the user has access to `Master Data`, resolving the issue where `router.beforeEach` previously blocked navigation and redirected to `/dashboard`.
- **Verification**:
  - Unit tests passed: `TestKelasMasterService` (0.31s).
  - TypeScript type-check passed: `npm run type-check` (`vue-tsc --noEmit`) with 0 errors.
  - Windows binary compiled: `docker compose run --rm gin-build` (`server.exe`) with 0 errors.

---

## \[Completed\] Combined Struct & Duplicate Validation with Descriptive Clues (`gin/internal/api/router.go`, `gin/internal/service/topic_service.go`, `gin/internal/api/router_test.go`)

### Summary & Changes Made:
- **Direct JSON Decoding on Topic Endpoints in `router.go`**:
  - Replaced Gin's `c.ShouldBindJSON(&payload)` with `json.NewDecoder(c.Request.Body).Decode(&payload)` on both `POST /topic` and `PATCH /topic`.
  - Previously, Gin's default struct validator was short-circuiting on struct validation tags before `topicService.Create` or `topicService.Update` was reached, discarding the duplicate name check.
  - With direct decoding, `validateTopic` runs inside the service and executes both `ValidateStruct` and `validateTopicName`, successfully bundling struct validation errors and duplicate name errors into a single combined response.
- **Descriptive Validation Clues in `respondValidationError`**:
  - Changed `respondValidationError` in `router.go` from static `"error": "Invalid Input"` to `errMsg := err.Error()`.
  - Now, HTTP 400 Bad Request responses explicitly include the human-readable validation error string (e.g. `"topic_name: Nama topik '...' sudah ada"`) alongside the structured `details` map.
- **Field Preservation on Partial Update**:
  - In `TopicService.Update`, preserved existing `TopicName` and `TopicCategory` when they are omitted (or empty) in `PATCH` requests, preventing accidental field wipes.
- **Verification & Testing**:
  - Added unit test cases `TestValidationErrorReturns400` and `TestCombinedValidationErrors` in `router_test.go`.
  - Rebuilt and started the `gin` container (`docker compose up -d --build gin`).
  - Tested live against the running backend with `admin` authentication:
    1. Struct error + duplicate name: Returned HTTP 400 with both `topic_category`, `topic_code`, and `topic_name` in `details` and in `error`.
    2. Struct valid + duplicate name: Returned HTTP 400 with `topic_name: Nama topik '...' sudah ada` in `error` and `details`.
    3. Update with unchanged name: Successfully preserved name and updated description.
  - Verified Windows compilation with `docker compose run --rm gin-build`.
  - Verified Vue type-check and client build with `npm run type-check` and `npm run build:client`.

- **IIS `httpErrors` PassThrough Configuration (`gin/dist/web.config`)**:
  - Changed `<httpErrors existingResponse="Replace">` to `<httpErrors existingResponse="PassThrough">`.
  - In IIS, `existingResponse="Replace"` intercepts and replaces *all* HTTP error responses (including 400 Bad Request, 422, and 500) generated by `httpPlatformHandler`/Gin with generic IIS error pages, stripping the JSON body.
  - Setting `existingResponse="PassThrough"` instructs IIS to leave application error response bodies intact so clients receive the JSON error details and messages.

---



### Summary & Changes Made:
- **Auto-Suggest on Create Dialog (`el-autocomplete`)**:
  - Replaced `<el-input>` for `topic_name` in create mode with `<el-autocomplete>`.
  - Configured suggestions to trigger when input contains 2 or more words (`split(/\s+/).filter(Boolean).length >= 2`).
  - Added request cancellation via `AbortController` (`suggestAbortController`) to avoid redundant requests.
  - Selecting an item from the suggestions dropdown immediately triggers `handleEdit(item.topic_code)` to switch the dialog into edit mode for that topic.
- **Duplicate Topic Name Error Link**:
  - On form submission (`handleSubmit`), captures backend `details.topic_name` validation error.
  - If a duplicate name error is detected (containing `"sudah ada"`), queries existing topic by name to retrieve its `topic_code`.
  - Appends an inline link `", update here instead"` next to the error message that opens the edit dialog for that existing topic (`handleEdit(duplicateTopicCode)`).
- **Form Submission Guard**:
  - Added `:disabled="submitting"` to `<el-form>` to disable all inputs during save.
  - Handled `finally { submitting.value = false }`.
- **Verification**:
  - Verified TypeScript types with `npm run type-check` (`vue-tsc --noEmit`), passing with 0 errors.
  - Verified production build with `npm run build:client`, passing with 0 errors.

---

## \[Completed\] Topic Name Uniqueness Validation & Combined Struct Validation (`gin/internal/service/topic_service.go`, `gin/internal/service/topic_service_test.go`)

### Summary & Changes Made:
- **`validateTopic` Helper with Combined Error Reporting**: Implemented `validateTopic(payload domain.Topic, excludeCode string) error` on `TopicService` in [topic_service.go](./gin/internal/service/topic_service.go).
  - Executes struct validation (`ValidateStruct`) and uniqueness check (`validateTopicName`) together.
  - Aggregates all validation error details (e.g. `topic_category`, `topic_name`, `description`, etc.) into a unified `Details map[string][]string`.
  - Returns a single combined `*ValidationError`, ensuring all validation failures are delivered simultaneously.
- **`validateTopicName` Helper**: Added `validateTopicName(topicName string, excludeCode string) error` method to `TopicService` in [topic_service.go](./gin/internal/service/topic_service.go).
  - Queries `T_BUS_TOPIC` for records where `TopicName = ?` and `Status = true`.
  - When `excludeCode` is provided, filters `TopicCode <> excludeCode` to allow updating an existing record without self-conflict.
  - Returns structured `*ValidationError` with key `topic_name` (`"Nama topik '<name>' sudah ada"`).
- **Create Endpoint Validation**:
  - Validates struct tags and uniqueness together on `Create`. Rejects duplicate active topic names alongside any field constraints.
- **Update Endpoint Validation**:
  - Sets `payload.TopicCode = id` when missing so struct validation passes cleanly.
  - If `topic_name` is unchanged, duplicate validation passes for its own record. If changed, verifies no other active record uses the new name.
  - Delivers any struct validation errors and duplicate name errors combined.
- **Verification**:
  - Added unit test cases in [topic_service_test.go](./gin/internal/service/topic_service_test.go) verifying duplicate rejection on create, update keeping current name, update to duplicate name, update to a new unique name, and multi-error combined delivery on both create and update.
  - Verified with `docker compose run --rm gin-build /bin/sh -c "go test -v ./internal/service -run TestTopicService"` (**PASS**) and compiled server binary with `docker compose run --rm gin-build` (**PASS**).

---

## \[Completed\] JWT Claims Redis Caching & User Token Eviction (`gin/internal/service/auth_service.go`, `gin/internal/api/middleware/auth.go`, `gin/internal/api/middleware/auth_test.go`)

### Summary & Changes Made:
- **`CacheJWTToken` Implementation**: Added `CacheJWTToken(userID int32, tokenString string)` method to `AuthService` in [auth_service.go](./gin/internal/service/auth_service.go).
  - Automatically called before successful return from `Login`.
  - Calculates Redis TTL based on token expiration minus 200ms (`exp - 200ms`).
  - Serializes all JWT claims as JSON and stores under `jwt:token:<token_string>`.
- **Active User Token Eviction**:
  - Tracks user's active tokens under Redis set `jwt:user:<userID>:tokens`.
  - When the same user logs in again, all previously cached tokens for that user are evicted (`DEL` on token keys and set key) before caching the new token.
- **Bypass Signature Parsing in Middleware**:
  - Updated `AuthMiddleware` in [auth.go](./gin/internal/api/middleware/auth.go).
  - Checks Redis for `jwt:token:<token_string>`. On cache hit, deserializes claims, sets `userID` context key (`c.Set("userID", claims["sub"])`), and calls `c.Next()` immediately (skipping signature parsing & verification).
  - Falls back to standard `jwt.Parse` validation if cache miss occurs or if Redis is unavailable.
- **Verification**: Added unit test `TestAuthMiddleware_Unauthorized` in [auth_test.go](./gin/internal/api/middleware/auth_test.go) and verified `go test -v ./internal/api/middleware` and `./internal/config`, passing cleanly with 0 errors.

---

## \[Completed\] Excel Report Download Activation Condition (`vue/src/views/report/UmatReportList.vue` & `vue/src/views/report/SxyReportList.vue`)

### Summary & Changes Made:
- **`hasActiveFilter` Computed Property**: Added a `hasActiveFilter` computed property in [UmatReportList.vue](./vue/src/views/report/UmatReportList.vue) and [SxyReportList.vue](./vue/src/views/report/SxyReportList.vue) that evaluates to `true` only when at least one filter field is active/filled.
- **Button Status & Guard Check**: Updated the **Download Report Excel** button `:disabled` state to include `|| !hasActiveFilter`, ensuring the button is disabled when no filters are set. Added a guard check inside `handleDownloadExcel()` showing a warning message if triggered without active filters.
- **Verification**: Executed `npm run type-check` (`vue-tsc --noEmit`), passing cleanly with 0 errors.

---

## \[Completed\] Service Worker Request Deduplication (`vue/src/strategies/dynamicnetworkcache.ts` & `vue/src/strategies/dynamiclownetworkcache.ts`)

### Problem & Root Cause:
When network conditions are slow, if a client cancels/aborts a fetch (e.g. component unmounting or user navigation/retry), the Service Worker strategy may still have an orphaned `fetch(request)` in-flight. When the client re-requests the same resource, the SW launched a duplicate network `fetch()`, causing redundant network traffic and wasted bandwidth.

### Summary & Changes Made:
- **`fetchDeduplicated` Method**: Implemented `protected fetchDeduplicated(request: Request, handler: StrategyHandler): Promise<Response>` on `DynamicNetworkCacheStrategy` in [dynamicnetworkcache.ts](./vue/src/strategies/dynamicnetworkcache.ts).
- **In-Flight Request Map**: Uses `protected inFlightRequests: Map<string, Promise<Response>>` to track active in-flight GET/HEAD requests by `${request.method}:${request.url}`.
- **Promise Re-use & Response Cloning**:
  - When an in-flight fetch promise exists for a request key, returns `existingPromise.then((response) => response.clone())`, allowing concurrent callers to receive independent response stream clones.
  - Automatically cleans up key entries upon settlement via `.finally(() => { this.inFlightRequests.delete(key); })`.
  - Non-GET/HEAD requests fall back directly to standard `handler.fetch(request)`.
- **`DynamicLowNetworkCacheStrategy` Inheritance**: Updated [dynamiclownetworkcache.ts](./vue/src/strategies/dynamiclownetworkcache.ts) to utilize `this.fetchDeduplicated(request, handler)` for both background updates and cache-miss network fetches, automatically extending deduplication benefits to low network strategy mode.
- **Verification**: Verified via `npm run type-check` (`vue-tsc --noEmit`), passing cleanly with 0 type errors.

---

## \[Completed\] `DynamicLowNetworkCacheStrategy` Implementation (`vue/src/strategies/dynamicnetworkcache.ts`)

### Summary & Changes Made:
- **`DynamicLowNetworkCacheStrategy`**: Created class inheriting `DynamicNetworkCacheStrategy` in [dynamicnetworkcache.ts](./vue/src/strategies/dynamicnetworkcache.ts).
- **Network Detection**: Automatically checks `navigator.connection` (`effectiveType <= 3g` e.g. `'slow-2g'`, `'2g'`, `'3g'` OR `saveData === true`). Falls back to standard `DynamicNetworkCacheStrategy` when network is fast.
- **Cache-First Priority & Configurable 30m Debounce**:
  - When low network mode is active: if cache is present, returns `cachedResponse` immediately.
  - Background fetch for the same URL is throttled/debounced with a configurable window (default `lowNetworkDebounceMs = 30 * 60 * 1000`, 30 minutes).
  - If cache is absent, fetches from network, updates cache (`handler.cachePut`), and returns the network response.
- **Verification**: Executed `npm run type-check` (`vue-tsc --noEmit`), completing cleanly with 0 type errors.

---


## \[Completed\] HTTPS Scheme Detection & Port 443 Logging (`gin/internal/api/router.go`)

### Summary & Changes Made:
- **`ForwardedHeaderMiddleware`**: Implemented `ForwardedHeaderMiddleware` in [router.go](./gin/internal/api/router.go) to inspect `X-Forwarded-Proto`, `X-Forwarded-Scheme`, `X-Forwarded-Ssl`, and `Front-End-Https` headers sent by reverse proxies (such as IIS HttpPlatformHandler).
- **HTTPS & Port 443 Forced Logging**: Updated `FilterSuccessLogMiddleware` to detect if `Request.URL.Scheme` or `X-Forwarded-Proto` is `"https"` (or `X-Forwarded-Port` is `"443"`), forcing `c.Request.URL.Scheme = "https"` and tagging log lines with `[HTTPS:443]` (or `[HTTPS:443-PANIC]`).
- **Verification**: Built and verified Go binary via `docker compose run --rm gin-build`, passing with exit code 0.

---

## \[Completed\] HTTP Referer Logging in `FilterSuccessLogMiddleware` (`gin/internal/api/router.go`)

### Summary & Changes Made:
- **HTTP Referer Logging**: Updated `FilterSuccessLogMiddleware` in [router.go](./gin/internal/api/router.go) to capture `c.Request.Referer()`. When the `Referer` header is present, it is automatically appended to `log.Printf` output (`| Ref: <referer_url>`) for both `[HTTP]` status >= 400 and `[PANIC]` log events.
- **Verification**: Compiled and verified Go server binary via `docker compose run --rm gin-build`, passing cleanly with exit code 0.

---

## \[Completed\] IIS HttpPlatformHandler & Client IP Extraction (`gin/dist/web.config`, `gin/internal/api/middleware/ratelimit.go`, `gin/internal/api/router.go`)

### Summary & Changes Made:
- **`middleware.GetClientIP(c)` Helper**: Created a robust `GetClientIP` function in [ratelimit.go](./gin/internal/api/middleware/ratelimit.go) that checks incoming proxy headers (`X-Forwarded-For`, `X-Real-IP`, `X-ARR-ClientIP`, `X-Original-For`) to resolve the client's actual remote IP, filtering out loopback addresses (`127.0.0.1`, `::1`) and stripping ports via `net.SplitHostPort`.
- **Rate Limiter & Middleware Integration**: Updated `LoginRateLimiter.GetKey` and `FilterSuccessLogMiddleware` in [router.go](./gin/internal/api/router.go) to use `middleware.GetClientIP(c)`.
- **Gin Proxy Trust Configuration**: Configured `r.ForwardedByClientIP = true` and `r.SetTrustedProxies(nil)` in `NewRouter` to trust proxy headers from upstream reverse proxies without fallback loopback locking.
- **Verification**: Built and verified Go binary via `docker compose run --rm gin-build`, passing with exit code 0.

---

## \[Completed\] UmatReportList UI Fixes & Clean Pagination Layout (`vue/src/views/report/UmatReportList.vue`)

### Summary & Changes Made:
- **Card Border & Filter Layout Fix**:
  - Removed redundant `border: 1px solid #e4e7ed` overrides on `.filter-card` and `.table-card` which were creating double/bolder borders on top of Element Plus default card styling. Updated styles to use Element Plus CSS variables (`var(--el-border-color-lighter)`, `var(--el-text-color-primary)`).
  - Grouped `Usia Dari` and `Usia Sampai` into a unified **Rentang Usia (Tahun)** control range container (`Dari` \[input\] `s/d` \[input\] `Sampai`).
  - Added dynamic validation rules (`:max="filters.usia_sampai"` for Dari and `:min="filters.usia_dari"` for Sampai) plus `onUsiaDariChange` / `onUsiaSampaiChange` handlers guaranteeing `usia_sampai >= usia_dari`.
  - Rebalanced Row 3 grid column spans (`Rentang Usia`, `Lulus SD3`, `Vege / Qing Kou`) to `:md="8"` (33.3% width each), filling the row cleanly.
  - Wrapped filter form in `<el-form label-position="top">` to ensure labels sit cleanly above inputs without squeezing controls horizontally.
- **Usia Filter Initialization & Execution Fix**: Changed `usia_dari` and `usia_sampai` defaults in `filters` reactive state from `0` to `null as number | null` (and added `placeholder="Dari"` / `placeholder="Sampai"`). Updated `fetchData()` and `handleDownloadExcel()` to pass `filters.usia_dari ?? undefined` and `filters.usia_sampai ?? undefined`. This prevents the age filter from automatically triggering for age 0 on initial page load or after clicking Reset Filter.
- **Pagination Layout String Cleanup**: Fixed redundant pagination layout expression string in [UmatReportList.vue](./vue/src/views/report/UmatReportList.vue), [SxyReportList.vue](./vue/src/views/report/SxyReportList.vue), and [DonasiSxyList.vue](./vue/src/views/transaction/DonasiSxyList.vue) to cleanly evaluate `isDesktop ? 'total, sizes, prev, pager, next, jumper' : 'total, sizes, prev, pager, next'`.
- **Verification**: Verified via `npm run type-check` (`vue-tsc --noEmit`) and production build (`npm run build`), passing with 0 errors.

---

## \[Completed\] Login Rate Limiter with Redis & Standard Headers (`gin/internal/api/router.go` & `vue/src/views/Login.vue`)

### Summary & Changes Made:
- **Rate Limiting Engine**: Implemented `LoginRateLimiter` in [ratelimit.go](./gin/internal/api/middleware/ratelimit.go) using `github.com/ulule/limiter/v3` with Redis store default (`REDIS_ADDR`) and memory store fallback if Redis is unconfigured or unreachable.
- **Rule**: 3 failed login attempts per 5 minutes per Client IP.
- **Standard Headers**: Configured headers on rate limited response:
  - `Retry-After`: Seconds remaining in lockout window.
  - `X-Retry-After`: Seconds remaining in lockout window.
  - `X-RateLimit-Limit`: `3`.
  - `X-RateLimit-Remaining`: Remaining allowed attempts.
  - `X-RateLimit-Reset`: Unix timestamp when window resets.
- **Lock Out Behavior**:
  - Automatically checks `limiter.Peek()` before processing login. If `limCtx.Reached` (count >= 3), returns `HTTP 429 Too Many Requests` with retry headers and error message `"Terlalu banyak percobaan login yang salah. Silakan coba lagi dalam 5 menit."`.
  - On login failure (`401`), increments failed attempt counter (`limiter.Increment()`).
  - On successful login (`200`), resets failed attempt counter (`limiter.Reset()`) so legitimate users are not locked out later.
- **Docker Compose**: Added `redis` service (`redis:7-alpine`) and set `REDIS_ADDR=redis:6379` for Gin container in [docker-compose.yml](./docker-compose.yml).
- **Frontend Vue**: Updated [Login.vue](./vue/src/views/Login.vue) and [auth.ts](./vue/src/stores/auth.ts) to handle rate limit responses. When 429 / `retry_after` occurs:
  - Form inputs and submit button are disabled (`:disabled="loading || isLockedOut"`).
  - A live countdown timer (`formattedCountdown`) is displayed.
  - Automatically re-enables login once the countdown reaches 0 seconds.
- **Verification**: Tested with unit test `TestLoginRateLimiter` in [router_test.go](./gin/internal/api/router_test.go), passing 100%.

---

## \[Completed\] Fixed Dev Service Worker ES Module Import Syntax Error (`vue/vite.config.ts`)

### Problem & Root Cause:
When using `VitePWA` with `devOptions.type = 'classic'` and `strategies: 'injectManifest'`, the browser registers the development service worker as a classic script (`type: 'classic'`). However, `dev-sw.js` generated by Vite contains top-level ES `import` statements (e.g. `import { precacheAndRoute ... } from 'workbox-precaching'`), causing the browser to throw `Uncaught SyntaxError: Cannot use import statement outside a module (at dev-sw.js?dev-sw:1:1)`.

### Solution:
Updated `devOptions.type` from `'classic'` to `'module'` in [vite.config.ts](./vue/vite.config.ts).
This tells `vite-plugin-pwa` to register the service worker as an ES module (`{ type: 'module' }`), allowing ES module `import` statements in `dev-sw.js` during development while building cleanly into bundled `sw.js` for production.

---

## \[Completed\] Added Debug Logging & Improved Route Matching (`vue/src/sw.ts` & `vue/src/strategies/dynamicnetworkcache.ts`)

### Changes Made:
- Added explicit console logging in [sw.ts](./vue/src/sw.ts) for matched routes (`[SW Route] Matched /v1/ or /api/: ...`).
- Added console logging in [dynamicnetworkcache.ts](./vue/src/strategies/dynamicnetworkcache.ts) (`[SW DynamicNetworkCacheStrategy] Intercepting: GET ...`).
- Expanded route matcher in `sw.ts` to cover both `/v1/` and `/api/` endpoints cleanly under `DynamicNetworkCacheStrategy`.

---

## \[Completed\] Registered `DynamicNetworkCacheStrategy` for Default GET `/v1/` Routes (`vue/src/sw.ts`)

### Changes Made:
- Updated `vue/src/sw.ts` default route matcher for GET `/v1/` endpoints (such as `/v1/penggalang-dana?page=3&limit=10`) from standard Workbox `NetworkFirst` to `DynamicNetworkCacheStrategy`.
- Set default `timeoutMs: 500` and `debounceMs: 5000` for general `/v1/` API endpoints under `api-cache`.
- Verified type safety via `npm run type-check` (`vue-tsc --noEmit`), passing with 0 errors.

---

## \[Completed\] TypeScript `any` Type Reduction Across `vue/src` (`vue/src`)

### Changes Made:
- **`lookupCache.ts`**: Replaced `any` types in `cachedFetchLookup` with generic types `<T = unknown, P = Record<string, unknown>>` and typed `CacheEntry<T>`.
- **`useQuickStreamSpeedTest.ts`**: Replaced `catch (error: any)` with `catch (error: unknown)` and `error instanceof Error` type guard.
- **`auth.ts`**: Replaced `[key: string]: any` with `[key: string]: unknown` in `User` interface, and replaced `catch (error: any)` with `catch (error: unknown)` using `axios.isAxiosError(error)`.
- **`KelasForm.vue` & `UmatForm.vue`**: Removed `as any` from `trimTargetFields` helper functions by typing `trimmed` as `Record<string, unknown>`.
- **`ChangePasswordModal.vue`, `Login.vue`, `ChangePasswordView.vue`**: Replaced `catch (error: any)` / `catch (err: any)` with `catch (error: unknown)` and proper error messaging.
- **`KelasDonasiBarangTable.vue`**: Replaced `catch (err: any)` with `catch (err: unknown)` using `axios.isAxiosError(err)`.
- **`dynamicnetworkcache.ts`**: Imported `StrategyHandler` from `workbox-strategies` to strongly type `handler` parameters and return values.
- **Verification**: Executed `npm run type-check` (`vue-tsc --noEmit`), completing with 0 errors.

---

## \[Completed\] Workbox `StrategyHandler` Integration (`vue/src/strategies/dynamicnetworkcache.ts`)

### Changes Made:
- Imported `StrategyHandler` directly from `workbox-strategies`.
- Replaced `any` parameter types on `handler: StrategyHandler` across `_handle` and `triggerDebouncedBackgroundUpdate`.
- `handler.cachePut(key, response)` now strongly types its return value as `Promise<boolean>` instead of `any`.
- Verified type safety via `npm run type-check` (`vue-tsc --noEmit`), passing with 0 errors.

---

## \[Completed\] Non-Blocking Async `handler.cachePut` via `handler.waitUntil` (`vue/src/strategies/dynamicnetworkcache.ts`)

### Changes Made:
Updated `DynamicNetworkCacheStrategy` to use `handler.waitUntil` for `cachePut`:
- `handler.cachePut(request, networkResponse.clone())` is executed inside `.then()` without `await`, returning `Response` immediately to Axios/main thread with zero delay.
- Passed the `putPromise` to `handler.waitUntil(putPromise)` to extend Service Worker lifecycle in background, guaranteeing IndexedDB cache write completes safely even if the main thread receives the response earlier.

---

## \[Completed\] Non-Blocking Parallel `DynamicNetworkCacheStrategy` with Fallback Race (`vue/src/strategies/dynamicnetworkcache.ts`)

### Changes Made:
Refactored `DynamicNetworkCacheStrategy._handle` to include a secondary `Promise.race` fallback on `timeoutMs`:
1. **Concurrent Initialization**: Starts `cachePromise` (`handler.cacheMatch`) and `networkPromise` (`handler.fetch`) concurrently at step 0 without blocking.
2. **Fast Network Path (< `timeoutMs`)**: Races `networkPromise` against `timeoutMs`. If `networkPromise` finishes first, clears timer and returns `networkResponse` immediately. Background `cachePut` is performed non-blockingly via `.then(...)`.
3. **Fallback Race Path (Network Not Finished Before `timeoutMs`)**:
   - Races `cachePromise` against `networkPromise`.
   - **If `cachePromise` wins**:
     - Checks if cached response exists (`if (cachedResponse)`). If present, returns `cachedResponse` immediately while `networkPromise` continues renewing the cache in the background.
     - If cache does not exist, awaits `networkPromise` to complete and add to cache.
   - **If `networkPromise` wins**: Returns `networkResponse` directly.

---

## \[Completed\] Form Submission Page Disabling & Finally Block Enforcement

### Changes Made:
Updated all CREATE/UPDATE forms and modal dialogs to disable the form/page during submit and enforce `try ... catch ... finally` state management for `submitting.value`:

1. **`KelasDonasiBarangTable.vue`**:
   - Added `v-loading="submitting"` to form container.
   - Added `:disabled="submitting"` to form buttons.
   - Updated `submitForm` to set `submitting.value = true` before async validation/request and reset in `finally` block.

2. **`KelasDonasiTable.vue`**:
   - Added `v-loading="submitting"` to form container.
   - Added `:disabled="submitting"` to form buttons.
   - Updated `submitForm` to set `submitting.value = true` before async validation/request and reset in `finally` block.

3. **`KelasForm.vue`**:
   - Added `v-loading="submitting"` to `el-card` form container.
   - Added `:disabled="submitting"` to form buttons.

4. **`KelasPengabdiTable.vue`**:
   - Added `v-loading="submitting"` to form container.
   - Added `:disabled="submitting"` to form buttons.
   - Updated `submitForm` to set `submitting.value = true` before async validation/request and reset in `finally` block.

5. **`KelasPesertaTable.vue`**:
   - Added `v-loading="submitting"` to form container.
   - Added `:disabled="submitting"` to form buttons.
   - Updated `submitForm` to set `submitting.value = true` before async validation/request and reset in `finally` block.

6. **`KelasTopikTable.vue`**:
   - Added `v-loading="submitting"` to form container.
   - Added `:disabled="submitting"` to form buttons.
   - Updated `submitForm` to set `submitting.value = true` before async validation/request and reset in `finally` block.

7. **`UmatForm.vue`**:
   - Added `v-loading="submitting"` to `el-form` container.
   - Added `:disabled="submitting"` to form action buttons.

---

## \[Completed\] Network Speed Monitoring & Axios Metadata Type Fix

### Changes Made:
1. **Axios Request Config Metadata Type Fix** (`vue/src/api/client.ts`):
   - Added module augmentation for `axios` to declare the `metadata` property on `InternalAxiosRequestConfig` and `AxiosRequestConfig`.
   - Updated `clearRequestTimer` signature to accept typed Axios configuration.

2. **Shared Composable State & 5s Throttle** (`vue/src/api/useQuickStreamSpeedTest.ts`):
   - Lifted state (`speedMbps`, `isTesting`, `lastTestCompletedAt`) to module scope so it is shared across all caller components.
   - Exported `canRunTest()` helper checking `!isTesting.value && (Date.now() - lastTestTime >= THROTTLE_MS)`.
   - Enforced 5-second throttling on `measureSpeedInOneSecond()` calls to prevent duplicate/overlapping tests.
   - Added `lastTestCompletedAt` timestamp state to signal test completions.

3. **Axios Speed Test Trigger Throttling** (`vue/src/api/client.ts`):
   - Updated the 1-second timeout callback in the Axios request interceptor to evaluate `canRunTest()` BEFORE logging warning or calling `measureSpeedInOneSecond()`.
   - Prevents multiple slow requests taking >= 1s from triggering redundant download tests or warning logs within any 5-second window.

4. **Notification on Low Speed (< 0.8 Mbps)** (`vue/src/layouts/MainLayout.vue`):
   - Added reactive `watch` on `lastTestCompletedAt`.
   - Triggered `ElNotification.warning` when `speedMbps < 0.8`.
   - Throttled notification triggers with a 5-second cooldown window (`NOTIFY_THROTTLE_MS = 5000`).

---

## \[Completed\] k6 Load Test Script (`vue/pentest.js`)

### Changes Made:
Updated `vue/pentest.js` ([pentest.js](./vue/pentest.js)) to focus strictly on performance load testing matching Postman collection specs:
1. **Authentication Flow**: Login via `POST /login` and extract JWT Bearer token.
2. **GET Umat Batch Parallel Pagination**: Hits page 1 first to read `meta.total` and determine `totalPages = Math.ceil(meta.total / limit)`. Then fires requests for pages 2 through `totalPages` concurrently in parallel using `http.batch()` without waiting sequentially for each page response.
3. **Automatic Configuration & Docker Resolution**: Reads `.env` using k6's `open()` helper and resolves backend URLs (`http://gin:8080`) when run inside Docker.

---

## \[Completed\] Gin Backend & IIS HttpPlatformHandler Concurrency Optimizations

### Changes Made:
1. **GORM Connection Pool & Performance Flags** (`gin/internal/database/db.go`):
   - Added `SkipDefaultTransaction: true` to bypass automatic transaction wrapping on read queries.
   - Added `PrepareStmt: true` to cache compiled SQL queries.
   - Increased connection pool limits to `SetMaxOpenConns(500)` and `SetMaxIdleConns(100)` to eliminate TCP handshake overhead during load spikes.

2. **Middleware Disk I/O Bottleneck Fix** (`gin/internal/api/router.go`):
   - Disabled synchronous file/stdout logger `gin.Logger()` in `release` mode to prevent disk write locking under high VUs.

3. **IIS HttpPlatformHandler Configuration** (`gin/dist/web.config`):
   - Added `requestQueueLimit="10000"` and `requestTimeout="00:05:00"` to prevent request dropping/socket exhaustion in IIS.
   - Added Go runtime environment variables: `GIN_MODE=release`, `GOGC=100`, `GOMEMLIMIT=2GiB`.

4. **SQL Server XML Memory Exhaustion Fix & Transient Retry** (`gin/internal/service/auth_service.go`):
   - Implemented thread-safe `sync.Map` in-memory menu caching with a 5-minute TTL for `SP_Login_Create_Xml` and sub-menu tree queries.
   - Added transient TCP connection retry policy (`executeWithRetry`) in `AuthService.Login` to withstand initial TCP SYN backlog refusal bursts under 500 VUs.
   - Added asynchronous connection pool pre-warming (50 warm sockets at startup) in `gin/internal/database/db.go`.

---

## \[Completed\] Redundant `kelasApi.getKelasById` Call Prevention

### Changes Made:
1. **`KelasPesertaTable.vue`**:
   - Updated `fetchKelasDetail()` to skip calling `kelasApi.getKelasById` if `props.startDate` and `props.endDate` are already provided by the parent component (`KelasEdit.vue` / `KelasView.vue`).
   - Removed `{ immediate: true }` from the `props.kelasId` watcher to prevent redundant fetch during component setup/mount.
   - Updated `onMounted` and `watch` handlers to only call `fetchKelasDetail()` when dates are missing.

2. **`KelasPengabdiTable.vue`**:
   - Updated `fetchKelasDetail()` to skip calling `kelasApi.getKelasById` if `props.startDate` and `props.endDate` are already provided by the parent component (`KelasEdit.vue` / `KelasView.vue`).
   - Removed `{ immediate: true }` from the `props.kelasId` watcher to prevent duplicate fetches.
   - Updated `onMounted` and `watch` handlers to only call `fetchKelasDetail()` when dates are missing.

---

## \[Completed\] Lookup API Request Deduplication & Lazy Mount Optimization (`LookupSelect.vue` & `lookupCache.ts`)

### Changes Made:
1. **Lookup In-Flight Request Deduplication & 30s Caching** (`vue/src/utils/lookupCache.ts`):
   - Created `cachedFetchLookup` utility using `WeakMap<Function, Map<string, CacheEntry>>` to automatically deduplicate simultaneous in-flight lookup requests for identical API functions & parameters.
   - Implemented a 30-second TTL cache for reference lookups (`getFotangLookup`, `getLookupPendidikan`, `getLookupPekerjaan`, etc.).

2. **Lazy Mount Loading & Cache Integration** (`vue/src/components/common/LookupSelect.vue`):
   - Integrated `cachedFetchLookup` into `LookupSelect.vue` to resolve multiple `LookupSelect` instances (e.g. 4 Fotang selectors in `UmatForm.vue`) into a single network call.
   - Updated `onMounted` hook to skip eager network requests when `modelValue` is empty and no initial option is provided, allowing dropdowns to load lazily on user interaction (`onVisibleChange`).

---

## \[Completed\] String Lookup Field Trimming (`UmatForm.vue` & `KelasForm.vue`)

### Changes Made:
1. **`UmatForm.vue`**:
   - Implemented `trimTargetFields` helper for target lookup string fields (`kelas_khusus`, `kelas_umum`, `fotang_chiutao`, `fotang_aktif`, `tcs`, `waktu_chiutao_mandarin`, `pendidikan`, `pekerjaan`, `jenis_kelamin`, `tempat_sd2`, `tempat_sd3`, `tim_kerja`, `posisi`, `status_umat`).
   - Applied trimming to `initialData` on watcher initialization to prevent trailing whitespace mismatch in `LookupSelect`.
   - Applied trimming to payload fields on `handleSubmit` before emitting `submit` event.

2. **`KelasForm.vue`**:
   - Implemented `trimTargetFields` helper for string lookup and text fields (`kode_kelas`, `kode_fotang`, `level`, `lokasi`, `pic`, `keterangan`, `mc1`..`mc5`).
   - Applied trimming to `initialData` on watcher initialization and payload fields on `handleSubmit`.

---

## \[Completed\] Validation Error HTTP Status Code Fix (HTTP 400 Bad Request)

### Changes Made:
1. **Router Error Handler Refactoring** (`gin/internal/api/router.go`):
   - Refactored `respondError(c, err)` and `respondValidationError(c, err)` to properly recognize `*service.ValidationError` via `errors.As(err, &vErr)` and check for validation keywords (`"tidak valid"`, `"is required"`, `"Validation failed"`, `"not found in lookup"`, `"invalid ID format"`).
   - Replaced all fragile string matching (`strings.Contains(err.Error(), "Validation failed")`) and hardcoded 404 error responses across all `Create`, `Update`, and `Delete` endpoints in `router.go` with unified `respondError(c, err)`.
   - Guaranteed that lookup and field validation errors (e.g. `fotang_aktif: field B_FOTANG nilai '31' tidak valid`) return **HTTP 400 Bad Request** with structured JSON payload containing `"error"` and `"details"` map.

2. **Unit Test Verification** (`gin/internal/api/router_test.go`):
   - Added unit test `TestValidationErrorReturns400` verifying that `service.ValidationError` produces HTTP status 400 Bad Request with error detail content. Verified that all Go unit tests pass (`go test ./...`).

---

## \[Completed\] JWT `userID` `float64` Type Assertion Fix in `UmatService`

### Changes Made:
1. **Unsafe Type Assertion Removal** (`gin/internal/service/umat_service.go`):
   - Replaced unsafe `int32(c.MustGet("userID").(int))` type assertions on lines 226, 328, and 358 with `getUserID(c)`.
   - `c.Set("userID", claims["sub"])` in `AuthMiddleware` stores `userID` from JWT claims as a `float64` (Standard JSON unmarshaling type in `golang-jwt`), causing `.(int)` to panic with `interface conversion: interface {} is float64, not int`.
   - Using `getUserID(c)` safely handles `float64`, `int`, `int32`, `string` or fallback `1` without panicking.

---

## \[Completed\] Seamless In-Place Gzip Compression & IIS `web.config` Setup

### Changes Made:
1. **In-Place Gzip Compression Plugin** (`vue/vite.config.ts`):
   - Created `compressInPlacePlugin` in `vite.config.ts` that compresses built static assets (`.js`, `.css`, images, `.svg`, `.html`, `.json`) in-place during `closeBundle()`.
   - All files retain their original filenames and extensions (no `.gz` postfix anywhere on disk or in URLs).
   - Confirmed `npm run build:client` reduces main JS bundle from 1.26MB to 398KB and CSS bundle from 363KB to 48KB directly on disk.

2. **IIS Seamless Header Configuration** (`vue/public/web.config`):
   - Configured an outbound rewrite rule (`Set Content-Encoding gzip for static assets`) in `web.config` to attach `Content-Encoding: gzip` HTTP response header for static files (`.js`, `.css`, images, `.svg`, `.html`, `.json`, etc.).

---

## \[Completed\] Persistent Offline Warning Notification (`vue/src/layouts/MainLayout.vue`)

### Changes Made:
- Imported `useOnline` from `@vueuse/core` in [MainLayout.vue](./vue/src/layouts/MainLayout.vue#L172).
- Added reactive watcher on `isOnline`:
  - **When Offline (`isOnline === false`)**: Displays an `ElNotification.warning` with `duration: 0` (persistent notification that stays open until network connectivity returns).
  - **When Online (`isOnline === true`)**: Automatically closes the active offline notification handle via `.close()`.
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## \[Completed\] PWA Asset Update Detection & Page Change Application (`vue/src/sw.ts` & `vue/src/utils/pwaUpdate.ts`)

### Changes Made:
1. **Service Worker Update Broadcast & Skip Waiting** ([sw.ts](./vue/src/sw.ts#L25-L45)):
   - Added message listener for `SKIP_WAITING` event to immediately activate new Service Workers.
   - Added `activate` event handler broadcasting `{ type: 'SW_UPDATED' }` via `postMessage` to all active window clients when new cached assets are activated.

2. **Router Navigation Update Guard** ([pwaUpdate.ts](./vue/src/utils/pwaUpdate.ts)):
   - Created `initPwaUpdate(router)` utility initialized in [main.ts](./vue/src/main.ts#L25).
   - Listens to Workbox `onNeedRefresh` and SW `SW_UPDATED` messages to mark `hasPendingUpdate = true`.
   - On page navigation (`router.beforeEach`), if a SW update is pending, triggers `updateSW(true)` (reloading seamlessly to apply updated assets on page change).
   - Automatically checks `registration.update()` on every route change (`router.afterEach`) to detect newly deployed builds.
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## \[Completed\] Umat Image Processing, Cross-Platform Validation & `UmatFoto` One-To-One Record ([umat_service.go](./gin/internal/service/umat_service.go))

### Changes Made:
1. **Optional Image Header Processing** ([umat_service.go](./gin/internal/service/umat_service.go)):
   - Updated `UmatService.Create` and `UmatService.Update` signatures to accept `fileHeader *multipart.FileHeader` (allowing `nil`).
2. **Image Validation Engine**:
   - Max file size limit: 8MB (`8 * 1024 * 1024` bytes).
   - Image format validation via MIME detection & `image.DecodeConfig` (JPEG, PNG, GIF).
   - Resolution validation: Width (200px - 2000px), Height (400px - 4000px).
   - Aspect ratio validation: 2:4 ratio (`width / height == 0.5` with tolerance `±0.03`).
   - Structured error format on validation failure: `{"error": "Invalid Input", "details": {"foto": [...]}}` with HTTP 400 status.
3. **Cross-Platform Filename Sanitization**:
   - Added `sanitizeFilename(name string) string` helper ensuring filenames are safe for Linux and Windows (sanitizes `< > : " / \ | ? * \x00-\x1f`, trims spaces, normalizes path separators, and prevents Windows reserved filenames like `CON`, `PRN`, `AUX`, `NUL`).
4. **Configurable Storage & `UmatFoto` Database Recording**:
   - Storage directory configurable via `UPLOAD_DIR` env variable (default `./uploads`).
   - Saves file to `./uploads/{Umat.Id}/{filename}` (`DocPath` = `{Umat.Id}`, `DocFile` = `{DocPath}/{filename}`).
   - Upserts record in `T_BUS_UMAT_FOTO` (`UmatFoto`) with `FileID` generated via `s.db.Raw("EXEC SP_APP_GenerateId ?, ?, ?", "FILE", nowStr, 1)`.
- Verified compilation with `go build ./...` and unit tests with `go test -run TestSanitizeFilename ./internal/service/...`.

---

## \[Completed\] Frontend Photo Capture, 2:4 Cropper & Postman Multipart Collection Update

### Changes Made:
1. **Postman Collection Update** ([apps-gin.postman_collection.json](./gin/postman/apps-gin.postman_collection.json)):
   - Added `Umat - Create (Multipart with Photo)` and `Umat - Update (Multipart with Photo)` endpoints with `formdata` containing `"data"` (JSON payload) and `"foto"` (file header).
2. **Vue Multipart API Support** ([umat.ts](./vue/src/api/umat.ts)):
   - Updated `umatApi.createUmat` and `umatApi.updateUmat` to accept optional `photoFile?: File | Blob | null`.
   - Sends `FormData` with `multipart/form-data` content-type header when `photoFile` is present.
3. **Photo Capture & 2:4 Aspect Ratio Cropper** ([UmatPhotoUpload.vue](./vue/src/components/umat/UmatPhotoUpload.vue)):
   - Installed `vue-advanced-cropper` package.
   - HTML5 File Input with `accept="image/jpeg,image/png,image/webp,image/gif"` and `capture="environment"` for mobile camera and gallery picking.
   - Custom MediaDevices API live camera stream modal (`navigator.mediaDevices.getUserMedia`) for desktop/mobile browser video feed capture.
   - Interactive cropper modal enforcing strict **2:4 aspect ratio** (`aspectRatio: 2/4`) and output resolution.
4. **Form Integration**:
   - Embedded `UmatPhotoUpload` into Section 1 of [UmatForm.vue](./vue/src/components/umat/UmatForm.vue).
   - Updated [UmatCreateList.vue](./vue/src/views/master-data/UmatCreateList.vue) and [UmatEditList.vue](./vue/src/views/master-data/UmatEditList.vue) to forward `photoFile` blob to `createUmat` and `updateUmat`.
5. **Responsive Layout Tuning**:
   - Updated `.photo-preview-container` in [UmatPhotoUpload.vue](./vue/src/components/umat/UmatPhotoUpload.vue) to `width: 100%`, `max-width: 210px`, and `aspect-ratio: 3 / 4` for responsive scaling across devices.
- Verified with `npx vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## \[Completed\] Lazy Rendering Form Cards with `useIntersectionObserver` ([UmatForm.vue](./vue/src/components/umat/UmatForm.vue))

### Changes Made:
1. **Intersection Observer Integration** ([UmatForm.vue](./vue/src/components/umat/UmatForm.vue#L571-L615)):
   - Imported `useIntersectionObserver` from `@vueuse/core`.
   - Created template ref containers (`cardRef2`, `cardRef3`, `cardRef4`) and boolean visibility flags (`isCard2Visible`, `isCard3Visible`, `isCard4Visible`).
   - Attached `useIntersectionObserver` with a `200px` root margin for smooth pre-rendering before entering the viewport, disconnecting each observer (`stop()`) once visible.
2. **Lazy Card Templates**:
   - Wrapped Card 2 (`Informasi Ciu Tao`), Card 3 (`Kontak & Fotang`), and Card 4 (`Kelas, Sidang Dharma, Dll`) with `v-if="isCardXVisible"`.
   - Displayed an `<el-card>` skeleton placeholder when cards are not yet visible to prevent layout shift.
3. **Form Submit & Edit Fallbacks**:
   - Added `revealAllCards()` helper ensuring all card sections are immediately rendered on form submission (`handleSubmit`), initial data load (`watch props.initialData`), or validation errors (`watch props.fieldErrors`).
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## \[Completed\] Accordion Header Action Margin Right Spacing

### Changes Made:
- Added `margin-right: 0.75rem;` to `.header-actions` across all 5 Kelas accordion table components to maintain clean spacing between action/refresh buttons and the Element Plus collapse expand arrow:
  1. [KelasDonasiBarangTable.vue](./vue/src/components/kelas/KelasDonasiBarangTable.vue#L348-L353)
  2. [KelasDonasiTable.vue](./vue/src/components/kelas/KelasDonasiTable.vue#L357-L362)
  3. [KelasPengabdiTable.vue](./vue/src/components/kelas/KelasPengabdiTable.vue#L711-L716)
  4. [KelasPesertaTable.vue](./vue/src/components/kelas/KelasPesertaTable.vue#L668-L673)
  5. [KelasTopikTable.vue](./vue/src/components/kelas/KelasTopikTable.vue#L603-L608)
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## \[Completed\] Conditional Rendering of Accordion `<el-tag>` Counter on Expanded State

### Changes Made:
- Applied `v-if="isExpanded"` to `<el-tag>` total counter badges in accordion headers across all 5 Kelas table components so the counter is hidden when folded and displayed only when expanded:
  1. [KelasDonasiBarangTable.vue](./vue/src/components/kelas/KelasDonasiBarangTable.vue#L12)
  2. [KelasDonasiTable.vue](./vue/src/components/kelas/KelasDonasiTable.vue#L12)
  3. [KelasPengabdiTable.vue](./vue/src/components/kelas/KelasPengabdiTable.vue#L12)
  4. [KelasPesertaTable.vue](./vue/src/components/kelas/KelasPesertaTable.vue#L12)
  5. [KelasTopikTable.vue](./vue/src/components/kelas/KelasTopikTable.vue#L12)
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## \[Completed\] Dynamic Theme-Aware Browser Autofill Styling ([Login.vue](./vue/src/views/Login.vue) & [style.css](./vue/src/style.css))

### Changes Made:
1. **Dynamic CSS Variables for Autofill**:
   - Replaced hardcoded `#141414` background and `#ffffff` text color in `-webkit-autofill` rules with Element Plus CSS variables:
     - `-webkit-box-shadow: 0 0 0 1000px var(--el-input-bg-color, var(--el-fill-color-blank)) inset !important;`
     - `-webkit-text-fill-color: var(--el-text-color-primary) !important;`
     - `caret-color: var(--el-text-color-primary) !important;`
2. **Vue Scoped Selector Deep Penetration & Global Fallback**:
   - Used `:deep(input:-webkit-autofill)` in [Login.vue](./vue/src/views/Login.vue#L216-L229) to pierce Element Plus shadow DOM structure (`.el-input__inner`).
   - Added global autofill rule to [style.css](./vue/src/style.css#L31-L40) ensuring all form input autofills across Light and Dark themes dynamically adapt without hardcoded colors.
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## \[Completed\] Global CORS Allowed Headers Update ([router.go](./gin/internal/api/router.go))

### Changes Made:
- Updated `corsConfig.AllowHeaders` in [router.go](./gin/internal/api/router.go#L176-L181) to globally include `User-Agent` and `X-Requested-With` headers alongside `Origin`, `Content-Length`, `Content-Type`, and `Authorization`:
  ```go
  corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "User-Agent", "X-Requested-With"}
  ```
- Verified compilation cleanly with `go build ./...`.




- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## \[Completed\] Global CORS Allowed Headers Update ([router.go](./gin/internal/api/router.go))

### Changes Made:
- Updated `corsConfig.AllowHeaders` in [router.go](./gin/internal/api/router.go#L176-L181) to globally include `User-Agent` and `X-Requested-With` headers alongside `Origin`, `Content-Length`, `Content-Type`, and `Authorization`:
  ```go
  corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "User-Agent", "X-Requested-With"}
  ```
- Verified compilation cleanly with `go build ./...`.

---

## \[Completed\] Programmatic `ElNotification` & `ElMessage` Global CSS Styles Import ([main.ts](./vue/src/main.ts))

### Changes Made:
1. **Global CSS Imports for Programmatic APIs**:
   - Added explicit imports for Element Plus programmatic feedback components in [main.ts](./vue/src/main.ts#L6-L9):
     - `import 'element-plus/es/components/notification/style/css'`
     - `import 'element-plus/es/components/message/style/css'`
     - `import 'element-plus/es/components/message-box/style/css'`
     - `import 'element-plus/es/components/loading/style/css'`
   - Resolves unstyled/invisible DOM elements when calling `ElNotification` or `ElMessage` from JavaScript/TypeScript code (bypassing `unplugin-vue-components` template scanner).
2. **Lifecycle Hook Enclosure**:
   - Enclosed `ElNotification.warning` in [ActivityList.vue](./vue/src/views/master-data/ActivityList.vue#L336-L342) inside `onMounted()` hook to prevent pre-mount evaluation issues.
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## \[Completed\] Configurable DNS Pre-cache & Pre-connect via Environment Variables

### Changes Made:
1. **Environment Configuration** ([.env.example](./vue/.env.example#L6-L8)):
   - Added `VITE_DNS_PREFETCH` variable to `.env.example` supporting a comma-separated list of domains/origins for DNS prefetching and preconnecting (e.g. `VITE_DNS_PREFETCH=https://speed.cloudflare.com,https://www.guangji.id`).
2. **Vite HTML Transformation Plugin** ([vite.config.ts](./vue/vite.config.ts#L50-L80)):
   - Implemented `dnsPrefetchPlugin` in Vite to read `env.VITE_DNS_PREFETCH`.
   - When set, parses domains and transforms `%VITE_DNS_PREFETCH_TAGS%` placeholder in `index.html` into `<link rel="dns-prefetch" href="...">` and `<link rel="preconnect" href="..." crossorigin />` tags.
   - When empty/not configured, cleanly removes `%VITE_DNS_PREFETCH_TAGS%` without rendering any pre-cache tags or extra whitespace.
3. **HTML Placeholder** ([index.html](./vue/index.html#L8)):
   - Added `%VITE_DNS_PREFETCH_TAGS%` placeholder inside `<head>` in `index.html`.
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## \[Completed\] Conditional Tooltip "Do fill filter first" on Report Excel Download Buttons ([SxyReportList.vue](./vue/src/views/report/SxyReportList.vue) & [UmatReportList.vue](./vue/src/views/report/UmatReportList.vue))

### Changes Made:
1. **Conditional Tooltip Integration**:
   - Wrapped the disabled "Download Report Excel" button in `<el-tooltip content="Do fill filter first" :disabled="!isDownloadTooltipActive" placement="top">` inside a `<span class="download-btn-wrapper">` element in both [SxyReportList.vue](./vue/src/views/report/SxyReportList.vue#L9-L17) and [UmatReportList.vue](./vue/src/views/report/UmatReportList.vue#L9-L17).
2. **Reactive Active Tooltip State**:
   - Defined `isDownloadTooltipActive = computed(() => dataList.value.length > 0 && !hasActiveFilter.value)`.
   - Tooltip displays "Do fill filter first" when data is shown in the table (`dataList.length > 0`) but no filter has been filled (`!hasActiveFilter`).
   - Automatically hides when the user fills any filter or when no dataset is present.
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.







