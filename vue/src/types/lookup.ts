export interface AppLookup {
  lookup_id?: string
  category_id?: string
  lookup_value?: string
  lookup_description?: string
  [key: string]: unknown
}

export interface LookupQueryParams {
  page?: number
  limit?: number
  lookup_description?: string
  lookup_value?: string
  lookup_id?: string
  search?: string
}

export interface LookupListResponse {
  data: AppLookup[]
  meta?: {
    page: number
    limit: number
    total: number
  }
  resource?: string
}
