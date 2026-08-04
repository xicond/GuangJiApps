import apiClient from './client'
import type {
  DonasiSxy,
  DonasiSxyQueryParams,
  DonasiSxyListResponse,
  DonasiSxySingleResponse,
  SxyReportQueryParams,
  SxyReportListResponse
} from '../types/donasiSxy'

export const donasiSxyApi = {
  /**
   * Fetch paginated list of Donasi Sxy with optional search filters.
   */
  async getDonasiSxys(
    params: DonasiSxyQueryParams = {},
    signal?: AbortSignal
  ): Promise<DonasiSxyListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

    if (params.no_kwitansi) cleanParams.no_kwitansi = params.no_kwitansi
    if (params.start_date) cleanParams.start_date = params.start_date
    if (params.end_date) cleanParams.end_date = params.end_date
    if (params.donatur) cleanParams.donatur = params.donatur

    const response = await apiClient.get<DonasiSxyListResponse>('/v1/donasi-sxy', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single DonasiSxy details by ID.
   */
  async getDonasiSxyById(
    id: number,
    signal?: AbortSignal
  ): Promise<DonasiSxySingleResponse> {
    const response = await apiClient.get<DonasiSxySingleResponse>(`/v1/donasi-sxy/${id}`, { signal })
    return response.data
  },

  /**
   * Create a new DonasiSxy record.
   */
  async createDonasiSxy(payload: Partial<DonasiSxy>): Promise<DonasiSxySingleResponse> {
    const response = await apiClient.post<DonasiSxySingleResponse>('/v1/donasi-sxy', payload)
    return response.data
  },

  /**
   * Update an existing DonasiSxy record.
   */
  async updateDonasiSxy(
    id: number,
    payload: Partial<DonasiSxy>
  ): Promise<DonasiSxySingleResponse> {
    const response = await apiClient.patch<DonasiSxySingleResponse>(`/v1/donasi-sxy/${id}`, payload)
    return response.data
  },

  /**
   * Delete a DonasiSxy record by ID.
   */
  async deleteDonasiSxy(id: number): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/v1/donasi-sxy/${id}`)
    return response.data
  },

  /**
   * Fetch paginated Sxy Report.
   */
  async getSxyReport(
    params: SxyReportQueryParams = {},
    signal?: AbortSignal
  ): Promise<SxyReportListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.donatur) cleanParams.donatur = params.donatur
    if (params.penggalang) cleanParams.penggalang = params.penggalang
    if (params.start_date) cleanParams.start_date = params.start_date
    if (params.end_date) cleanParams.end_date = params.end_date
    if (params.fotang) cleanParams.fotang = params.fotang

    const response = await apiClient.get<SxyReportListResponse>('/v1/donasi-sxy/report', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Download Sxy Report Excel binary file.
   */
  async downloadSxyReportExcel(
    params: SxyReportQueryParams = {},
    signal?: AbortSignal
  ): Promise<Blob> {
    const cleanParams: Record<string, any> = {}
    if (params.donatur) cleanParams.donatur = params.donatur
    if (params.penggalang) cleanParams.penggalang = params.penggalang
    if (params.start_date) cleanParams.start_date = params.start_date
    if (params.end_date) cleanParams.end_date = params.end_date
    if (params.fotang) cleanParams.fotang = params.fotang

    const response = await apiClient.get<Blob>('/v1/donasi-sxy/report/excel', {
      params: cleanParams,
      responseType: 'blob',
      signal
    })
    return response.data
  }
}

export default donasiSxyApi

