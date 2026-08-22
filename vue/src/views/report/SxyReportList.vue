<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Laporan Sxy</h2>
        <p class="page-subtitle">Daftar laporan transaksi donasi SXY dan download file Excel</p>
      </div>
      <el-button type="success" size="large" :icon="Download" class="download-btn"
        :disabled="dataList.length === 0 || loading || isDownloading || !hasActiveFilter" @click="handleDownloadExcel">
        Downloading Report Excel
      </el-button>
    </div>

    <!-- Filter Card -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-header">
        <el-icon class="filter-icon">
          <Search />
        </el-icon>
        <span class="filter-title">Filter & Pencarian Data</span>
      </div>

      <el-row :gutter="16" class="filter-row">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Donatur" :label-position="isMobile ? 'top' : 'right'">
            <LookupSelect v-model="filters.donatur" placeholder="Pilih Donatur"
              :fetch-api="sxyDonaturApi.getSxyDonaturs" value-key="id" label-key="nama" clearable
              @change="onFilterChange" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Penggalang Dana" :label-position="isMobile ? 'top' : 'right'">
            <LookupSelect v-model="filters.penggalang" placeholder="Pilih Penggalang Dana"
              :fetch-api="penggalangDanaApi.getPenggalangDanas" value-key="id" label-key="nama" clearable
              @change="onFilterChange" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Fotang" :label-position="isMobile ? 'top' : 'right'">
            <LookupSelect v-model="filters.fotang" placeholder="Pilih Fotang" :fetch-api="fotangApi.getFotangLookupSxy"
              value-key="lookup_value" label-key="lookup_description" clearable @change="onFilterChange" />
          </el-form-item>
        </el-col>

      </el-row>
      <el-row :gutter="16" class="filter-row">

        <el-col :xs="24" :sm="12" :md="12">
          <el-form-item label="Rentang Tanggal" :label-position="isMobile ? 'top' : 'right'">
            <el-date-picker v-model="dateRange" type="daterange" range-separator="s/d" start-placeholder="Tgl Mulai"
              end-placeholder="Tgl Selesai" value-format="YYYY-MM-DD" clearable style="width: 100%"
              @change="onDateRangeChange" :single-panel="isMobile" />
          </el-form-item>
        </el-col>
      </el-row>

      <div class="filter-actions">
        <el-button :icon="Refresh" @click="resetFilters">Reset Filter</el-button>
      </div>
    </el-card>

    <!-- Table Card -->
    <el-card shadow="never" class="table-card">

      <el-table v-loading="loading" :data="dataList" stripe border height="500" show-summary
        :summary-method="getSummaries" style="width: 100%" empty-text="Tidak ada data sxy report yang ditemukan">
        <el-table-column prop="no_kwitansi" label="No Kwitansi" width="110" align="center"
          :fixed="!isMobile ? 'left' : false" />

        <el-table-column prop="tanggal" label="Tanggal Transaksi" width="110" align="center">
          <template #default="{ row }">
            <span>{{ row.tanggal || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="tanggal" label="Tanggal Transfer" width="110" align="center">
          <template #default="{ row }">
            <span>{{ row.tanggal_transfer || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="fotang" label="Fotang" width="120" align="center">
          <template #default="{ row }">
            <span>{{ row.fotang || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="donatur" label="Donatur" min-width="160">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.donatur || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="atas_nama" label="Atas Nama" min-width="160">
          <template #default="{ row }">
            <span>{{ row.atas_nama || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="penggalang_dana" label="Penggalang Dana" min-width="160">
          <template #default="{ row }">
            <span>{{ row.penggalang_dana || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="tipe_sumbangan" label="Tipe Sumbangan" width="150" align="center">
          <template #default="{ row }">
            <span>{{ row.tipe_sumbangan || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="jumlah" label="Nominal" width="150" align="right">
          <template #default="{ row }">
            <span class="font-semibold text-success">{{ formatCurrency(row.jumlah) }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="nokupon" label="No Kupon" width="120" align="center">
          <template #default="{ row }">
            <span>{{ row.nokupon || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="keterangan" label="Alamat" min-width="180">
          <template #default="{ row }">
            <span>{{ row.keterangan || '-' }}</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-container">
        <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.limit"
          :page-sizes="[10, 20, 50, 100]" :total="pagination.total"
          :layout="isDesktop ? 'total, sizes, prev, pager, next, jumper' : 'total, sizes, prev, pager, next'"
          @size-change="handleSizeChange" @current-change="handlePageChange" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import { ElMessage, ElNotification, ElLoading } from 'element-plus'
import { Search, Refresh, Download } from '@element-plus/icons-vue'
import donasiSxyApi from '../../api/donasiSxy'
import fotangApi from '../../api/fotang'
import umatApi from '../../api/umat'
import penggalangDanaApi from '../../api/penggalangDana'
import type { SxyDonasiReportItem } from '../../types/donasiSxy'
import LookupSelect from '../../components/common/LookupSelect.vue'
import sxyDonaturApi from '@/api/sxyDonatur'

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

const dataList = shallowRef<SxyDonasiReportItem[]>([])
const loading = ref(false)
const isDownloading = ref(false)
const totalJumlah = ref(0)

const pagination = reactive({
  page: 1,
  limit: 10,
  total: 0
})

const filters = reactive({
  donatur: '',
  penggalang: '',
  start_date: '',
  end_date: '',
  fotang: ''
})

const hasActiveFilter = computed(() => {
  return (
    !!filters.donatur ||
    !!filters.penggalang ||
    !!filters.start_date ||
    !!filters.end_date ||
    !!filters.fotang
  )
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

function getSummaries(param: { columns: any[]; data: any[] }) {
  const { columns } = param
  const sums: string[] = []
  columns.forEach((column, index) => {
    let seq = 0
    if (isMobile.value && column.property === 'tipe_sumbangan') {
      sums[index] = 'Total'
      return
    }
    if (!isMobile.value && index === 0) {
      sums[index] = 'Total'
      return
    }
    if (column.property === 'jumlah') {
      sums[index] = formatCurrency(totalJumlah.value)
      return
    }
    sums[index] = ''
  })
  return sums
}

async function fetchData() {
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()
  loading.value = true

  try {
    const res = await donasiSxyApi.getSxyReport(
      {
        page: pagination.page,
        limit: pagination.limit,
        donatur: filters.donatur,
        penggalang: filters.penggalang,
        start_date: filters.start_date,
        end_date: filters.end_date,
        fotang: filters.fotang
      },
      currentAbortController.signal
    )
    dataList.value = res.data || []
    pagination.total = res.meta?.total || 0
    totalJumlah.value = res.meta?.total_jumlah || 0
  } catch (err: any) {
    if (err.name === 'CanceledError' || err.name === 'AbortError') return
    console.error('Error fetching sxy report:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal memuat data laporan SXY')
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

function onDateRangeChange() {
  pagination.page = 1
  fetchData()
}

function resetFilters() {
  filters.donatur = ''
  filters.penggalang = ''
  filters.start_date = ''
  filters.end_date = ''
  filters.fotang = ''
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
    const blob = await donasiSxyApi.downloadSxyReportExcel({
      donatur: filters.donatur,
      penggalang: filters.penggalang,
      start_date: filters.start_date,
      end_date: filters.end_date,
      fotang: filters.fotang
    })

    const url = window.URL.createObjectURL(new Blob([blob]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `rpt_sxy_transaksi_${new Date().toISOString().slice(0, 10)}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)

    ElNotification({
      title: 'Berhasil',
      message: 'Report Excel SXY berhasil di-download',
      type: 'success'
    })
  } catch (err: any) {
    console.error('Failed to download report excel:', err)
    let errorMessage = err.response?.data?.error || err.message || 'Gagal mendownload report Excel'
    if (err.response?.data instanceof Blob) {
      try {
        const text = await err.response.data.text()
        const parsed = JSON.parse(text)
        if (parsed?.error) errorMessage = parsed.error
      } catch (e) {
        // preserve default errorMessage if not JSON
      }
    }
    ElMessage.error(errorMessage)
  } finally {
    loadingInstance.close()
    isDownloading.value = false
  }
}

onMounted(() => {
  fetchData()
})

onUnmounted(() => {
  if (currentAbortController) currentAbortController.abort()
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<style scoped>
.view-container {
  display: flex;
  flex-direction: column;
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
  margin: 0.25rem 0 0 0;
}

.filter-card {
  border-radius: 8px;
  background-color: var(--el-bg-color);
}

.filter-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.filter-row {
  margin-bottom: -0.5rem;
}

.filter-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 0.5rem;
}

.table-card {
  border-radius: 8px;
}

.summary-banner {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background-color: var(--el-color-success-light-9);
  border: 1px solid var(--el-color-success-light-5);
  padding: 0.75rem 1rem;
  border-radius: 6px;
  margin-bottom: 1rem;
}

.summary-label {
  font-weight: 600;
  color: var(--el-color-success-dark-2);
}

.summary-value {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--el-color-success);
}

.summary-count {
  font-size: 0.875rem;
  color: var(--el-text-color-secondary);
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 1.25rem;
}

.font-semibold {
  font-weight: 600;
}

.text-success {
  color: var(--el-color-success);
}
</style>
