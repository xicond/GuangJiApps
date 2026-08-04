import type { AppLookup } from './lookup'

export interface Topic {
  topic_code: string
  topic_name: string
  topic_category?: string
  description?: string
  // status?: boolean
  mod_act?: string
  mod_by?: string
  mod_date?: string
  topic_category_info?: AppLookup
}

export interface TopicQueryParams {
  page?: number
  limit?: number
  topic_code?: string
  topic_name?: string
  topic_category?: string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface TopicListResponse {
  data: Topic[]
  meta: PaginatedMeta
  resource?: string
}

export interface TopicSingleResponse {
  data: Topic
  resource?: string
}
