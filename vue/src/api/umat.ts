import apiClient from './client'
import type {
  Umat,
  UmatQueryParams,
  UmatListResponse,
  UmatSingleResponse,
  UmatReportQueryParams,
  UmatReportResponse
} from '../types/umat'

export const umatApi = {
  /**
   * Fetch paginated list of umats with optional search filters.
   * Supports AbortSignal for canceling outdated requests.
   */
  async getUmats(
    params: UmatQueryParams = {},
    signal?: AbortSignal
  ): Promise<UmatListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.alias) cleanParams.alias = params.alias
    if (params.namaindonesia) cleanParams.namaindonesia = params.namaindonesia
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description
    if (params.namamandarin) cleanParams.namamandarin = params.namamandarin
    if (params.tahunchiutaomandarin) {
      cleanParams.tahunchiutaomandarin = params.tahunchiutaomandarin
    }

    const response = await apiClient.get<UmatListResponse>('/v1/umats', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single Umat details by ID.
   */
  async getUmatById(
    id: number | string,
    signal?: AbortSignal
  ): Promise<UmatSingleResponse> {
    const response = await apiClient.get<UmatSingleResponse>(`/v1/umats/${id}`, { signal })
    return response.data
  },

  /**
   * Create a new Umat record with optional photo file.
   */
  async createUmat(
    payload: Partial<Umat>,
    photoFile?: File | Blob | null
  ): Promise<UmatSingleResponse> {
    if (photoFile) {
      const formData = new FormData()
      formData.append('data', JSON.stringify(payload))
      formData.append('foto', photoFile, (photoFile as File).name || 'foto.jpg')
      const response = await apiClient.post<UmatSingleResponse>('/v1/umats', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      return response.data
    }
    const response = await apiClient.post<UmatSingleResponse>('/v1/umats', payload)
    return response.data
  },

  /**
   * Perform OCR image recognition on an uploaded document/image file.
   */
  async ocrUmat(file: File): Promise<{ data: any; resource: string }> {
    const formData = new FormData()
    formData.append('file', file, file.name)
    const response = await apiClient.post<{ data: any; resource: string }>(
      '/v1/umats/ocr',
      formData,
      {
        headers: { 'Content-Type': 'multipart/form-data' }
      }
    )
    return response.data
  },

  /**
   * Update an existing Umat record with optional photo file.
   */
  async updateUmat(
    id: number | string,
    payload: Partial<Umat>,
    photoFile?: File | Blob | null
  ): Promise<UmatSingleResponse> {
    if (photoFile) {
      const formData = new FormData()
      formData.append('data', JSON.stringify(payload))
      formData.append('foto', photoFile, (photoFile as File).name || 'foto.jpg')
      const response = await apiClient.patch<UmatSingleResponse>(`/v1/umats/${id}`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      return response.data
    }
    const response = await apiClient.patch<UmatSingleResponse>(`/v1/umats/${id}`, payload)
    return response.data
  },

  /**
   * Soft/Hard Delete a Umat record by ID.
   */
  async deleteUmat(id: number | string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/umats/${id}`)
    return response.data
  },

  /**
   * Fetch Umat report list with parameters.
   */
  async getUmatReport(
    params: UmatReportQueryParams = {},
    signal?: AbortSignal
  ): Promise<UmatReportResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.fotang_aktif) cleanParams.fotang_aktif = params.fotang_aktif
    if (params.fotang_chiutao) cleanParams.fotang_chiutao = params.fotang_chiutao
    if (params.nama_mandarin) cleanParams.nama_mandarin = params.nama_mandarin
    if (params.start_date) cleanParams.start_date = params.start_date
    if (params.end_date) cleanParams.end_date = params.end_date
    if (params.nama_indo) cleanParams.nama_indo = params.nama_indo
    if (params.pengajak) cleanParams.pengajak = params.pengajak
    if (params.alias) cleanParams.alias = params.alias
    if (params.usia_dari !== undefined && params.usia_dari !== '') cleanParams.usia_dari = params.usia_dari
    if (params.usia_sampai !== undefined && params.usia_sampai !== '') cleanParams.usia_sampai = params.usia_sampai
    if (params.is_lulus_sd) cleanParams.is_lulus_sd = params.is_lulus_sd
    if (params.is_vege) cleanParams.is_vege = params.is_vege
    if (params.status_umat) cleanParams.status_umat = params.status_umat

    const response = await apiClient.get<UmatReportResponse>('/v1/umats/report', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Download Umat report as Excel file blob.
   */
  async downloadUmatReportExcel(
    params: UmatReportQueryParams = {},
    signal?: AbortSignal
  ): Promise<Blob> {
    const cleanParams: Record<string, any> = {}

    if (params.fotang_aktif) cleanParams.fotang_aktif = params.fotang_aktif
    if (params.fotang_chiutao) cleanParams.fotang_chiutao = params.fotang_chiutao
    if (params.nama_mandarin) cleanParams.nama_mandarin = params.nama_mandarin
    if (params.start_date) cleanParams.start_date = params.start_date
    if (params.end_date) cleanParams.end_date = params.end_date
    if (params.nama_indo) cleanParams.nama_indo = params.nama_indo
    if (params.pengajak) cleanParams.pengajak = params.pengajak
    if (params.alias) cleanParams.alias = params.alias
    if (params.usia_dari !== undefined && params.usia_dari !== '') cleanParams.usia_dari = params.usia_dari
    if (params.usia_sampai !== undefined && params.usia_sampai !== '') cleanParams.usia_sampai = params.usia_sampai
    if (params.is_lulus_sd) cleanParams.is_lulus_sd = params.is_lulus_sd
    if (params.is_vege) cleanParams.is_vege = params.is_vege
    if (params.status_umat) cleanParams.status_umat = params.status_umat

    const response = await apiClient.get('/v1/umats/report/excel', {
      params: cleanParams,
      responseType: 'blob',
      signal
    })
    return response.data
  }
}

export default umatApi
