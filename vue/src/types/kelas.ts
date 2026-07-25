export interface Kelas {
  lookup_id: string
  category_id?: string
  lookup_value: string
  lookup_description?: string
  status?: boolean
  mod_act?: string
  mod_by?: string
  mod_date?: string
}

export interface KelasQueryParams {
  page?: number
  limit?: number
  lookup_id?: string
  lookup_value?: string
  lookup_description?: string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface KelasListResponse {
  data: Kelas[]
  meta: PaginatedMeta
  resource?: string
}

export interface KelasSingleResponse {
  data: Kelas
  resource?: string
}
