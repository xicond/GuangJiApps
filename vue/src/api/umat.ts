import apiClient from './client'
import type {
  Umat,
  UmatQueryParams,
  UmatListResponse,
  UmatSingleResponse
} from '../types/umat'

export const umatApi = {
  /**
   * Fetch paginated list of umats with optional search filters.
   * Supports AbortSignal for canceling outdated requests.
   */
  async getUmats(
    params: UmatQueryParams = {},
    signal?: AbortSignal
  ): Promise<UmatListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.alias) cleanParams.alias = params.alias
    if (params.namaindonesia) cleanParams.namaindonesia = params.namaindonesia
    if (params.namamandarin) cleanParams.namamandarin = params.namamandarin
    if (params.tahunchiutaomandarin) {
      cleanParams.tahunchiutaomandarin = params.tahunchiutaomandarin
    }

    const response = await apiClient.get<UmatListResponse>('/v1/umats', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single Umat details by ID.
   */
  async getUmatById(
    id: number | string,
    signal?: AbortSignal
  ): Promise<UmatSingleResponse> {
    const response = await apiClient.get<UmatSingleResponse>(`/v1/umats/${id}`, { signal })
    return response.data
  },

  /**
   * Create a new Umat record.
   */
  async createUmat(payload: Partial<Umat>): Promise<UmatSingleResponse> {
    const response = await apiClient.post<UmatSingleResponse>('/v1/umats', payload)
    return response.data
  },

  /**
   * Update an existing Umat record.
   */
  async updateUmat(
    id: number | string,
    payload: Partial<Umat>
  ): Promise<UmatSingleResponse> {
    const response = await apiClient.patch<UmatSingleResponse>(`/v1/umats/${id}`, payload)
    return response.data
  },

  /**
   * Soft/Hard Delete a Umat record by ID.
   */
  async deleteUmat(id: number | string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/umats/${id}`)
    return response.data
  }
}

export default umatApi
