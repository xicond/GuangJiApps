import apiClient from './client'
import type {
  Activity,
  ActivityQueryParams,
  ActivityListResponse,
  ActivitySingleResponse
} from '../types/activity'

export const activityApi = {
  /**
   * Fetch paginated list of Activity with optional search filters.
   */
  async getActivitys(
    params: ActivityQueryParams = {},
    signal?: AbortSignal
  ): Promise<ActivityListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.event_code) cleanParams.event_code = params.event_code
    if (params.event_name) cleanParams.event_name = params.event_name
    if (params.event_category) cleanParams.event_category = params.event_category

    const response = await apiClient.get<ActivityListResponse>('/v1/activities', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single Activity details by ID.
   */
  async getActivityById(
    id: string,
    signal?: AbortSignal
  ): Promise<ActivitySingleResponse> {
    const response = await apiClient.get<ActivitySingleResponse>(`/v1/activity?code=${id}`, { signal })
    return response.data
  },

  /**
   * Create a new Activity record.
   */
  async createActivity(payload: Partial<Activity>): Promise<ActivitySingleResponse> {
    const response = await apiClient.post<ActivitySingleResponse>('/v1/activity', payload)
    return response.data
  },

  /**
   * Update an existing Activity record.
   */
  async updateActivity(
    id: string,
    payload: Partial<Activity>
  ): Promise<ActivitySingleResponse> {
    const response = await apiClient.patch<ActivitySingleResponse>(`/v1/activity?code=${id}`, payload)
    return response.data
  },

  /**
   * Delete a Activity record by ID.
   */
  async deleteActivity(id: string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/activity?code=${id}`)
    return response.data
  }
}

export default activityApi
