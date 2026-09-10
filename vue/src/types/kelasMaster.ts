export interface KelasMaster {
  lookup_id: string
  category_id?: string
  lookup_value?: string
  lookup_description?: string
  status?: boolean
  mod_act?: string
  mod_by?: string
  mod_date?: string
}

export interface KelasMasterQueryParams {
  page?: number
  limit?: number
  lookup_id?: string
  lookup_value?: string
  lookup_description?: string
  status?: string | boolean
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface KelasMasterListResponse {
  data: KelasMaster[]
  meta: PaginatedMeta
  resource?: string
}

export interface KelasMasterSingleResponse {
  data: KelasMaster
  resource?: string
}
