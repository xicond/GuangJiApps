export interface AppLookup {
  lookup_id: string
  category_id?: string
  lookup_value: string
  lookup_description?: string
}

export interface LookupQueryParams {
  page?: number
  limit?: number
  lookup_description?: string
  // lookup_value?: string
  // lookup_id?: string
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
