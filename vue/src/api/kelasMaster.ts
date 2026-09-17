import apiClient from './client'
import type {
  KelasMaster,
  KelasMasterQueryParams,
  KelasMasterListResponse,
  KelasMasterSingleResponse
} from '../types/kelasMaster'

export const kelasMasterApi = {
  /**
   * Fetch paginated list of Kelas Master with optional search filters.
   */
  async getKelasMasters(
    params: KelasMasterQueryParams = {},
    signal?: AbortSignal
  ): Promise<KelasMasterListResponse> {
    const cleanParams: Record<string, string | number | boolean> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.lookup_id) cleanParams.lookup_id = params.lookup_id
    if (params.lookup_value) cleanParams.lookup_value = params.lookup_value
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description
    if (params.status !== undefined && params.status !== '') cleanParams.status = params.status

    const response = await apiClient.get<KelasMasterListResponse>('/v1/kelas-masters', {
      params: cleanParams,
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Fetch single Kelas Master details by ID or code.
   */
  async getKelasMasterById(
    id: string,
    signal?: AbortSignal
  ): Promise<KelasMasterSingleResponse> {
    const response = await apiClient.get<KelasMasterSingleResponse>('/v1/kelas-master', {
      params: { id },
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Create a new Kelas Master record.
   */
  async createKelasMaster(
    payload: Partial<KelasMaster>,
    signal?: AbortSignal
  ): Promise<KelasMasterSingleResponse> {
    const response = await apiClient.post<KelasMasterSingleResponse>('/v1/kelas-master', payload, {
      signal
    })
    return response.data
  },

  /**
   * Update an existing Kelas Master record.
   */
  async updateKelasMaster(
    id: string,
    payload: Partial<KelasMaster>,
    signal?: AbortSignal
  ): Promise<KelasMasterSingleResponse> {
    const response = await apiClient.patch<KelasMasterSingleResponse>('/v1/kelas-master', payload, {
      params: { id },
      signal
    })
    return response.data
  },

  /**
   * Delete a Kelas Master record by ID.
   */
  async deleteKelasMaster(
    id: string,
    signal?: AbortSignal
  ): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>('/v1/kelas-master', {
      params: { id },
      signal
    })
    return response.data
  }
}
