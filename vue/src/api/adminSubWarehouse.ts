import apiClient from './client'
import type {
  AdminSubWarehouse,
  AdminSubWarehouseQueryParams,
  AdminSubWarehouseListResponse,
  AdminSubWarehouseSingleResponse
} from '../types/adminSubWarehouse'

export const adminSubWarehouseApi = {
  /**
   * Fetch paginated list of Admin Sub Warehouse with optional search filters.
   */
  async getAdminSubWarehouses(
    params: AdminSubWarehouseQueryParams = {},
    signal?: AbortSignal
  ): Promise<AdminSubWarehouseListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.full_name) cleanParams.full_name = params.full_name
    if (params.pic) cleanParams.pic = params.pic
    if (params.doc_code) cleanParams.doc_code = params.doc_code
    if (params.sub_wh_type) cleanParams.sub_wh_type = params.sub_wh_type

    const response = await apiClient.get<AdminSubWarehouseListResponse>('/v1/admin-sub-warehouses', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single AdminSubWarehouse details by ID.
   */
  async getAdminSubWarehouseById(
    id: number,
    signal?: AbortSignal
  ): Promise<AdminSubWarehouseSingleResponse> {
    const response = await apiClient.get<AdminSubWarehouseSingleResponse>(`/v1/admin-sub-warehouses/${id}`, { signal })
    return response.data
  },

  /**
   * Create a new AdminSubWarehouse record.
   */
  async createAdminSubWarehouse(payload: Partial<AdminSubWarehouse>): Promise<AdminSubWarehouseSingleResponse> {
    const response = await apiClient.post<AdminSubWarehouseSingleResponse>('/v1/admin-sub-warehouses', payload)
    return response.data
  },

  /**
   * Update an existing AdminSubWarehouse record.
   */
  async updateAdminSubWarehouse(
    id: number,
    payload: Partial<AdminSubWarehouse>
  ): Promise<AdminSubWarehouseSingleResponse> {
    const response = await apiClient.patch<AdminSubWarehouseSingleResponse>(`/v1/admin-sub-warehouses/${id}`, payload)
    return response.data
  },

  /**
   * Delete a AdminSubWarehouse record by ID.
   */
  async deleteAdminSubWarehouse(id: number): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/admin-sub-warehouses/${id}`)
    return response.data
  }
}

export default adminSubWarehouseApi
