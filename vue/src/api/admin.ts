import apiClient from './client'
import type {
  Admin,
  AdminQueryParams,
  AdminListResponse,
  AdminSingleResponse,
  AdminGroupListResponse,
  AdminDepartmentListResponse
} from '../types/admin'

export const adminApi = {
  /**
   * Fetch paginated list of Admins with optional filters.
   * Supports AbortSignal for canceling stale in-flight requests.
   */
  async getAdmins(
    params: AdminQueryParams = {},
    signal?: AbortSignal
  ): Promise<AdminListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.username) cleanParams.username = params.username
    if (params.group_name) cleanParams.group_name = params.group_name

    const response = await apiClient.get<AdminListResponse>('/v1/admins', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single Admin by ID.
   */
  async getAdminById(
    id: number | string,
    signal?: AbortSignal
  ): Promise<AdminSingleResponse> {
    const response = await apiClient.get<AdminSingleResponse>(`/v1/admins/${id}`, { signal })
    return response.data
  },

  /**
   * Create a new Admin record.
   */
  async createAdmin(payload: Partial<Admin>): Promise<AdminSingleResponse> {
    const response = await apiClient.post<AdminSingleResponse>('/v1/admins', payload)
    return response.data
  },

  /**
   * Update an existing Admin record.
   */
  async updateAdmin(
    id: number | string,
    payload: Partial<Admin>
  ): Promise<AdminSingleResponse> {
    const response = await apiClient.patch<AdminSingleResponse>(`/v1/admins/${id}`, payload)
    return response.data
  },

  /**
   * Delete an Admin record by ID.
   */
  async deleteAdmin(id: number | string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/admins/${id}`)
    return response.data
  },

  /**
   * Fetch Admin Groups for selection dropdown.
   */
  async getAdminGroups(signal?: AbortSignal): Promise<AdminGroupListResponse> {
    const response = await apiClient.get<AdminGroupListResponse>('/v1/admin-groups', {
      params: { limit: 100 },
      signal
    })
    return response.data
  },

  /**
   * Fetch Department for selection dropdown.
   */
  async getDepartments(signal?: AbortSignal): Promise<AdminDepartmentListResponse> {
    const response = await apiClient.get<AdminDepartmentListResponse>('/v1/department', { signal })
    return response.data
  }
}

export default adminApi
