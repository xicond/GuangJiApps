export interface Activity {
  event_code: string
  event_name: string
  event_category?: string
  description?: string
  status?: boolean
  mod_act?: string
  mod_by?: string
  mod_date?: string
}

export interface ActivityQueryParams {
  page?: number
  limit?: number
  event_code?: string
  event_name?: string
  event_category?: string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface ActivityListResponse {
  data: Activity[]
  meta: PaginatedMeta
  resource?: string
}

export interface ActivitySingleResponse {
  data: Activity
  resource?: string
}
