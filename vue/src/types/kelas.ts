import type { AppLookup } from './lookup'

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
  kelas_name?: AppLookup
  fotang_name?: AppLookup
}

export interface KelasPeserta {
  detailid?: string
  trx_id?: string
  id_peserta?: string
  nama_indonesia?: string
  nama_mandarin?: string
  fotang_aktif_desc?: string
  fotang_ciutao_desc?: string
  tanggal_ciu_tao_int?: string
  pengajak?: string
  penanggung?: string
  lulus?: boolean
  keterangan_lulus?: string
  ikrar1?: boolean
  ikrar2?: boolean
  ikrar3?: boolean
  ikrar4?: boolean
  ikrar5?: boolean
  ikrar6?: boolean
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

export interface KelasPesertaListResponse {
  data: KelasPeserta[]
  meta?: PaginatedMeta
  resource?: string
}
