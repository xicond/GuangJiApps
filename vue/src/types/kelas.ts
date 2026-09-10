import type { AppLookup } from './lookup'
import type { Umat } from './umat'

export interface Kelas {
  trx_id?: number | string
  kode_kelas?: string
  start_date?: string
  end_date?: string
  kode_fotang?: string
  lokasi?: string
  pic?: string
  keterangan?: string
  // status?: boolean
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
  detailid?: string | number
  detail_id?: string | number
  trx_id?: string | number
  id_peserta?: string | number
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
  sumbangan?: number
  barang?: string
  tim_kerja?: string
  keterangan?: string
  anak?: string
  suster?: string
  menginap?: string
  makanan_pagi?: string
  makanan_siang?: string
  makanan_malam?: string
  umat?: Partial<Umat>
}

export interface KelasQueryParams {
  page?: number
  limit?: number
  // lookup_id?: string
  // lookup_value?: string
  // lookup_description?: string
  kelas?: string
  lookup_description?: string
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

export interface KelasPesertaPrevious {
  id_peserta: number
  id?: number
  kode?: string
  nama_indonesia?: string
  nama_mandarin?: string
  alias?: string
  marga?: string
  alamat?: string
  fotang_aktif?: number
  fotang_aktif_desc?: string
  fotang_ciu_tao?: number
  fotang_ciu_tao_desc?: string
  pengajak?: string
  penanggung?: string
}

export interface KelasPesertaPreviousListResponse {
  data: KelasPesertaPrevious[]
  meta?: PaginatedMeta
  resource?: string
}

export interface KelasPesertaBulkPayload {
  trx_id?: number | string
  id_peserta: (number | string)[]
  sumbangan?: number
  barang?: string
  tim_kerja?: string
  keterangan?: string
  status?: boolean
  lulus?: boolean
  keterangan_lulus?: string
  anak?: string
  suster?: string
  menginap?: string
  makanan_pagi?: string
  makanan_siang?: string
  makanan_malam?: string
  umat?: Partial<Umat>
}

export interface KelasPesertaBulkResponse {
  data: KelasPeserta[]
  resource?: string
}

export interface KelasPengabdi {
  detailid?: number | string
  detail_id?: number | string
  trx_id?: number | string
  id_pengabdi?: number | string
  nama_indonesia?: string
  nama_mandarin?: string
  fotang_aktif_desc?: string
  tim_kerja_desc?: string
  sub_kerja_desc?: string
  sumbangan?: number
  barang?: string
  tim_kerja?: string
  tim_kerja_report?: string
  keterangan?: string
  hari?: string
  sub_kerja?: string
  anak?: string
  suster?: string
  menginap?: string
  makanan_pagi?: string
  makanan_siang?: string
  makanan_malam?: string
}

export interface KelasPengabdiListResponse {
  data: KelasPengabdi[]
  meta?: PaginatedMeta
  resource?: string
}

export interface KelasTopik {
  detailid?: number | string
  detail_id?: number | string
  trx_id?: number | string
  kode_topik?: string
  nama_topik?: string
  topic_category?: string
  topic_desc?: string
  urutan?: number
  topik_date?: string
  penceramah?: number | string
  penceramah_ext?: string
  penterjemah?: string
  durasi?: number
  keterangan?: string
  penceramah_nama_indonesia?: string
}

export interface KelasTopikListResponse {
  data: KelasTopik[]
  meta?: PaginatedMeta
  resource?: string
}

export interface KelasDonasi {
  detailid?: number | string
  detail_id?: number | string
  trx_id?: number | string
  donatur?: string
  donasi?: number
}

export interface KelasDonasiListResponse {
  data: KelasDonasi[]
  meta?: PaginatedMeta
  resource?: string
}

export interface KelasDonasiBarang {
  detailid?: number | string
  detail_id?: number | string
  trx_id?: number | string
  donatur?: string
  barang?: string
}

export interface KelasDonasiBarangListResponse {
  data: KelasDonasiBarang[]
  meta?: PaginatedMeta
  resource?: string
}

export interface KelasKendaraan {
  detailid?: number | string
  detail_id?: number | string
  trx_id?: number | string
  no_polisi?: string
  pengendara?: string
  tipe_kendaraan?: string
  fotang?: string
  hari?: string
  keterangan?: string
  status?: boolean
}

export interface KelasKendaraanListResponse {
  data: KelasKendaraan[]
  meta?: PaginatedMeta
  resource?: string
}

export interface KelasPengeluaran {
  detailid?: number | string
  detail_id?: number | string
  trx_id?: number | string
  tim_kerja?: string
  tim_kerja_desc?: string
  keterangan?: string
  biaya?: number
  status?: boolean
}

export interface KelasPengeluaranListResponse {
  data: KelasPengeluaran[]
  meta?: PaginatedMeta
  resource?: string
}
