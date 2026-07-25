export interface DonasiSxy {
  id: number
  no_kwitansi: string
  tanggal?: string
  donatur_id?: number
  penggalang_id?: number
  jumlah?: number
  tipe_sumbangan?: number
  no_kupon?: string
  keterangan?: string
  status?: boolean
  created_by?: number
  created_date?: string
  updated_by?: number
  updated_date?: string
  tanggal_transfer?: string
  atas_nama?: string
  ttk_sent?: boolean
}

export interface DonasiSxyQueryParams {
  page?: number
  limit?: number
  no_kwitansi?: string
  donatur_id?: number
  tanggal?: string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface DonasiSxyListResponse {
  data: DonasiSxy[]
  meta: PaginatedMeta
  resource?: string
}

export interface DonasiSxySingleResponse {
  data: DonasiSxy
  resource?: string
}
