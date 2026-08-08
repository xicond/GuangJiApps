import apiClient from './client'
import type { LookupQueryParams, LookupListResponse } from '../types/lookup'

export const fotangApi = {
  /**
   * Fetch paginated list of Fotang lookups with optional search filters.
   */
  async getFotangLookup(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, string | number> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.lookup_value) cleanParams.lookup_value = params.lookup_value
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description
    if (params.search) cleanParams.search = params.search

    const response = await apiClient.get<LookupListResponse>('/v1/fotang/lookup', {
      params: cleanParams,
      signal
    })
    return response.data
  },
  async getFotangLookupSxy(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, string | number> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.lookup_value) cleanParams.lookup_value = params.lookup_value
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description
    if (params.search) cleanParams.search = params.search

    const response = await apiClient.get<LookupListResponse>('/v1/fotang/lookup-sxy', {
      params: cleanParams,
      signal
    })
    return response.data
  }
}

export default fotangApi
