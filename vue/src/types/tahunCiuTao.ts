export interface TahunCiuTao {
  tahun_mandarin: string
  start_date?: string
  end_date?: string
  description?: string
  status?: boolean
  mod_act?: string
  mod_by?: string
  mod_date?: string
}

export interface TahunCiuTaoQueryParams {
  page?: number
  limit?: number
  tahun_mandarin?: string
  description?: string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface TahunCiuTaoListResponse {
  data: TahunCiuTao[]
  meta: PaginatedMeta
  resource?: string
}

export interface TahunCiuTaoSingleResponse {
  data: TahunCiuTao
  resource?: string
}
