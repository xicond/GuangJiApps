interface CacheEntry<T = unknown> {
  timestamp: number
  data?: T
  promise?: Promise<T>
}

// eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
const lookupCacheMap = new WeakMap<Function, Map<string, CacheEntry<unknown>>>()
const CACHE_TTL_MS = 60000 // 30 seconds TTL for lookup options

/**
 * Executes a lookup API function with in-flight deduplication and short-term caching.
 */
export async function cachedFetchLookup<T = unknown, P = Record<string, unknown>>(
  fetchApiFn: (params: P, signal?: AbortSignal) => Promise<T>,
  params: P,
  signal?: AbortSignal,
  forceRefresh = false
): Promise<T> {
  let fnCache = lookupCacheMap.get(fetchApiFn) as Map<string, CacheEntry<T>> | undefined
  if (!fnCache) {
    fnCache = new Map<string, CacheEntry<T>>()
    lookupCacheMap.set(fetchApiFn, fnCache as Map<string, CacheEntry<unknown>>)
  }

  const key = JSON.stringify(params)
  const now = Date.now()
  const existing = fnCache.get(key)

  if (!forceRefresh && existing) {
    if (existing.promise) {
      return existing.promise
    }
    if (existing.data !== undefined && now - existing.timestamp < CACHE_TTL_MS) {
      return existing.data
    }
  }

  const promise = (async () => {
    try {
      const res = await fetchApiFn(params, signal)
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
