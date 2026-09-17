import apiClient from './client'
import type {
  AdminGroup,
  AdminGroupQueryParams,
  AdminGroupListResponse,
  AdminGroupSingleResponse
} from '../types/adminGroup'

export const adminGroupApi = {
  /**
   * Fetch paginated list of Admin Group with optional search filters.
   */
  async getAdminGroups(
    params: AdminGroupQueryParams = {},
    signal?: AbortSignal
  ): Promise<AdminGroupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.group_name) cleanParams.group_name = params.group_name
    if (params.group_desc) cleanParams.group_desc = params.group_desc

    const response = await apiClient.get<AdminGroupListResponse>('/v1/admin-groups', {
      params: cleanParams,
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Fetch single AdminGroup details by ID.
   */
  async getAdminGroupById(
    id: number,
    signal?: AbortSignal
  ): Promise<AdminGroupSingleResponse> {
    const response = await apiClient.get<AdminGroupSingleResponse>(`/v1/admin-groups/${id}`, {
      signal, fetchOptions: {
        priority: 'high'
      },
    })
    return response.data
  },

  /**
   * Create a new AdminGroup record.
   */
  async createAdminGroup(payload: Partial<AdminGroup>): Promise<AdminGroupSingleResponse> {
    const response = await apiClient.post<AdminGroupSingleResponse>('/v1/admin-groups', payload)
    return response.data
  },

  /**
   * Update an existing AdminGroup record.
   */
  async updateAdminGroup(
    id: number,
    payload: Partial<AdminGroup>
  ): Promise<AdminGroupSingleResponse> {
    const response = await apiClient.patch<AdminGroupSingleResponse>(`/v1/admin-groups/${id}`, payload)
    return response.data
  },

  /**
   * Delete a AdminGroup record by ID.
   */
  async deleteAdminGroup(id: number): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/admin-groups/${id}`)
    return response.data
  }
}

export default adminGroupApi
