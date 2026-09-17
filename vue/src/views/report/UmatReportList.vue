<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Laporan Master Umat</h2>
        <p class="page-subtitle">Daftar laporan master data Umat dan download file Excel</p>
      </div>
      <el-tooltip v-model:visible="isVisible" :trigger="''" content="Do fill filter first"
        :disabled="!isDownloadTooltipActive" placement="bottom">
        <span ref="triggerRef" class="inline-block cursor-pointer download-btn-wrapper" @mouseenter="onMouseEnter"
          @mouseleave="onMouseLeave" @touchstart.prevent="onClick">
          <el-button type="success" size="large" :icon="Download" class="download-btn"
            :disabled="dataList.length === 0 || loading || isDownloading || !hasActiveFilter"
            @click="handleDownloadExcel">
            Download Report Excel
          </el-button>
        </span>
      </el-tooltip>
    </div>

    <!-- Filter Card -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-header">
        <el-icon class="filter-icon">
          <Search />
        </el-icon>
        <span class="filter-title">Filter & Pencarian Data Umat</span>
      </div>

      <el-form label-position="top">
        <el-row :gutter="16" class="filter-row">
          <el-col :xs="24" :sm="12" :md="6">
            <el-form-item label="Fotang Aktif">
              <LookupSelect v-model="filters.fotang_aktif" placeholder="Cari Fotang Aktif"
                :fetch-api="fotangApi.getFotangLookup" value-key="lookup_value" label-key="lookup_description" clearable
                @change="onFilterChange" />
            </el-form-item>
          </el-col>

          <el-col :xs="24" :sm="12" :md="6">
            <el-form-item label="Fotang Chiu Tao">
              <LookupSelect v-model="filters.fotang_chiutao" placeholder="Cari Fotang Chiu Tao"
                :fetch-api="fotangApi.getFotangLookup" value-key="lookup_value" label-key="lookup_description" clearable
                @change="onFilterChange" />
            </el-form-item>
          </el-col>

          <el-col :xs="24" :sm="12" :md="6">
            <el-form-item label="Status Umat">
              <LookupSelect v-model="filters.status_umat" placeholder="Pilih Status Umat"
                :fetch-api="lookupApi.getLookupStatus" value-key="lookup_value" label-key="lookup_description" clearable
                @change="onFilterChange" auto-populate />
            </el-form-item>
          </el-col>

          <el-col :xs="24" :sm="12" :md="6">
            <el-form-item label="Nama Ciu Tao">
              <el-input v-model="filters.nama_indo" placeholder="Cari Nama Ciu Tao..." clearable
                @input="onFilterChange" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16" class="filter-row">
          <el-col :xs="24" :sm="12" :md="6">
            <el-form-item label="Nama Lain">
              <el-input v-model="filters.nama_mandarin" placeholder="Cari Nama Lain..." clearable
                @input="onFilterChange" />
            </el-form-item>
          </el-col>

          <el-col :xs="24" :sm="12" :md="6">
            <el-form-item label="Pengajak">
              <el-input v-model="filters.pengajak" placeholder="Cari Pengajak..." clearable @input="onFilterChange" />
            </el-form-item>
          </el-col>

          <el-col :xs="24" :sm="12" :md="6">
            <el-form-item label="Alias / Pin Yin">
              <el-input v-model="filters.alias" placeholder="Cari Alias / Pin Yin..." clearable
                @input="onFilterChange" />
            </el-form-item>
          </el-col>

          <el-col :xs="24" :sm="12" :md="6">
            <el-form-item label="Rentang Tanggal Chiu Tao">
              <el-date-picker v-model="dateRange" type="daterange" range-separator="s/d" start-placeholder="Tgl Mulai"
                end-placeholder="Tgl Selesai" value-format="YYYY-MM-DD" clearable style="width: 100%"
                @change="onDateRangeChange" :single-panel="isMobile" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16" class="filter-row">
          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="Rentang Usia (Tahun)">
              <div class="age-range-container">
                <el-input-number v-model="filters.usia_dari" :min="1"
                  :max="filters.usia_sampai != null ? filters.usia_sampai : 120" controls-position="right"
                  placeholder="Dari" class="age-input" @change="onUsiaDariChange" />
                <span class="range-separator">s/d</span>
                <el-input-number v-model="filters.usia_sampai" :min="filters.usia_dari != null ? filters.usia_dari : 0"
                  :max="120" controls-position="right" placeholder="Sampai" class="age-input"
                  @change="onUsiaSampaiChange" />
              </div>
            </el-form-item>
          </el-col>

          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="Lulus SD3">
              <el-select v-model="filters.is_lulus_sd" placeholder="Semua" clearable style="width: 100%"
                @change="onFilterChange">
                <el-option label="Semua" value="0" />
                <el-option label="Ya" value="1" />
                <el-option label="Tidak" value="2" />
              </el-select>
            </el-form-item>
          </el-col>

          <el-col :xs="24" :sm="12" :md="8">
            <el-form-item label="Vege / Qing Kou">
              <el-select v-model="filters.is_vege" placeholder="Semua" clearable style="width: 100%"
                @change="onFilterChange">
                <el-option label="Semua" value="0" />
                <el-option label="Ya" value="1" />
                <el-option label="Tidak" value="2" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <div class="filter-actions">
        <el-button :icon="Refresh" @click="resetFilters">Reset Filter</el-button>
      </div>
    </el-card>

    <!-- Table Card -->
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="dataList" stripe border height="550" style="width: 100%"
        empty-text="Tidak ada data laporan umat yang ditemukan">
        <el-table-column label="No" v-if="isDesktop" :index="getRowIndex" type="index" width="80" fixed="left" />


        <el-table-column prop="nama_indonesia" label="Nama Chiu Tao" min-width="125"
          :fixed="!isDesktop ? false : 'left'">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.nama_indonesia.trim() || row.nama_mandarin.trim() || row.alias.trim() ||
              '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="kode" v-if="isDesktop" label="Kode Umat" width="120" align="center" />
        <el-table-column prop="jenis_kelamin" label="Jenis Kelamin" width="120" align="center" />

        <el-table-column prop="nama_mandarin" label="Nama Lain" width="140" align="center">
          <template #default="{ row }">
            <span>{{ row.nama_mandarin || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="alias" label="Alias / Pin Yin / Pin Yin" width="120">
          <template #default="{ row }">
            <span>{{ row.alias || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="tanggal_chiu_tao_int" label="Tgl Chiu Tao Masehi" width="120" align="center">
          <template #default="{ row }">
            <span>{{ row.tanggal_chiu_tao_int || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="tanggal_chiu_tao_man" label="Tgl Chiu Tao Mandarin" width="120" align="center">
          <template #default="{ row }">
            <span>{{ row.tanggal_chiu_tao_man || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="tahun_chiu_tao_mandarin" label="Tahun Chiu Tao (Man)" width="150" align="center">
          <template #default="{ row }">
            <span>{{ row.tahun_chiu_tao_mandarin || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="waktu_chiu_tao_mandarin" label="Waktu Chiu Tao" width="130" align="center">
          <template #default="{ row }">
            <span>{{ row.waktu_chiu_tao_mandarin || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="alamat" label="Alamat" min-width="200">
          <template #default="{ row }">
            <span>{{ row.alamat + ' ' + row.alamat2 || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="wilayah" label="Wilayah" min-width="200">
          <template #default="{ row }">
            <span>{{ row.wilayah || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="usia_thn" label="Usia" width="80" align="center">
          <template #default="{ row }">
            <span>{{ row.usia_thn || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="pengajak_manual" label="Pengajak" width="150">
          <template #default="{ row }">
            <span>{{ row.pengajak_manual || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="penanggung_manual" label="Penanggung" width="150">
          <template #default="{ row }">
            <span>{{ row.penanggung_manual || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="tcs" label="TCS" width="120" align="center">
          <template #default="{ row }">
            <span>{{ row.tcs || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="telepon" label="Telepon" width="120" align="center">
          <template #default="{ row }">
            <span>{{ row.telepon || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="mobile" label="Mobile" width="120" align="center">
          <template #default="{ row }">
            <span>{{ row.mobile || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="fotang_ciu_tao_desc" label="Fotang Chiu Tao" width="150" align="center">
          <template #default="{ row }">
            <span>{{ row.fotang_ciu_tao_desc || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="fotang_aktif_desc" label="Fotang Aktif" width="150" align="center">
          <template #default="{ row }">
            <span>{{ row.fotang_aktif_desc || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="pekerjaan_desc" label="Pekerjaan" width="140">
          <template #default="{ row }">
            <span>{{ row.pekerjaan_desc || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="pendidikan_desc" label="Pendidikan" width="130">
          <template #default="{ row }">
            <span>{{ row.pendidikan_desc || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="status_umat_desc" label="Status Umat" width="130" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status_umat_desc)" size="small">
              {{ row.status_umat_desc || '-' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="keterangan" label="Keterangan" min-width="180">
          <template #default="{ row }">
            <span>{{ row.keterangan || '-' }}</span>
          </template>
        </el-table-column>

        <!-- <el-table-column prop="tanggal_lahir" label="Tgl Lahir" width="120" align="center">
          <template #default="{ row }">
            <span>{{ row.tanggal_lahir || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="tempat_lahir" label="Tempat Lahir" width="130">
          <template #default="{ row }">
            <span>{{ row.tempat_lahir || '-' }}</span>
          </template>
        </el-table-column>


        <el-table-column prop="email" label="Email" min-width="160">
          <template #default="{ row }">
            <span>{{ row.email || '-' }}</span>
          </template>
        </el-table-column> -->

        <el-table-column prop="tanggal_sd3" label="Tgl SD3" width="120" align="center">
          <template #default="{ row }">
            <span>{{ row.tanggal_sd3 || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="tempat_sd3_desc" label="Tempat SD3" width="140">
          <template #default="{ row }">
            <span>{{ row.tempat_sd3_desc || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="ikrar1" label="Ikrar 1" width="140">
          <template #default="{ row }">
            <span>{{ row.ikrar1 || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="ikrar2" label="Ikrar 2" width="140">
          <template #default="{ row }">
            <span>{{ row.ikrar2 || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="ikrar3" label="Ikrar 3" width="140">
          <template #default="{ row }">
            <span>{{ row.ikrar3 || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="ikrar4" label="Ikrar 4" width="140">
          <template #default="{ row }">
            <span>{{ row.ikrar4 || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="ikrar5" label="Ikrar 5" width="140">
          <template #default="{ row }">
            <span>{{ row.ikrar5 || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="ikrar6" label="Ikrar 6" width="140">
          <template #default="{ row }">
            <span>{{ row.ikrar6 || '-' }}</span>
          </template>
        </el-table-column>

        <!-- <el-table-column prop="kelas_umum_desc" label="Kelas Umum" width="140">
          <template #default="{ row }">
            <span>{{ row.kelas_umum_desc || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="kelas_khusus_desc" label="Kelas Khusus" width="140">
          <template #default="{ row }">
            <span>{{ row.kelas_khusus_desc || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="tanggal_ching_khou" label="Tgl Qing Kou" width="130" align="center">
          <template #default="{ row }">
            <span>{{ row.tanggal_ching_khou || '-' }}</span>
          </template>
        </el-table-column> -->



        <!-- <el-table-column prop="uang_pahala" label="Uang Pahala" width="130" align="right">
          <template #default="{ row }">
            <span>{{ formatCurrency(row.uang_pahala) }}</span>
          </template>
        </el-table-column> -->
      </el-table>

      <!-- Pagination -->
      <div class="pagination-container">
        <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.limit"
          :page-sizes="[10, 20, 50, 100]" :total="pagination.total"
          :layout="(!isMobile ? 'total, ->,' : (Math.ceil(pagination.total / pagination.limit) < 6 ? '-> ,' : '')) + 'prev, pager, next, jumper'"
          :pager-count="isMobile ? 3 : 6" @size-change="handleSizeChange" @current-change="handlePageChange" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, reactive, computed, onMounted } from 'vue'
import { onClickOutside, useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import { ElMessage, ElLoading } from 'element-plus'
import { Search, Refresh, Download } from '@element-plus/icons-vue'
import umatApi from '../../api/umat'
import fotangApi from '../../api/fotang'
import lookupApi from '../../api/lookup'
import type { UmatReportItem } from '../../types/umat'
import LookupSelect from '../../components/common/LookupSelect.vue'

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)
// const  { width } = useWindowSize()

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

let deviceType: string = ""
const getDeviceType = (): string => {
  const ua = navigator.userAgent
  const maxTouchPoints = navigator.maxTouchPoints || 0

  // 1. Deteksi iPadOS modern (UA mengandung Macintosh/MacIntel, tapi mendukung multi-touch)
  const isIPadOS = /Macintosh/i.test(ua) && maxTouchPoints > 1

  // 2. Deteksi Tablet umum (Android tablet, Playbook, Silk, atau iPadOS)
  if (/tablet|playbook|silk/i.test(ua) || isIPadOS) {
    return 'tablet'
  }

  // 3. Deteksi Mobile (Android Phone, iPhone, iPod, dll)
  if (/mobi|android|iphone|ipod/i.test(ua)) {
    // ElMessage.warning('Mobile detected')
    return 'mobile'
  }

  // 4. Selebihnya dianggap Desktop (termasuk Mac asli yang maxTouchPoints-nya 0)
  // ElMessage.warning('Desktop detected')
  return 'desktop'
}

onMounted(() => {
  getDeviceType()
})

// 1. Deteksi apakah device mendukung hover (Desktop dengan mouse/trackpad)
const canHover = (): boolean => {
  return isDesktop.value || (deviceType === 'desktop')
}

const isVisible = ref(false)
const triggerRef = ref<HTMLElement | null>(null)

// 2. Handlers untuk Desktop (Hover)
const onMouseEnter = () => {
  if (canHover()) isVisible.value = true
}

const onMouseLeave = () => {
  if (canHover()) isVisible.value = false
}

// 3. Handler untuk Mobile / Touch (Click/Tap)
const onClick = () => {
  if (!canHover()) {
    if (isVisible.value = !isVisible.value) {
      setTimeout(() => {
        isVisible.value = false
      }, 2100)
    }
  }
}

// 4. Tutup otomatis saat area luar disentuh/diklik (Khusus Mobile)
onClickOutside(triggerRef, () => {
  if (!canHover()) isVisible.value = false
})

const dataList = shallowRef<UmatReportItem[]>([])
const loading = ref(false)
const isDownloading = ref(false)

const pagination = reactive({
  page: 1,
  limit: 10,
  total: 0
})

const getRowIndex = (index: number) => (pagination.page - 1) * pagination.limit + index + 1

const filters = reactive({
  fotang_aktif: '',
  fotang_chiutao: '',
  nama_mandarin: '',
  start_date: '',
  end_date: '',
  nama_indo: '',
  pengajak: '',
  alias: '',
  usia_dari: null as number | null,
  usia_sampai: null as number | null,
  is_lulus_sd: '0',
  is_vege: '0',
  status_umat: ''
})

const hasActiveFilter = computed(() => {
  return (
    !!filters.fotang_aktif ||
    !!filters.fotang_chiutao ||
    !!filters.nama_mandarin.trim() ||
    !!filters.start_date ||
    !!filters.end_date ||
    !!filters.nama_indo.trim() ||
    !!filters.pengajak.trim() ||
    !!filters.alias.trim() ||
    filters.usia_dari != null ||
    filters.usia_sampai != null ||
    (filters.is_lulus_sd !== '0' && !!filters.is_lulus_sd) ||
    (filters.is_vege !== '0' && !!filters.is_vege) ||
    !!filters.status_umat
  )
})

const isDownloadTooltipActive = computed(() => {
  return dataList.value.length > 0 && !hasActiveFilter.value && !isDownloading.value
})

const dateRange = computed({
  get: () => {
    if (filters.start_date && filters.end_date) {
      return [filters.start_date, filters.end_date]
    }
    return []
  },
  set: (val: [string, string] | null) => {
    if (val && val.length === 2) {
      filters.start_date = val[0]
      filters.end_date = val[1]
    } else {
      filters.start_date = ''
      filters.end_date = ''
    }
  }
})

let currentAbortController: AbortController | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null

function formatCurrency(val?: number) {
  if (val == null || isNaN(val)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(val)
}

function getStatusTagType(statusDesc?: string) {
  if (!statusDesc) return 'info'
  const lower = statusDesc.toLowerCase()
  if (lower.includes('aktif')) return 'success'
  if (lower.includes('pindah')) return 'warning'
  if (lower.includes('meninggal')) return 'danger'
  return 'info'
}

async function fetchData() {
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()
  loading.value = true

  try {
    const res = await umatApi.getUmatReport(
      {
        page: pagination.page,
        limit: pagination.limit,
        fotang_aktif: filters.fotang_aktif,
        fotang_chiutao: filters.fotang_chiutao,
        nama_mandarin: filters.nama_mandarin,
        start_date: filters.start_date,
        end_date: filters.end_date,
        nama_indo: filters.nama_indo,
        pengajak: filters.pengajak,
        alias: filters.alias,
        usia_dari: filters.usia_dari ?? undefined,
        usia_sampai: filters.usia_sampai ?? undefined,
        is_lulus_sd: filters.is_lulus_sd,
        is_vege: filters.is_vege,
        status_umat: filters.status_umat
      },
      currentAbortController.signal
    )
    dataList.value = res.data || []
    pagination.total = res.meta?.total || 0
  } catch (err: any) {
    if (err.name === 'CanceledError' || err.name === 'AbortError') return
    console.error('Error fetching umat report:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal memuat data laporan Umat')
  } finally {
    loading.value = false
  }
}

function onFilterChange() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    pagination.page = 1
    fetchData()
  }, 300)
}

function onUsiaDariChange(val: number | null) {
  if (val != null && filters.usia_sampai != null && filters.usia_sampai < val) {
    filters.usia_sampai = val
  }
  onFilterChange()
}

function onUsiaSampaiChange(val: number | null) {
  if (val != null && filters.usia_dari != null && filters.usia_dari > val) {
    filters.usia_dari = val
  }
  onFilterChange()
}

function onDateRangeChange() {
  pagination.page = 1
  fetchData()
}

function resetFilters() {
  filters.fotang_aktif = ''
  filters.fotang_chiutao = ''
  filters.nama_mandarin = ''
  filters.start_date = ''
  filters.end_date = ''
  filters.nama_indo = ''
  filters.pengajak = ''
  filters.alias = ''
  filters.usia_dari = null
  filters.usia_sampai = null
  filters.is_lulus_sd = '0'
  filters.is_vege = '0'
  filters.status_umat = ''
  pagination.page = 1
  fetchData()
}

function handleSizeChange(val: number) {
  pagination.limit = val
  pagination.page = 1
  fetchData()
}

function handlePageChange(val: number) {
  pagination.page = val
  fetchData()
}

async function handleDownloadExcel() {
  if (!hasActiveFilter.value) {
    ElMessage.warning('Setidaknya satu filter harus diisi untuk mendownload report')
    return
  }

  if (dataList.value.length === 0) {
    ElMessage.warning('Tidak ada data untuk di-download')
    return
  }

  isDownloading.value = true
  const loadingInstance = ElLoading.service({
    lock: true,
    text: 'Downloading Report Excel ...',
    background: 'rgba(0, 0, 0, 0.7)'
  })

  try {
    const blob = await umatApi.downloadUmatReportExcel({
      fotang_aktif: filters.fotang_aktif,
      fotang_chiutao: filters.fotang_chiutao,
      nama_mandarin: filters.nama_mandarin,
      start_date: filters.start_date,
      end_date: filters.end_date,
      nama_indo: filters.nama_indo,
      pengajak: filters.pengajak,
      alias: filters.alias,
      usia_dari: filters.usia_dari ?? undefined,
      usia_sampai: filters.usia_sampai ?? undefined,
      is_lulus_sd: filters.is_lulus_sd,
      is_vege: filters.is_vege,
      status_umat: filters.status_umat
    })

    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `laporan_umat_${new Date().toISOString().slice(0, 10)}.xlsx`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    ElMessage.success('Berhasil mengunduh Laporan Umat Excel')
  } catch (err: any) {
    console.error('Error downloading Excel:', err)
    ElMessage.error(err.message || 'Gagal mengunduh file Excel')
  } finally {
    loadingInstance.close()
    isDownloading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.view-container {
  display: flex;
  flex-direction: column;
  /* padding: 24px; */
  max-width: 100%;
  gap: 1.25rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--el-text-color-primary);
  margin: 0;
}

.page-subtitle {
  font-size: 0.875rem;
  color: var(--el-text-color-secondary);
  margin: 0;
}

.filter-card {
  margin-bottom: 24px;
  border-radius: 8px;
}

.filter-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.filter-icon {
  font-size: 18px;
  color: var(--el-color-primary);
}

.filter-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.filter-row {
  margin-bottom: 8px;
}

.filter-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.table-card {
  border-radius: 8px;
}

.pagination-container {
  display: flex;
  /* justify-content: flex-end;
  margin-top: 20px; */
}

.font-semibold {
  font-weight: 600;
}

.age-range-container {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.age-input {
  flex: 1;
  width: 100%;
}

.range-separator {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
}

.download-btn-wrapper {
  display: inline-block;
}
</style>
