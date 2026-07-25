export interface GroupMenuMapping {
  menu_id: number
  parent_id?: number
  menu_name: string
  page_url?: string
  sequence?: number
  menu_desc?: string
  parent_level_1?: number
  flag_active?: boolean
}

export interface GroupMenuMappingQueryParams {
  page?: number
  limit?: number
  menu_name?: string
  page_url?: string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface GroupMenuMappingListResponse {
  data: GroupMenuMapping[]
  meta: PaginatedMeta
  resource?: string
}

export interface GroupMenuMappingSingleResponse {
  data: GroupMenuMapping
  resource?: string
}
