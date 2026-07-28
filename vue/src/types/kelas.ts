export interface Kelas {
  trx_id?: number | string
  kode_kelas?: string
  start_date?: string
  end_date?: string
  kode_fotang?: string
  lokasi?: string
  pic?: string
  keterangan?: string
  status?: boolean
  level?: string
  mc1?: string
  mc2?: string
  mc3?: string
  mc4?: string
  mc5?: string
  deadline?: string
  kelas_desc?: string
  fotang_desc?: string
}

export interface KelasQueryParams {
  page?: number
  limit?: number
  lookup_id?: string
  lookup_value?: string
  lookup_description?: string
  kelas?: string
  start_date?: string
  end_date?: string
  fotang?: string
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
