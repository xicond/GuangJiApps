# Documentation - Completed Tasks

## [Completed] HTTPS Scheme Detection & Port 443 Logging (`gin/internal/api/router.go`)

### Summary & Changes Made:
- **`ForwardedHeaderMiddleware`**: Implemented `ForwardedHeaderMiddleware` in [router.go](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/router.go) to inspect `X-Forwarded-Proto`, `X-Forwarded-Scheme`, `X-Forwarded-Ssl`, and `Front-End-Https` headers sent by reverse proxies (such as IIS HttpPlatformHandler).
- **HTTPS & Port 443 Forced Logging**: Updated `FilterSuccessLogMiddleware` to detect if `Request.URL.Scheme` or `X-Forwarded-Proto` is `"https"` (or `X-Forwarded-Port` is `"443"`), forcing `c.Request.URL.Scheme = "https"` and tagging log lines with `[HTTPS:443]` (or `[HTTPS:443-PANIC]`).
- **Verification**: Built and verified Go binary via `docker compose run --rm gin-build`, passing with exit code 0.

---

## [Completed] HTTP Referer Logging in `FilterSuccessLogMiddleware` (`gin/internal/api/router.go`)

### Summary & Changes Made:
- **HTTP Referer Logging**: Updated `FilterSuccessLogMiddleware` in [router.go](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/router.go) to capture `c.Request.Referer()`. When the `Referer` header is present, it is automatically appended to `log.Printf` output (`| Ref: <referer_url>`) for both `[HTTP]` status >= 400 and `[PANIC]` log events.
- **Verification**: Compiled and verified Go server binary via `docker compose run --rm gin-build`, passing cleanly with exit code 0.

---

## [Completed] IIS HttpPlatformHandler & Client IP Extraction (`gin/dist/web.config`, `gin/internal/api/middleware/ratelimit.go`, `gin/internal/api/router.go`)

### Summary & Changes Made:
- **`middleware.GetClientIP(c)` Helper**: Created a robust `GetClientIP` function in [ratelimit.go](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/middleware/ratelimit.go) that checks incoming proxy headers (`X-Forwarded-For`, `X-Real-IP`, `X-ARR-ClientIP`, `X-Original-For`) to resolve the client's actual remote IP, filtering out loopback addresses (`127.0.0.1`, `::1`) and stripping ports via `net.SplitHostPort`.
- **Rate Limiter & Middleware Integration**: Updated `LoginRateLimiter.GetKey` and `FilterSuccessLogMiddleware` in [router.go](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/api/router.go) to use `middleware.GetClientIP(c)`.
- **Gin Proxy Trust Configuration**: Configured `r.ForwardedByClientIP = true` and `r.SetTrustedProxies(nil)` in `NewRouter` to trust proxy headers from upstream reverse proxies without fallback loopback locking.
- **Verification**: Built and verified Go binary via `docker compose run --rm gin-build`, passing with exit code 0.

---

## [Completed] UmatReportList UI Fixes & Clean Pagination Layout (`vue/src/views/report/UmatReportList.vue`)

### Summary & Changes Made:
- **Card Border & Filter Layout Fix**:
  - Removed redundant `border: 1px solid #e4e7ed` overrides on `.filter-card` and `.table-card` which were creating double/bolder borders on top of Element Plus default card styling. Updated styles to use Element Plus CSS variables (`var(--el-border-color-lighter)`, `var(--el-text-color-primary)`).
  - Grouped `Usia Dari` and `Usia Sampai` into a unified **Rentang Usia (Tahun)** control range container (`Dari` [input] `s/d` [input] `Sampai`).
  - Added dynamic validation rules (`:max="filters.usia_sampai"` for Dari and `:min="filters.usia_dari"` for Sampai) plus `onUsiaDariChange` / `onUsiaSampaiChange` handlers guaranteeing `usia_sampai >= usia_dari`.
  - Rebalanced Row 3 grid column spans (`Rentang Usia`, `Lulus SD3`, `Vege / Qing Kou`) to `:md="8"` (33.3% width each), filling the row cleanly.
  - Wrapped filter form in `<el-form label-position="top">` to ensure labels sit cleanly above inputs without squeezing controls horizontally.
