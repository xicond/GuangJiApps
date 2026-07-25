export interface AdminSubWarehouse {
  sub_wh_id: number
  wh_id?: number
  full_name: string
  pic?: string
  doc_code?: string
  cru_id?: number
  update_uid?: number
  lst_update?: string
  flag_productions?: boolean
  sub_wh_type?: string
}

export interface AdminSubWarehouseQueryParams {
  page?: number
  limit?: number
  full_name?: string
  pic?: string
  doc_code?: string
  sub_wh_type?: string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface AdminSubWarehouseListResponse {
  data: AdminSubWarehouse[]
  meta: PaginatedMeta
  resource?: string
}

export interface AdminSubWarehouseSingleResponse {
  data: AdminSubWarehouse
  resource?: string
}
