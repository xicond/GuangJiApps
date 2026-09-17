import apiClient from './client'
import type {
  PenggalangDana,
  PenggalangDanaQueryParams,
  PenggalangDanaListResponse,
  PenggalangDanaSingleResponse
} from '../types/penggalangDana'

export const penggalangDanaApi = {
  /**
   * Fetch paginated list of Penggalang Dana with optional search filters.
   */
  async getPenggalangDanas(
    params: PenggalangDanaQueryParams = {},
    signal?: AbortSignal
  ): Promise<PenggalangDanaListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.nama) cleanParams.nama = params.nama
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description
    if (params.mandarin) cleanParams.mandarin = params.mandarin
    if (params.fotang) cleanParams.fotang = params.fotang

    const response = await apiClient.get<PenggalangDanaListResponse>('/v1/penggalang-dana', {
      params: cleanParams,
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Fetch single PenggalangDana details by ID.
   */
  async getPenggalangDanaById(
    id: number | string,
    signal?: AbortSignal
  ): Promise<PenggalangDanaSingleResponse> {
    const response = await apiClient.get<PenggalangDanaSingleResponse>(`/v1/penggalang-dana/${id}`, {
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Create a new PenggalangDana record.
   */
  async createPenggalangDana(payload: Partial<PenggalangDana>): Promise<PenggalangDanaSingleResponse> {
    const response = await apiClient.post<PenggalangDanaSingleResponse>('/v1/penggalang-dana', payload)
    return response.data
  },

  /**
   * Update an existing PenggalangDana record.
   */
  async updatePenggalangDana(
    id: number | string,
    payload: Partial<PenggalangDana>
  ): Promise<PenggalangDanaSingleResponse> {
    const response = await apiClient.patch<PenggalangDanaSingleResponse>(`/v1/penggalang-dana/${id}`, payload)
    return response.data
  },

  /**
   * Delete a PenggalangDana record by ID.
   */
  async deletePenggalangDana(id: number | string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/penggalang-dana/${id}`)
    return response.data
  }
}

export default penggalangDanaApi