- **Usia Filter Initialization & Execution Fix**: Changed `usia_dari` and `usia_sampai` defaults in `filters` reactive state from `0` to `null as number | null` (and added `placeholder="Dari"` / `placeholder="Sampai"`). Updated `fetchData()` and `handleDownloadExcel()` to pass `filters.usia_dari ?? undefined` and `filters.usia_sampai ?? undefined`. This prevents the age filter from automatically triggering for age 0 on initial page load or after clicking Reset Filter.
- **Pagination Layout String Cleanup**: Fixed redundant pagination layout expression string in [UmatReportList.vue](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/views/report/UmatReportList.vue), [SxyReportList.vue](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/views/report/SxyReportList.vue), and [DonasiSxyList.vue](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/views/transaction/DonasiSxyList.vue) to cleanly evaluate `isDesktop ? 'total, sizes, prev, pager, next, jumper' : 'total, sizes, prev, pager, next'`.
- **Verification**: Verified via `npm run type-check` (`vue-tsc --noEmit`) and production build (`npm run build`), passing with 0 errors.

---

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

## [Completed] Fixed Dev Service Worker ES Module Import Syntax Error (`vue/vite.config.ts`)

### Problem & Root Cause:
When using `VitePWA` with `devOptions.type = 'classic'` and `strategies: 'injectManifest'`, the browser registers the development service worker as a classic script (`type: 'classic'`). However, `dev-sw.js` generated by Vite contains top-level ES `import` statements (e.g. `import { precacheAndRoute ... } from 'workbox-precaching'`), causing the browser to throw `Uncaught SyntaxError: Cannot use import statement outside a module (at dev-sw.js?dev-sw:1:1)`.

