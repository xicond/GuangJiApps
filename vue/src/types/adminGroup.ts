export interface AdminGroup {
  group_id: number
  group_name: string
  r_insert?: boolean
  r_edit?: boolean
  r_delete?: boolean
  r_reporting?: boolean
  r_position_id?: number
  group_desc?: string
}

export interface AdminGroupQueryParams {
  page?: number
  limit?: number
  group_name?: string
  group_desc?: string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface AdminGroupListResponse {
  data: AdminGroup[]
  meta: PaginatedMeta
  resource?: string
}

export interface AdminGroupSingleResponse {
  data: AdminGroup
  resource?: string
}
