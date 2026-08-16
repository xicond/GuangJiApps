import apiClient from './client'
import type {
  Topic,
  TopicQueryParams,
  TopicListResponse,
  TopicSingleResponse
} from '../types/topic'

export const topicApi = {
  /**
   * Fetch paginated list of Topic with optional search filters.
   */
  async getTopics(
    params: TopicQueryParams = {},
    signal?: AbortSignal
  ): Promise<TopicListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.topic_code) cleanParams.topic_code = params.topic_code
    if (params.topic_name) cleanParams.topic_name = params.topic_name
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description
    if (params.topic_category) cleanParams.topic_category = params.topic_category

    const response = await apiClient.get<TopicListResponse>('/v1/topics', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch topics for lookup from GET /v1/topics
   */
  async getTopikLookup(
    params: TopicQueryParams = {},
    signal?: AbortSignal
  ): Promise<TopicListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.topic_code) cleanParams.topic_code = params.topic_code
    if (params.topic_name) cleanParams.topic_name = params.topic_name
    if (params.topic_category) cleanParams.topic_category = params.topic_category

    const response = await apiClient.get<TopicListResponse>('/v1/topics', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single Topic details by ID/Code.
   */
  async getTopicById(
    id: string,
    signal?: AbortSignal
  ): Promise<TopicSingleResponse> {
    const response = await apiClient.get<TopicSingleResponse>('/v1/topic', {
      params: { code: id },
      signal
    })
    return response.data
  },

  /**
   * Create a new Topic record.
   */
  async createTopic(payload: Partial<Topic>): Promise<TopicSingleResponse> {
    const response = await apiClient.post<TopicSingleResponse>('/v1/topic', payload)
    return response.data
  },

  /**
   * Update an existing Topic record by ID/Code.
   */
  async updateTopic(
    id: string,
    payload: Partial<Topic>
  ): Promise<TopicSingleResponse> {
    const response = await apiClient.patch<TopicSingleResponse>('/v1/topic', payload, {
      params: { code: id }
    })
    return response.data
  },

  /**
   * Delete a Topic record by ID/Code.
   */
  async deleteTopic(id: string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>('/v1/topic', {
      params: { code: id }
    })
    return response.data
  }
}

export default topicApi
