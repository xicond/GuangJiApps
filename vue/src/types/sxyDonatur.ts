export interface SxyDonatur {
  id?: number
  no?: string
  nama: string
  mandarin?: string
  keterangan?: string
  lookup_fothang?: number
  fotang?: string
  alamat?: string
  telepon?: string
  mobile?: string
  email?: string
  // status?: boolean
  created_by?: number
  created_date?: string
  updated_by?: number
  updated_date?: string
}

export interface SxyDonaturQueryParams {
  page?: number
  limit?: number
  nama?: string
  lookup_description?: string
  mandarin?: string
  fotang?: number | string
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface SxyDonaturListResponse {
  data: SxyDonatur[]
  meta: PaginatedMeta
  resource?: string
}

export interface SxyDonaturSingleResponse {
  data: SxyDonatur
  resource?: string
}