### Solution:
Updated `devOptions.type` from `'classic'` to `'module'` in [vite.config.ts](file:///Users/xicond/Workspace/www/GuangJiApps/vue/vite.config.ts).
This tells `vite-plugin-pwa` to register the service worker as an ES module (`{ type: 'module' }`), allowing ES module `import` statements in `dev-sw.js` during development while building cleanly into bundled `sw.js` for production.

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

---

## [Completed] Seamless In-Place Gzip Compression & IIS `web.config` Setup

### Changes Made:
1. **In-Place Gzip Compression Plugin** (`vue/vite.config.ts`):
   - Created `compressInPlacePlugin` in `vite.config.ts` that compresses built static assets (`.js`, `.css`, images, `.svg`, `.html`, `.json`) in-place during `closeBundle()`.
   - All files retain their original filenames and extensions (no `.gz` postfix anywhere on disk or in URLs).
   - Confirmed `npm run build:client` reduces main JS bundle from 1.26MB to 398KB and CSS bundle from 363KB to 48KB directly on disk.

2. **IIS Seamless Header Configuration** (`vue/public/web.config`):
   - Configured an outbound rewrite rule (`Set Content-Encoding gzip for static assets`) in `web.config` to attach `Content-Encoding: gzip` HTTP response header for static files (`.js`, `.css`, images, `.svg`, `.html`, `.json`, etc.).

---

## [Completed] Persistent Offline Warning Notification (`vue/src/layouts/MainLayout.vue`)

### Changes Made:
- Imported `useOnline` from `@vueuse/core` in [MainLayout.vue](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/layouts/MainLayout.vue#L172).
- Added reactive watcher on `isOnline`:
  - **When Offline (`isOnline === false`)**: Displays an `ElNotification.warning` with `duration: 0` (persistent notification that stays open until network connectivity returns).
  - **When Online (`isOnline === true`)**: Automatically closes the active offline notification handle via `.close()`.
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## [Completed] PWA Asset Update Detection & Page Change Application (`vue/src/sw.ts` & `vue/src/utils/pwaUpdate.ts`)

### Changes Made:
1. **Service Worker Update Broadcast & Skip Waiting** ([sw.ts](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/sw.ts#L25-L45)):
   - Added message listener for `SKIP_WAITING` event to immediately activate new Service Workers.
   - Added `activate` event handler broadcasting `{ type: 'SW_UPDATED' }` via `postMessage` to all active window clients when new cached assets are activated.

2. **Router Navigation Update Guard** ([pwaUpdate.ts](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/utils/pwaUpdate.ts)):
   - Created `initPwaUpdate(router)` utility initialized in [main.ts](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/main.ts#L25).
   - Listens to Workbox `onNeedRefresh` and SW `SW_UPDATED` messages to mark `hasPendingUpdate = true`.
   - On page navigation (`router.beforeEach`), if a SW update is pending, triggers `updateSW(true)` (reloading seamlessly to apply updated assets on page change).
   - Automatically checks `registration.update()` on every route change (`router.afterEach`) to detect newly deployed builds.
- Verified with `vue-tsc --noEmit`, passing cleanly with 0 type errors.

---

## [Completed] Umat Image Processing, Cross-Platform Validation & `UmatFoto` One-To-One Record ([umat_service.go](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/umat_service.go))

### Changes Made:
1. **Optional Image Header Processing** ([umat_service.go](file:///Users/xicond/Workspace/www/GuangJiApps/gin/internal/service/umat_service.go)):
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

## [Completed] Frontend Photo Capture, 2:4 Cropper & Postman Multipart Collection Update

### Changes Made:
1. **Postman Collection Update** ([apps-gin.postman_collection.json](file:///Users/xicond/Workspace/www/GuangJiApps/gin/postman/apps-gin.postman_collection.json)):
   - Added `Umat - Create (Multipart with Photo)` and `Umat - Update (Multipart with Photo)` endpoints with `formdata` containing `"data"` (JSON payload) and `"foto"` (file header).
2. **Vue Multipart API Support** ([umat.ts](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/api/umat.ts)):
   - Updated `umatApi.createUmat` and `umatApi.updateUmat` to accept optional `photoFile?: File | Blob | null`.
   - Sends `FormData` with `multipart/form-data` content-type header when `photoFile` is present.
3. **Photo Capture & 2:4 Aspect Ratio Cropper** ([UmatPhotoUpload.vue](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/components/umat/UmatPhotoUpload.vue)):
   - Installed `vue-advanced-cropper` package.
   - HTML5 File Input with `accept="image/jpeg,image/png,image/webp,image/gif"` and `capture="environment"` for mobile camera and gallery picking.
   - Custom MediaDevices API live camera stream modal (`navigator.mediaDevices.getUserMedia`) for desktop/mobile browser video feed capture.
   - Interactive cropper modal enforcing strict **2:4 aspect ratio** (`aspectRatio: 2/4`) and output resolution.
4. **Form Integration**:
   - Embedded `UmatPhotoUpload` into Section 1 of [UmatForm.vue](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/components/umat/UmatForm.vue).
   - Updated [UmatCreateList.vue](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/views/master-data/UmatCreateList.vue) and [UmatEditList.vue](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/views/master-data/UmatEditList.vue) to forward `photoFile` blob to `createUmat` and `updateUmat`.
5. **Responsive Layout Tuning**:
   - Updated `.photo-preview-container` in [UmatPhotoUpload.vue](file:///Users/xicond/Workspace/www/GuangJiApps/vue/src/components/umat/UmatPhotoUpload.vue) to `width: 100%`, `max-width: 210px`, and `aspect-ratio: 3 / 4` for responsive scaling across devices.
- Verified with `npx vue-tsc --noEmit`, passing cleanly with 0 type errors.
