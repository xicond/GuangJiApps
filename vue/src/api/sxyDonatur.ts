import apiClient from './client'
import type {
  SxyDonatur,
  SxyDonaturQueryParams,
  SxyDonaturListResponse,
  SxyDonaturSingleResponse
} from '../types/sxyDonatur'

export const sxyDonaturApi = {
  /**
   * Fetch paginated list of Sxy Donatur with optional search filters.
   */
  async getSxyDonaturs(
    params: SxyDonaturQueryParams = {},
    signal?: AbortSignal
  ): Promise<SxyDonaturListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.nama) cleanParams.nama = params.nama
    if (params.mandarin) cleanParams.mandarin = params.mandarin
    if (params.fotang) cleanParams.fotang = params.fotang

    const response = await apiClient.get<SxyDonaturListResponse>('/v1/sxy-donatur', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single SxyDonatur details by ID.
   */
  async getSxyDonaturById(
    id: number | string,
    signal?: AbortSignal
  ): Promise<SxyDonaturSingleResponse> {
    const response = await apiClient.get<SxyDonaturSingleResponse>(`/v1/sxy-donatur/${id}`, { signal })
    return response.data
  },

  /**
   * Create a new SxyDonatur record.
   */
  async createSxyDonatur(payload: Partial<SxyDonatur>): Promise<SxyDonaturSingleResponse> {
    const response = await apiClient.post<SxyDonaturSingleResponse>('/v1/sxy-donatur', payload)
    return response.data
  },

  /**
   * Update an existing SxyDonatur record.
   */
  async updateSxyDonatur(
    id: number | string,
    payload: Partial<SxyDonatur>
  ): Promise<SxyDonaturSingleResponse> {
    const response = await apiClient.patch<SxyDonaturSingleResponse>(`/v1/sxy-donatur/${id}`, payload)
    return response.data
  },

  /**
   * Delete a SxyDonatur record by ID.
   */
  async deleteSxyDonatur(id: number | string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/sxy-donatur/${id}`)
    return response.data
  }
}

export default sxyDonaturApi
