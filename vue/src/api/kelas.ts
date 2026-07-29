import apiClient from './client'
import type {
  Kelas,
  KelasQueryParams,
  KelasListResponse,
  KelasSingleResponse,
  KelasPesertaListResponse
} from '../types/kelas'
import type { LookupQueryParams, LookupListResponse } from '../types/lookup'

export const kelasApi = {
  /**
   * Fetch paginated list of Kelas lookups with optional search filters.
   */
  async getKelasLookup(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/kelas/lookup', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch paginated list of Kelas with optional search filters.
   */
  async getKelasList(
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
    if (params.kelas) cleanParams.kelas = params.kelas
    if (params.start_date) cleanParams.start_date = params.start_date
    if (params.end_date) cleanParams.end_date = params.end_date
    if (params.fotang) cleanParams.fotang = params.fotang

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
    id: string | number,
    signal?: AbortSignal
  ): Promise<KelasSingleResponse> {
    const response = await apiClient.get<KelasSingleResponse>(`/v1/kelas/${id}`, { signal })
    return response.data
  },

  /**
   * Fetch list of participants (peserta) for a specific Kelas by ID with optional pagination.
   */
  async getKelasPeserta(
    id: string | number,
    params: { page?: number; limit?: number } = {},
    signal?: AbortSignal
  ): Promise<KelasPesertaListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    const response = await apiClient.get<KelasPesertaListResponse>(`/v1/kelas/${id}/peserta`, {
      params: cleanParams,
      signal
    })
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
    id: string | number,
    payload: Partial<Kelas>
  ): Promise<KelasSingleResponse> {
    const response = await apiClient.patch<KelasSingleResponse>(`/v1/kelas/${id}`, payload)
    return response.data
  },

  /**
   * Delete a Kelas record by ID.
   */
  async deleteKelas(id: string | number): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/kelas/${id}`)
    return response.data
  }
}

export default kelasApi
