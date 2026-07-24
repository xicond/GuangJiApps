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
  status?: boolean
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
