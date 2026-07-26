export interface SxyDonatur {
  no?: string
  nama: string
  mandarin?: string
  keterangan?: string
  lookup_fothang?: number
  alamat?: string
  telepon?: string
  mobile?: string
  email?: string
  status?: boolean
  created_by?: number
  created_date?: string
  updated_by?: number
  updated_date?: string
}

export interface SxyDonaturQueryParams {
  page?: number
  limit?: number
  nama?: string
  mandarin?: string
  fotang?: number
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
