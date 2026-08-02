interface CacheEntry {
  timestamp: number
  data?: any
  promise?: Promise<any>
}

const lookupCacheMap = new WeakMap<Function, Map<string, CacheEntry>>()
const CACHE_TTL_MS = 30000 // 30 seconds TTL for lookup options

/**
 * Executes a lookup API function with in-flight deduplication and short-term caching.
 */
export async function cachedFetchLookup(
  fetchApiFn: (params: any, signal?: AbortSignal) => Promise<any>,
  params: Record<string, any>,
  signal?: AbortSignal,
  forceRefresh = false
): Promise<any> {
  let fnCache = lookupCacheMap.get(fetchApiFn)
  if (!fnCache) {
    fnCache = new Map<string, CacheEntry>()
    lookupCacheMap.set(fetchApiFn, fnCache)
  }

  const key = JSON.stringify(params)
  const now = Date.now()
  const existing = fnCache.get(key)

  if (!forceRefresh && existing) {
    // Return in-flight promise if currently fetching
    if (existing.promise) {
      return existing.promise
    }
    // Return cached response if within TTL
    if (existing.data && now - existing.timestamp < CACHE_TTL_MS) {
      return existing.data
    }
  }

  // Create new request promise
  const promise = (async () => {
    try {
      const res = await fetchApiFn(params)
      fnCache!.set(key, { timestamp: Date.now(), data: res })
      return res
    } catch (err) {
      fnCache!.delete(key)
      throw err
    }
  })()

  fnCache.set(key, { timestamp: now, promise })
  return promise
}

/**
 * Clear cache for a specific fetch function.
 */
export function clearLookupCache(fetchApiFn?: Function) {
  if (fetchApiFn) {
    lookupCacheMap.delete(fetchApiFn)
  }
}
