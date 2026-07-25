import apiClient from './client'
import type {
  Kelas,
  KelasQueryParams,
  KelasListResponse,
  KelasSingleResponse
} from '../types/kelas'

export const kelasApi = {
  /**
   * Fetch paginated list of Kelas with optional search filters.
   */
  async getKelass(
    params: KelasQueryParams = {},
    signal?: AbortSignal
  ): Promise<KelasListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.lookup_id) cleanParams.lookup_id = params.lookup_id
    if (params.lookup_value) cleanParams.lookup_value = params.lookup_value
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<KelasListResponse>('/v1/kelas', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single Kelas details by ID.
   */
  async getKelasById(
    id: string,
    signal?: AbortSignal
  ): Promise<KelasSingleResponse> {
    const response = await apiClient.get<KelasSingleResponse>(`/v1/kelas/${id}`, { signal })
    return response.data
  },

  /**
   * Create a new Kelas record.
   */
  async createKelas(payload: Partial<Kelas>): Promise<KelasSingleResponse> {
    const response = await apiClient.post<KelasSingleResponse>('/v1/kelas', payload)
    return response.data
  },

  /**
   * Update an existing Kelas record.
   */
  async updateKelas(
    id: string,
    payload: Partial<Kelas>
  ): Promise<KelasSingleResponse> {
    const response = await apiClient.patch<KelasSingleResponse>(`/v1/kelas/${id}`, payload)
    return response.data
  },

  /**
   * Delete a Kelas record by ID.
   */
  async deleteKelas(id: string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/kelas/${id}`)
    return response.data
  }
}

export default kelasApi
