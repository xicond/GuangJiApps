export interface DonasiSxy {
  id?: number
  no_kwitansi: string
  tanggal?: string
  donatur_id?: number
  penggalang_id?: number
  jumlah?: number
  tipe_sumbangan?: number
  no_kupon?: string
  keterangan?: string
  // status?: boolean
  created_by?: number
  created_date?: string
  updated_by?: number
  updated_date?: string
  tanggal_transfer?: string
  atas_nama?: string
  ttk_sent?: boolean
  tipe_sumbangan_desc?: string
  nama_penggalang?: string
  nama_donatur?: string
}

export interface DonasiSxyQueryParams {
  page?: number
  limit?: number
  no_kwitansi?: string
  start_date?: string
  end_date?: string
  donatur?: string
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

export interface SxyDonasiReportItem {
  no_kwitansi?: string
  tanggal?: string
  tanggal_transfer?: string
  atas_nama?: string
  donatur?: string
  penggalang_dana?: string
  tipe_sumbangan?: string
  jumlah?: number
  nokupon?: string
  keterangan?: string
  fotang?: string
}

export interface SxyReportQueryParams {
  page?: number
  limit?: number
  donatur?: string
  penggalang?: string
  start_date?: string
  end_date?: string
  fotang?: string | number
}

export interface SxyReportListMeta extends PaginatedMeta {
  total_jumlah?: number
}

export interface SxyReportListResponse {
  data: SxyDonasiReportItem[]
  meta: SxyReportListMeta
  resource?: string
}

