import apiClient from './client'
import type {
  GroupMenuMapping,
  GroupMenuMappingQueryParams,
  GroupMenuMappingListResponse,
  GroupMenuMappingSingleResponse
} from '../types/groupMenuMapping'

export const groupMenuMappingApi = {
  /**
   * Fetch paginated list of Group Menu Mapping with optional search filters.
   */
  async getGroupMenuMappings(
    params: GroupMenuMappingQueryParams = {},
    signal?: AbortSignal
  ): Promise<GroupMenuMappingListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.menu_name) cleanParams.menu_name = params.menu_name
    if (params.page_url) cleanParams.page_url = params.page_url

    const response = await apiClient.get<GroupMenuMappingListResponse>('/v1/group-menu-mappings', {
      params: cleanParams,
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Fetch single GroupMenuMapping details by ID.
   */
  async getGroupMenuMappingById(
    id: number,
    signal?: AbortSignal
  ): Promise<GroupMenuMappingSingleResponse> {
    const response = await apiClient.get<GroupMenuMappingSingleResponse>(`/v1/group-menu-mappings/${id}`, {
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Create a new GroupMenuMapping record.
   */
  async createGroupMenuMapping(payload: Partial<GroupMenuMapping>): Promise<GroupMenuMappingSingleResponse> {
    const response = await apiClient.post<GroupMenuMappingSingleResponse>('/v1/group-menu-mappings', payload)
    return response.data
  },

  /**
   * Update an existing GroupMenuMapping record.
   */
  async updateGroupMenuMapping(
    id: number,
    payload: Partial<GroupMenuMapping>
  ): Promise<GroupMenuMappingSingleResponse> {
    const response = await apiClient.patch<GroupMenuMappingSingleResponse>(`/v1/group-menu-mappings/${id}`, payload)
    return response.data
  },

  /**
   * Delete a GroupMenuMapping record by ID.
   */
  async deleteGroupMenuMapping(id: number): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/group-menu-mappings/${id}`)
    return response.data
  }
}

export default groupMenuMappingApi
