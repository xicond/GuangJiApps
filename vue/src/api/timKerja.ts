import apiClient from './client'
import type {
  TimKerja,
  TimKerjaQueryParams,
  TimKerjaListResponse,
  TimKerjaSingleResponse
} from '../types/timKerja'

export const timKerjaApi = {
  /**
   * Fetch paginated list of Tim Kerja with optional search filters.
   */
  async getTimKerjas(
    params: TimKerjaQueryParams = {},
    signal?: AbortSignal
  ): Promise<TimKerjaListResponse> {
    const cleanParams: Record<string, string | number> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    // if (params.lookup_id) cleanParams.lookup_id = params.lookup_id
    if (params.lookup_value) cleanParams.lookup_value = params.lookup_value
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<TimKerjaListResponse>('/v1/tim-kerja', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single TimKerja details by ID.
   */
  async getTimKerjaById(
    id: string,
    signal?: AbortSignal
  ): Promise<TimKerjaSingleResponse> {
    const response = await apiClient.get<TimKerjaSingleResponse>(`/v1/tim-kerja/${id}`, { signal })
    return response.data
  },

  /**
   * Create a new TimKerja record.
   */
  async createTimKerja(payload: Partial<TimKerja>): Promise<TimKerjaSingleResponse> {
    const response = await apiClient.post<TimKerjaSingleResponse>('/v1/tim-kerja', payload)
    return response.data
  },

  /**
   * Update an existing TimKerja record.
   */
  async updateTimKerja(
    id: string,
    payload: Partial<TimKerja>
  ): Promise<TimKerjaSingleResponse> {
    const response = await apiClient.patch<TimKerjaSingleResponse>(`/v1/tim-kerja/${id}`, payload)
    return response.data
  },

  /**
   * Delete a TimKerja record by ID.
   */
  async deleteTimKerja(id: string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/tim-kerja/${id}`)
    return response.data
  }
}

export default timKerjaApi
