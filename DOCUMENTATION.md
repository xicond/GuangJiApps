# Documentation - Completed Tasks

## [Completed] Login Rate Limiter with Redis & Standard Headers (`gin/internal/api/router.go` & `vue/src/views/Login.vue`)

### Summary & Changes Made:
- **Rate Limiting Engine**: Implemented `LoginRateLimiter` in [ratelimit.go](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/middleware/ratelimit.go) using `github.com/ulule/limiter/v3` with Redis store default (`REDIS_ADDR`) and memory store fallback if Redis is unconfigured or unreachable.
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
- **Docker Compose**: Added `redis` service (`redis:7-alpine`) and set `REDIS_ADDR=redis:6379` for Gin container in [docker-compose.yml](file:///Users/xicond/Workspace/www/GuangJiApps/docker-compose.yml).
- **Frontend Vue**: Updated [Login.vue](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/views/Login.vue) and [auth.ts](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/stores/auth.ts) to handle rate limit responses. When 429 / `retry_after` occurs:
  - Form inputs and submit button are disabled (`:disabled="loading || isLockedOut"`).
  - A live countdown timer (`formattedCountdown`) is displayed.
  - Automatically re-enables login once the countdown reaches 0 seconds.
- **Verification**: Tested with unit test `TestLoginRateLimiter` in [router_test.go](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/router_test.go), passing 100%.

---

## [Completed] Fixed Dev Service Worker TypeScript Import Transpilation (`vue/vite.config.ts`)

### Problem & Root Cause:
When using `VitePWA` with `devOptions.type = 'module'` and `strategies: 'injectManifest'`, Vite served `sw.ts` as an ES Module in dev mode. When `sw.ts` imported local `.ts` strategy files (`import { DynamicNetworkCacheStrategy } from './strategies/dynamicnetworkcache'`), Vite rewrote the import path to `/src/strategies/dynamicnetworkcache.ts`.
Browsers cannot execute uncompiled TypeScript (`.ts`) ES module imports directly inside Service Workers, causing the browser Service Worker initialization to fail with a syntax/mime-type error, preventing all `registerRoute` definitions from executing.

### Solution:
Updated `devOptions.type` from `'module'` to `'classic'` in [vite.config.ts](file:///Users/xicond/Workspace/www/GuangJiApps/vue/vite.config.ts).
This causes Vite and `esbuild` to bundle `sw.ts` and all imported local TypeScript modules into a single standalone JavaScript file (`dev-sw.js`), resolving all import paths into plain JavaScript and allowing `registerRoute` and `DynamicNetworkCacheStrategy` to execute cleanly in browser Service Workers.

---

## [Completed] Added Debug Logging & Improved Route Matching (`vue/src/sw.ts` & `vue/src/strategies/dynamicnetworkcache.ts`)

### Changes Made:
- Added explicit console logging in [sw.ts](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/sw.ts) for matched routes (`[SW Route] Matched /v1/ or /api/: ...`).
- Added console logging in [dynamicnetworkcache.ts](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/strategies/dynamicnetworkcache.ts) (`[SW DynamicNetworkCacheStrategy] Intercepting: GET ...`).
- Expanded route matcher in `sw.ts` to cover both `/v1/` and `/api/` endpoints cleanly under `DynamicNetworkCacheStrategy`.

---

## [Completed] Registered `DynamicNetworkCacheStrategy` for Default GET `/v1/` Routes (`vue/src/sw.ts`)

### Changes Made:
- Updated `vue/src/sw.ts` default route matcher for GET `/v1/` endpoints (such as `/v1/penggalang-dana?page=3&limit=10`) from standard Workbox `NetworkFirst` to `DynamicNetworkCacheStrategy`.
- Set default `timeoutMs: 500` and `debounceMs: 5000` for general `/v1/` API endpoints under `api-cache`.
- Verified type safety via `npm run type-check` (`vue-tsc --noEmit`), passing with 0 errors.

---

## [Completed] TypeScript `any` Type Reduction Across `vue/src` (`vue/src`)

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

## [Completed] Workbox `StrategyHandler` Integration (`vue/src/strategies/dynamicnetworkcache.ts`)

### Changes Made:
- Imported `StrategyHandler` directly from `workbox-strategies`.
- Replaced `any` parameter types on `handler: StrategyHandler` across `_handle` and `triggerDebouncedBackgroundUpdate`.
- `handler.cachePut(key, response)` now strongly types its return value as `Promise<boolean>` instead of `any`.
- Verified type safety via `npm run type-check` (`vue-tsc --noEmit`), passing with 0 errors.

---

## [Completed] Non-Blocking Async `handler.cachePut` via `handler.waitUntil` (`vue/src/strategies/dynamicnetworkcache.ts`)

### Changes Made:
Updated `DynamicNetworkCacheStrategy` to use `handler.waitUntil` for `cachePut`:
- `handler.cachePut(request, networkResponse.clone())` is executed inside `.then()` without `await`, returning `Response` immediately to Axios/main thread with zero delay.
- Passed the `putPromise` to `handler.waitUntil(putPromise)` to extend Service Worker lifecycle in background, guaranteeing IndexedDB cache write completes safely even if the main thread receives the response earlier.

---

## [Completed] Non-Blocking Parallel `DynamicNetworkCacheStrategy` with Fallback Race (`vue/src/strategies/dynamicnetworkcache.ts`)

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

## [Completed] Form Submission Page Disabling & Finally Block Enforcement

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

## [Completed] Network Speed Monitoring & Axios Metadata Type Fix

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

## [Completed] k6 Load Test Script (`vue/pentest.js`)

### Changes Made:
Updated `vue/pentest.js` ([pentest.js](file:///Users/xicond/Workspace/www/GuangJiApps/vue/pentest.js)) to focus strictly on performance load testing matching Postman collection specs:
1. **Authentication Flow**: Login via `POST /login` and extract JWT Bearer token.
2. **GET Umat Batch Parallel Pagination**: Hits page 1 first to read `meta.total` and determine `totalPages = Math.ceil(meta.total / limit)`. Then fires requests for pages 2 through `totalPages` concurrently in parallel using `http.batch()` without waiting sequentially for each page response.
3. **Automatic Configuration & Docker Resolution**: Reads `.env` using k6's `open()` helper and resolves backend URLs (`http://gin:8080`) when run inside Docker.

---

## [Completed] Gin Backend & IIS HttpPlatformHandler Concurrency Optimizations

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

## [Completed] Redundant `kelasApi.getKelasById` Call Prevention

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

## [Completed] Lookup API Request Deduplication & Lazy Mount Optimization (`LookupSelect.vue` & `lookupCache.ts`)

### Changes Made:
1. **Lookup In-Flight Request Deduplication & 30s Caching** (`vue/src/utils/lookupCache.ts`):
   - Created `cachedFetchLookup` utility using `WeakMap<Function, Map<string, CacheEntry>>` to automatically deduplicate simultaneous in-flight lookup requests for identical API functions & parameters.
   - Implemented a 30-second TTL cache for reference lookups (`getFotangLookup`, `getLookupPendidikan`, `getLookupPekerjaan`, etc.).

2. **Lazy Mount Loading & Cache Integration** (`vue/src/components/common/LookupSelect.vue`):
   - Integrated `cachedFetchLookup` into `LookupSelect.vue` to resolve multiple `LookupSelect` instances (e.g. 4 Fotang selectors in `UmatForm.vue`) into a single network call.
   - Updated `onMounted` hook to skip eager network requests when `modelValue` is empty and no initial option is provided, allowing dropdowns to load lazily on user interaction (`onVisibleChange`).

---

## [Completed] String Lookup Field Trimming (`UmatForm.vue` & `KelasForm.vue`)

### Changes Made:
1. **`UmatForm.vue`**:
   - Implemented `trimTargetFields` helper for target lookup string fields (`kelas_khusus`, `kelas_umum`, `fotang_chiutao`, `fotang_aktif`, `tcs`, `waktu_chiutao_mandarin`, `pendidikan`, `pekerjaan`, `jenis_kelamin`, `tempat_sd2`, `tempat_sd3`, `tim_kerja`, `posisi`, `status_umat`).
   - Applied trimming to `initialData` on watcher initialization to prevent trailing whitespace mismatch in `LookupSelect`.
   - Applied trimming to payload fields on `handleSubmit` before emitting `submit` event.

2. **`KelasForm.vue`**:
   - Implemented `trimTargetFields` helper for string lookup and text fields (`kode_kelas`, `kode_fotang`, `level`, `lokasi`, `pic`, `keterangan`, `mc1`..`mc5`).
   - Applied trimming to `initialData` on watcher initialization and payload fields on `handleSubmit`.

---

## [Completed] Validation Error HTTP Status Code Fix (HTTP 400 Bad Request)

### Changes Made:
1. **Router Error Handler Refactoring** (`gin/internal/api/router.go`):
   - Refactored `respondError(c, err)` and `respondValidationError(c, err)` to properly recognize `*service.ValidationError` via `errors.As(err, &vErr)` and check for validation keywords (`"tidak valid"`, `"is required"`, `"Validation failed"`, `"not found in lookup"`, `"invalid ID format"`).
   - Replaced all fragile string matching (`strings.Contains(err.Error(), "Validation failed")`) and hardcoded 404 error responses across all `Create`, `Update`, and `Delete` endpoints in `router.go` with unified `respondError(c, err)`.
   - Guaranteed that lookup and field validation errors (e.g. `fotang_aktif: field B_FOTANG nilai '31' tidak valid`) return **HTTP 400 Bad Request** with structured JSON payload containing `"error"` and `"details"` map.

2. **Unit Test Verification** (`gin/internal/api/router_test.go`):
   - Added unit test `TestValidationErrorReturns400` verifying that `service.ValidationError` produces HTTP status 400 Bad Request with error detail content. Verified that all Go unit tests pass (`go test ./...`).

---

## [Completed] JWT `userID` `float64` Type Assertion Fix in `UmatService`

### Changes Made:
1. **Unsafe Type Assertion Removal** (`gin/internal/service/umat_service.go`):
   - Replaced unsafe `int32(c.MustGet("userID").(int))` type assertions on lines 226, 328, and 358 with `getUserID(c)`.
   - `c.Set("userID", claims["sub"])` in `AuthMiddleware` stores `userID` from JWT claims as a `float64` (Standard JSON unmarshaling type in `golang-jwt`), causing `.(int)` to panic with `interface conversion: interface {} is float64, not int`.
   - Using `getUserID(c)` safely handles `float64`, `int`, `int32`, `string` or fallback `1` without panicking.
