import apiClient from './client'
import type {
  Kelas,
  KelasPeserta,
  KelasPengabdi,
  KelasTopik,
  KelasDonasi,
  KelasDonasiBarang,
  KelasKendaraan,
  KelasKendaraanListResponse,
  KelasPengeluaran,
  KelasPengeluaranListResponse,
  KelasQueryParams,
  KelasListResponse,
  KelasSingleResponse,
  KelasPesertaListResponse,
  KelasPesertaPreviousListResponse,
  KelasPesertaBulkPayload,
  KelasPesertaBulkResponse,
  KelasPengabdiListResponse,
  KelasTopikListResponse,
  KelasDonasiListResponse,
  KelasDonasiBarangListResponse
} from '../types/kelas'
import type { LookupQueryParams, LookupListResponse } from '../types/lookup'

export const kelasApi = {
  /**
   * Fetch paginated list of Kelas lookups with optional search filters.
   */
  async getKelasLookup(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, string | number> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    // if (params.lookup_id) cleanParams.lookup_id = params.lookup_id
    if (params.lookup_value) cleanParams.lookup_value = params.lookup_value
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description
    if (params.search) cleanParams.search = params.search

    const response = await apiClient.get<LookupListResponse>('/v1/kelas/lookup', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch paginated list of Kelas with optional search filters.
   */
  async getKelasList(
    params: KelasQueryParams = {},
    signal?: AbortSignal
  ): Promise<KelasListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    // if (params.lookup_id) cleanParams.lookup_id = params.lookup_id
    // if (params.lookup_value) cleanParams.lookup_value = params.lookup_value
    // if (params.lookup_description) cleanParams.lookup_description = params.lookup_description
    if (params.kelas) cleanParams.kelas = params.kelas
    if (params.start_date) cleanParams.start_date = params.start_date
    if (params.end_date) cleanParams.end_date = params.end_date
    if (params.fotang) cleanParams.fotang = params.fotang

    const response = await apiClient.get<KelasListResponse>('/v1/kelas', {
      params: cleanParams,
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Fetch single Kelas details by ID.
   */
  async getKelasById(
    id: string | number,
    signal?: AbortSignal
  ): Promise<KelasSingleResponse> {
    const response = await apiClient.get<KelasSingleResponse>(`/v1/kelas/${id}`, {
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Download Excel report for a specific Kelas by ID.
   */
  async downloadReport(id: string | number): Promise<Blob> {
    const response = await apiClient.get(`/v1/kelas/${id}/report`, {
      fetchOptions: {
        priority: 'low'
      },
      responseType: 'blob'
    })
    return response.data
  },

  /**
   * Fetch list of participants (peserta) for a specific Kelas by ID with optional pagination.
   */
  async getKelasPeserta(
    id: string | number,
    params: { page?: number; limit?: number } = {},
    signal?: AbortSignal
  ): Promise<KelasPesertaListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    const response = await apiClient.get<KelasPesertaListResponse>(`/v1/kelas/${id}/peserta`, {
      params: cleanParams,
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Create a new Kelas record.
   */
  async createKelas(payload: Partial<Kelas>): Promise<KelasSingleResponse> {
    const response = await apiClient.post<KelasSingleResponse>('/v1/kelas', payload)
    return response.data
  },

  /**
   * Update an existing Kelas record.
   */
  async updateKelas(
    id: string | number,
    payload: Partial<Kelas>
  ): Promise<KelasSingleResponse> {
    const response = await apiClient.patch<KelasSingleResponse>(`/v1/kelas/${id}`, payload)
    return response.data
  },

  /**
   * Delete a Kelas record by ID.
   */
  async deleteKelas(id: string | number): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/kelas/${id}`)
    return response.data
  },

  /**
   * Fetch single KelasPeserta by Detail ID.
   */
  async getKelasPesertaById(id: string | number, kelasId: string | number = 0): Promise<{ data: KelasPeserta }> {
    const response = await apiClient.get<{ data: KelasPeserta }>(`/v1/kelas/${kelasId}/peserta/${id}`, {
      fetchOptions: {
        priority: 'high'
      }
    })
    return response.data
  },

  /**
   * Create a new KelasPeserta record.
   */
  async createKelasPeserta(payload: Partial<KelasPeserta>): Promise<{ data: KelasPeserta }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.post<{ data: KelasPeserta }>(`/v1/kelas/${kelasId}/peserta`, payload)
    return response.data
  },

  /**
   * Update an existing KelasPeserta record.
   */
  async updateKelasPeserta(
    id: string | number,
    payload: Partial<KelasPeserta>
  ): Promise<{ data: KelasPeserta }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.patch<{ data: KelasPeserta }>(`/v1/kelas/${kelasId}/peserta/${id}`, payload)
    return response.data
  },

  /**
   * Delete a KelasPeserta record by Detail ID.
   */
  async deleteKelasPeserta(id: string | number, kelasId: string | number = 0): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/kelas/${kelasId}/peserta/${id}`)
    return response.data
  },

  /**
   * Load previous participants for a class (from previous class level/code)
   */
  async loadPreviousKelasPeserta(
    id: string | number,
    params: { page?: number; limit?: number } = {},
    signal?: AbortSignal
  ): Promise<KelasPesertaPreviousListResponse> {
    const response = await apiClient.get<KelasPesertaPreviousListResponse>(`/v1/kelas/${id}/peserta/load-previous`, {
      params: { page: params.page || 1, limit: params.limit || 10 },
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  /**
   * Bulk insert participants into a class
   */
  async createKelasPesertaBulk(
    payload: KelasPesertaBulkPayload
  ): Promise<KelasPesertaBulkResponse> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.post<KelasPesertaBulkResponse>(`/v1/kelas/${kelasId}/peserta/bulk`, payload)
    return response.data
  },

  // --- KelasPengabdi API ---
  async getKelasPengabdi(
    kelasId: string | number,
    params: { page?: number; limit?: number } = {},
    signal?: AbortSignal
  ): Promise<KelasPengabdiListResponse> {
    const response = await apiClient.get<KelasPengabdiListResponse>(`/v1/kelas/${kelasId}/pengabdi`, {
      params: { page: params.page || 1, limit: params.limit || 10 },
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  async createKelasPengabdi(payload: Partial<KelasPengabdi>): Promise<{ data: KelasPengabdi }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.post<{ data: KelasPengabdi }>(`/v1/kelas/${kelasId}/pengabdi`, payload)
    return response.data
  },

  async updateKelasPengabdi(
    id: string | number,
    payload: Partial<KelasPengabdi>
  ): Promise<{ data: KelasPengabdi }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.patch<{ data: KelasPengabdi }>(`/v1/kelas/${kelasId}/pengabdi/${id}`, payload)
    return response.data
  },

  async deleteKelasPengabdi(id: string | number, kelasId: string | number = 0): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/kelas/${kelasId}/pengabdi/${id}`)
    return response.data
  },

  // --- KelasTopik API ---
  async getKelasTopik(
    kelasId: string | number,
    params: { page?: number; limit?: number } = {},
    signal?: AbortSignal
  ): Promise<KelasTopikListResponse> {
    const response = await apiClient.get<KelasTopikListResponse>(`/v1/kelas/${kelasId}/topik`, {
      params: { page: params.page || 1, limit: params.limit || 10 },
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  async createKelasTopik(payload: Partial<KelasTopik>): Promise<{ data: KelasTopik }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.post<{ data: KelasTopik }>(`/v1/kelas/${kelasId}/topik`, payload)
    return response.data
  },

  async updateKelasTopik(
    id: string | number,
    payload: Partial<KelasTopik>
  ): Promise<{ data: KelasTopik }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.patch<{ data: KelasTopik }>(`/v1/kelas/${kelasId}/topik/${id}`, payload)
    return response.data
  },

  async deleteKelasTopik(id: string | number, kelasId: string | number = 0): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/kelas/${kelasId}/topik/${id}`)
    return response.data
  },

  // --- KelasDonasi API ---
  async getKelasDonasi(
    kelasId: string | number,
    params: { page?: number; limit?: number } = {},
    signal?: AbortSignal
  ): Promise<KelasDonasiListResponse> {
    const response = await apiClient.get<KelasDonasiListResponse>(`/v1/kelas/${kelasId}/donasi`, {
      params: { page: params.page || 1, limit: params.limit || 10 },
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  async createKelasDonasi(payload: Partial<KelasDonasi>): Promise<{ data: KelasDonasi }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.post<{ data: KelasDonasi }>(`/v1/kelas/${kelasId}/donasi`, payload)
    return response.data
  },

  async updateKelasDonasi(
    id: string | number,
    payload: Partial<KelasDonasi>
  ): Promise<{ data: KelasDonasi }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.patch<{ data: KelasDonasi }>(`/v1/kelas/${kelasId}/donasi/${id}`, payload)
    return response.data
  },

  async deleteKelasDonasi(id: string | number, kelasId: string | number = 0): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/kelas/${kelasId}/donasi/${id}`)
    return response.data
  },

  // --- KelasDonasiBarang API ---
  async getKelasDonasiBarang(
    kelasId: string | number,
    params: { page?: number; limit?: number } = {},
    signal?: AbortSignal
  ): Promise<KelasDonasiBarangListResponse> {
    const response = await apiClient.get<KelasDonasiBarangListResponse>(`/v1/kelas/${kelasId}/donasi-barang`, {
      params: { page: params.page || 1, limit: params.limit || 10 },
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  async createKelasDonasiBarang(payload: Partial<KelasDonasiBarang>): Promise<{ data: KelasDonasiBarang }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.post<{ data: KelasDonasiBarang }>(`/v1/kelas/${kelasId}/donasi-barang`, payload)
    return response.data
  },

  async updateKelasDonasiBarang(
    id: string | number,
    payload: Partial<KelasDonasiBarang>
  ): Promise<{ data: KelasDonasiBarang }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.patch<{ data: KelasDonasiBarang }>(`/v1/kelas/${kelasId}/donasi-barang/${id}`, payload)
    return response.data
  },

  async deleteKelasDonasiBarang(id: string | number, kelasId: string | number = 0): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/kelas/${kelasId}/donasi-barang/${id}`)
    return response.data
  },

  // --- KelasKendaraan API ---
  async getKelasKendaraan(
    kelasId: string | number,
    params: { page?: number; limit?: number } = {},
    signal?: AbortSignal
  ): Promise<KelasKendaraanListResponse> {
    const response = await apiClient.get<KelasKendaraanListResponse>(`/v1/kelas/${kelasId}/kendaraan`, {
      params: { page: params.page || 1, limit: params.limit || 10 },
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  async createKelasKendaraan(payload: Partial<KelasKendaraan>): Promise<{ data: KelasKendaraan }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.post<{ data: KelasKendaraan }>(`/v1/kelas/${kelasId}/kendaraan`, payload)
    return response.data
  },

  async updateKelasKendaraan(
    id: string | number,
    payload: Partial<KelasKendaraan>
  ): Promise<{ data: KelasKendaraan }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.patch<{ data: KelasKendaraan }>(`/v1/kelas/${kelasId}/kendaraan/${id}`, payload)
    return response.data
  },

  async deleteKelasKendaraan(id: string | number, kelasId: string | number = 0): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/kelas/${kelasId}/kendaraan/${id}`)
    return response.data
  },

  // --- KelasPengeluaran API ---
  async getKelasPengeluaran(
    kelasId: string | number,
    params: { page?: number; limit?: number } = {},
    signal?: AbortSignal
  ): Promise<KelasPengeluaranListResponse> {
    const response = await apiClient.get<KelasPengeluaranListResponse>(`/v1/kelas/${kelasId}/pengeluaran`, {
      params: { page: params.page || 1, limit: params.limit || 10 },
      fetchOptions: {
        priority: 'high'
      },
      signal
    })
    return response.data
  },

  async createKelasPengeluaran(payload: Partial<KelasPengeluaran>): Promise<{ data: KelasPengeluaran }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.post<{ data: KelasPengeluaran }>(`/v1/kelas/${kelasId}/pengeluaran`, payload)
    return response.data
  },

  async updateKelasPengeluaran(
    id: string | number,
    payload: Partial<KelasPengeluaran>
  ): Promise<{ data: KelasPengeluaran }> {
    const kelasId = payload.trx_id || 0
    const response = await apiClient.patch<{ data: KelasPengeluaran }>(`/v1/kelas/${kelasId}/pengeluaran/${id}`, payload)
    return response.data
  },

  async deleteKelasPengeluaran(id: string | number, kelasId: string | number = 0): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/kelas/${kelasId}/pengeluaran/${id}`)
    return response.data
  }
}

export default kelasApi
