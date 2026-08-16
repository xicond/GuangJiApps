export interface Umat {
  id: number
  kode: string
  alias?: string
  nama_indonesia: string
  marga?: string
  nama_mandarin?: string
  alamat?: string
  alamat2?: string
  telepon?: string
  mobile?: string
  tempat_lahir?: string
  tanggal_lahir?: string | null
  usia?: number
  wilayah?: string
  jenis_kelamin?: string
  pekerjaan?: string
  pendidikan?: string
  tanggal_chiutao_int?: string | null
  tanggal_chiutao_man?: string
  tahun_chiutao_mandarin?: string
  waktu_chiutao_mandarin?: string
  pengajak?: string
  pengajak_manual?: string
  penanggung?: string
  penanggung_manual?: string
  tcs?: string
  uang_pahala?: number
  fotang_chiutao?: string
  fotang_aktif?: string
  sd2?: boolean
  tempat_sd2?: string
  tanggal_sd2?: string | null
  sd3?: boolean
  tempat_sd3?: string
  tanggal_sd3?: string | null
  kelas_umum?: string
  kelas_khusus?: string
  ching_khou?: boolean
  tanggal_ching_khou?: string | null
  tanggal_ancuo?: string | null
  nama_cetya_rumah?: string
  meninggal?: boolean
  tanggal_meninggal?: string | null
  tim_kerja?: string
  posisi?: string
  status_umat?: string
  keterangan?: string
  email?: string
  image_path?: string
  // status?: boolean
  mod_act?: string
  mod_by?: number
  mod_date?: string
  ikrar_1?: boolean
  ikrar_2?: boolean
  ikrar_3?: boolean
  ikrar_4?: boolean
  ikrar_5?: boolean
  ikrar_6?: boolean
  ren_chai_pan?: boolean
  tanggal_ren_chai_pan?: string | null
  lien_ciang_pan?: boolean
  tanggal_lien_ciang_pan?: string | null
  ciang_yen_pan?: boolean
  tanggal_ciang_yen_pan?: string | null
  active_status?: boolean
  nama_fotang_lain?: string
  nama_tcs_lain?: string
  kode_buku?: string
}

export interface UmatQueryParams {
  page?: number
  limit?: number
  alias?: string
  namaindonesia?: string
  lookup_description?: string
  namamandarin?: string
  tahunchiutaomandarin?: string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface UmatListResponse {
  data: Umat[]
  meta: PaginatedMeta
  resource: string
}

export interface UmatSingleResponse {
  data: Umat
  resource: string
}

export interface UmatReportItem {
  row_no: number
  id: number
  kode: string
  tanggal_chiu_tao_int?: string
  tanggal_chiu_tao_man?: string
  tahun_chiu_tao_mandarin?: string
  waktu_chiu_tao_mandarin?: string
  nama_indonesia: string
  nama_mandarin?: string
  alias?: string
  alamat?: string
  alamat2?: string
  uang_pahala?: number
  usia_thn?: number
  tanggal_lahir?: string
  tempat_lahir?: string
  pengajak_manual?: string
  penanggung_manual?: string
  tcs?: string
  telepon?: string
  mobile?: string
  email?: string
  fotang_ciu_tao_desc?: string
  fotang_aktif_desc?: string
  jenis_kelamin?: string
  wilayah?: string
  pekerjaan_desc?: string
  pendidikan_desc?: string
  tanggal_sd3?: string
  tempat_sd3_desc?: string
  tanggal_ching_khou?: string
  keterangan?: string
  kelas_umum_desc?: string
  kelas_khusus_desc?: string
  tanggal_an_cuo?: string
  nama_cetya_rumah?: string
  status_umat_desc?: string
  ikrar1?: boolean
  ikrar2?: boolean
  ikrar3?: boolean
  ikrar4?: boolean
  ikrar5?: boolean
  ikrar6?: boolean
  total_row?: number
}

export interface UmatReportQueryParams {
  page?: number
  limit?: number
  fotang_aktif?: string
  fotang_chiutao?: string
  nama_mandarin?: string
  start_date?: string
  end_date?: string
  nama_indo?: string
  pengajak?: string
  alias?: string
  usia_dari?: number | string
  usia_sampai?: number | string
  is_lulus_sd?: string
  is_vege?: string
  status_umat?: string
}

export interface UmatReportResponse {
  data: UmatReportItem[]
  meta: PaginatedMeta
  resource: string
}
