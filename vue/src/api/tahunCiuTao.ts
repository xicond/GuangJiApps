import apiClient from './client'
import type {
  TahunCiuTao,
  TahunCiuTaoQueryParams,
  TahunCiuTaoListResponse,
  TahunCiuTaoSingleResponse
} from '../types/tahunCiuTao'

export const tahunCiuTaoApi = {
  /**
   * Fetch paginated list of Tahun Ciu Tao with optional search filters.
   */
  async getTahunCiuTaos(
    params: TahunCiuTaoQueryParams = {},
    signal?: AbortSignal
  ): Promise<TahunCiuTaoListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.tahun_mandarin) cleanParams.tahun_mandarin = params.tahun_mandarin
    if (params.description) cleanParams.description = params.description

    const response = await apiClient.get<TahunCiuTaoListResponse>('/v1/tahun-ciu-tao/list', {
      params: cleanParams,
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Fetch single TahunCiuTao details by ID.
   */
  async getTahunCiuTaoById(
    id: string,
    signal?: AbortSignal
  ): Promise<TahunCiuTaoSingleResponse> {
    const response = await apiClient.get<TahunCiuTaoSingleResponse>(`/v1/tahun-ciu-tao?tahun=${id}`, {
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Create a new TahunCiuTao record.
   */
  async createTahunCiuTao(payload: Partial<TahunCiuTao>): Promise<TahunCiuTaoSingleResponse> {
    const response = await apiClient.post<TahunCiuTaoSingleResponse>('/v1/tahun-ciu-tao', payload)
    return response.data
  },

  /**
   * Update an existing TahunCiuTao record.
   */
  async updateTahunCiuTao(
    id: string,
    payload: Partial<TahunCiuTao>
  ): Promise<TahunCiuTaoSingleResponse> {
    const response = await apiClient.patch<TahunCiuTaoSingleResponse>(`/v1/tahun-ciu-tao?tahun=${id}`, payload)
    return response.data
  },

  /**
   * Delete a TahunCiuTao record by ID.
   */
  async deleteTahunCiuTao(id: string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/tahun-ciu-tao?tahun=${id}`)
    return response.data
  }
}

export default tahunCiuTaoApi
