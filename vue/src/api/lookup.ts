import apiClient from './client'
import type { LookupQueryParams, LookupListResponse } from '../types/lookup'

export const lookupApi = {
  /**
   * Waktu Ciu Tao Lookup (Category B_WAKTUCIUTAO)
   */
  async getLookupWaktuCiuTao(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/waktu-ciu-tao', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Gender Lookup (Category B_GENDER)
   */
  async getLookupGender(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/gender', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * TCS Lookup (Category B_TCS)
   */
  async getLookupTcs(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/tcs', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fotang Lookup (Category B_FOTANG)
   */
  async getLookupFotang(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/fotang', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Kelas Lookup (Category B_KELAS / B_KELASKHUSUS)
   */
  async getLookupKelas(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/kelas', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Pendidikan Lookup (Category B_PENDIDIKAN)
   */
  async getLookupPendidikan(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/pendidikan', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Kelas Umum Lookup (Category B_KELASUMUM)
   */
  async getLookupKelasUmum(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/kelas-umum', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Pekerjaan Lookup (Category B_PEKERJAAN)
   */
  async getLookupPekerjaan(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/pekerjaan', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Keluarga Lookup (Category B_KELUARGA)
   */
  async getLookupKeluarga(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/keluarga', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Status Lookup (Category B_STATUS)
   */
  async getLookupStatus(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/status', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Kelas Level Lookup (Category B_KLS_LEVEL)
   */
  async getLookupKelasLevel(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/lookup/kelas-level', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Tim Kerja Lookup (/v1/tim-kerja/lookup)
   */
  async getLookupTimKerja(
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>('/v1/tim-kerja/lookup', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Sub Kerja Lookup (/v1/tim-kerja/lookup/:id/sub)
   */
  async getLookupSubKerja(
    timKerjaId: number | string,
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>(`/v1/tim-kerja/lookup/${timKerjaId}/sub`, {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Dynamic Lookup by Category ID
   */
  async getLookupByCategory(
    categoryID: string,
    params: LookupQueryParams = {},
    signal?: AbortSignal
  ): Promise<LookupListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }
    if (params.lookup_description) cleanParams.lookup_description = params.lookup_description

    const response = await apiClient.get<LookupListResponse>(`/v1/lookup/category/${categoryID}`, {
      params: cleanParams,
      signal
    })
    return response.data
  }
}

export default lookupApi
