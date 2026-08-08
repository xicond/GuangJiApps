export interface TimKerja {
  // lookup_id: string
  category_id?: string
  lookup_value: string
  lookup_description?: string
  // status?: boolean
  mod_act?: string
  mod_by?: string
  mod_date?: string
}

export interface TimKerjaQueryParams {
  page?: number
  limit?: number
  // lookup_id?: string
  lookup_value?: string
  lookup_description?: string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface TimKerjaListResponse {
  data: TimKerja[]
  meta: PaginatedMeta
  resource?: string
}

export interface TimKerjaSingleResponse {
  data: TimKerja
  resource?: string
}
