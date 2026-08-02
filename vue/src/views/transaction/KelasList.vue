<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Transaksi Kelas</h2>
        <p class="page-subtitle">Kelola daftar data kelas, pencarian, serta pembaruan profil</p>
      </div>
      <el-button
        type="primary"
        size="large"
        :icon="Plus"
        class="create-btn"
        @click="handleCreate"
      >
        Tambah Kelas Baru
      </el-button>
    </div>

    <!-- Filter Card -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-header">
        <el-icon class="filter-icon"><Search /></el-icon>
        <span class="filter-title">Filter & Pencarian Data</span>
      </div>

      <el-row :gutter="16" class="filter-row">
        <el-col :xs="24" :sm="12" :md="6">
          <el-form-item label="Kelas" :label-position="isMobile? 'top' : 'right'">
            <LookupSelect
              v-model="filters.kelas"
              placeholder="Pilih kelas..."
              :fetch-api="kelasApi.getKelasLookup"
              @change="onFilterChange"
            />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="6">
          <el-form-item label="Fotang" :label-position="isMobile? 'top' : 'right'">
            <LookupSelect
              v-model="filters.fotang"
              placeholder="Pilih fotang..."
              :fetch-api="fotangApi.getFotangLookup"
              @change="onFilterChange"
            />
          </el-form-item>
        </el-col>

        <el-form-item label="Periode" :label-position="isMobile ? 'top' : 'right'">
            <el-date-picker
              v-model="filters.date_range"
              type="daterange"
              range-separator="s/d"
              start-placeholder="Start Date"
              end-placeholder="End Date"
              value-format="YYYY-MM-DD"
              :disabled-date="disabledDate"
              @change="fetchData"
              :clearable="false"
              style="width: 100%"
              :single-panel="isMobile"
            />
          </el-form-item>
      </el-row>

      <div class="filter-actions">
        <el-button :icon="Refresh" @click="resetFilters">Reset Filter</el-button>
      </div>
    </el-card>

    <!-- Table Card -->
    <el-card shadow="never" class="table-card">
      <el-table
        v-loading="loading"
        :data="dataList"
        stripe
        border
        height="475"
        style="width: 100%"
        empty-text="Tidak ada data kelas yang ditemukan"
      >
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="70" align="center" fixed="left" />

        <!-- <el-table-column prop="lookup_id" label="Kode Kelas" width="130" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info" class="font-mono">{{ row.lookup_id }}</el-tag>
          </template>
        </el-table-column> -->

        <el-table-column prop="kelas_desc" label="Nama Kelas" min-width="200">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.kelas_desc || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="fotang_desc" label="Fotang"  min-width="220"  >
          <template #default="{ row }">
            <span>{{ row.fotang_desc || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="start_date" label="Start Date"  min-width="220"  >
          <template #default="{ row }">
            <span>{{ row.start_date || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="end_date" label="End Date"  min-width="220"  >
          <template #default="{ row }">
            <span>{{ row.end_date || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="lokasi" label="Lokasi"  min-width="220"  >
          <template #default="{ row }">
            <span>{{ row.lokasi || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="pic" label="PIC"  min-width="220"  >
          <template #default="{ row }">
            <span>{{ row.pic || '-' }}</span>
          </template>
        </el-table-column>

        <!-- Actions Column -->
        <el-table-column label="Aksi" width="190" align="center" :fixed="!isDesktop ? false : 'right'">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                type="info"
                size="small"
                circle
                :icon="View"
                title="Lihat Detail Kelas & Peserta"
                @click="handleView(row.trx_id)"
              />

              <el-button
                type="primary"
                size="small"
                circle
                :icon="Edit"
                title="Edit Kelas"
                @click="handleEdit(row.trx_id)"
              />

              <el-button
                type="success"
                size="small"
                circle
                :icon="Download"
                title="Download Report Excel"
                :loading="downloadingId === row.trx_id"
                @click="handleDownloadReport(row.trx_id)"
              />

              <el-popconfirm
                title="Apakah Anda yakin ingin menghapus data ini?"
                confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal"
                confirm-button-type="danger"
                @confirm="handleDelete(row.trx_id)"
              >
                <template #reference>
                  <el-button
                    type="danger"
                    size="small"
                    circle
                    :icon="Delete"
                    title="Hapus Kelas"
                  />
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.limit"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          :layout="'total, sizes, prev, pager, next' + (isDesktop ? ', jumper' : '')"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, reactive, onMounted, onUnmounted } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import { useRouter } from 'vue-router'
import { ElMessage, ElNotification, ElLoading } from 'element-plus'
import {
  Search,
  Refresh,
  Plus,
  Edit,
  View,
  Delete,
  Download
} from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { kelasApi } from '../../api/kelas'
import { fotangApi } from '../../api/fotang'
import LookupSelect from '../../components/common/LookupSelect.vue'
import type { Kelas, KelasQueryParams } from '../../types/kelas'

const router = useRouter()

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

// Memory Optimization: shallowRef for table dataset
const dataList = shallowRef<Kelas[]>([])
const loading = ref(false)
const downloadingId = ref<string | number | null>(null)

// Pagination state
const pagination = reactive({
  page: 1,
  limit: 10,
  total: 0
})

const getRowIndex = (index: number) => {
  return (pagination.page - 1) * pagination.limit + index + 1
}

const disabledDate = (time: Date) => {
  return time.getTime() > Date.now()
}

const today = dayjs().format('YYYY-MM-DD')
type KelasFilter = Omit<KelasQueryParams, 'start_date' | 'end_date'> & {
  kelas: '',
  fotang: ''
  date_range?: string[] | null
}


// Search Filter state
const filters = reactive<KelasFilter>({
  kelas: '',
  date_range: ['2018-07-01', today] as [string, string],
  fotang: ''
})


let currentAbortController: AbortController | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null

async function fetchData() {
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()
  loading.value = true
  const [start_date, end_date] = filters.date_range || [null, null]

  try {
    const res = await kelasApi.getKelasList(
      {
        page: pagination.page,
        limit: pagination.limit,
        kelas: filters.kelas,
        start_date: start_date?start_date:undefined,
        end_date: end_date?end_date:undefined,
        fotang: filters.fotang
      },
      currentAbortController.signal
    )

    dataList.value = res.data || []
    pagination.total = res.meta?.total || 0
  } catch (err: any) {
    if (err.name === 'CanceledError' || err.name === 'AbortError') return
    console.error('Error fetching data:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal memuat data')
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

function resetFilters() {
  filters.kelas = ''
  filters.fotang = ''
  pagination.page = 1
  fetchData()
}

function handleSizeChange(newLimit: number) {
  pagination.limit = newLimit
  pagination.page = 1
  fetchData()
}

function handlePageChange(newPage: number) {
  pagination.page = newPage
  fetchData()
}

function handleView(id: string | number) {
  router.push(`/transaction/kelas/view/${id}`)
}

function handleCreate() {
  router.push('/transaction/kelas/create')
}

function handleEdit(id: string | number) {
  router.push(`/transaction/kelas/edit/${id}`)
}

async function handleDelete(id: string) {
  try {
    await kelasApi.deleteKelas(id)
    ElNotification({
      title: 'Berhasil',
      message: `Data kelas ${id} berhasil dihapus`,
      type: 'success'
    })
    fetchData()
  } catch (err: any) {
    console.error('Failed to delete item:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal menghapus data')
  }
}

async function handleDownloadReport(id: string | number) {
  downloadingId.value = id
  const loadingInstance = ElLoading.service({
    lock: true,
    text: 'Mendownload report Excel, mohon tunggu...',
    background: 'rgba(0, 0, 0, 0.7)'
  })
  try {
    const blob = await kelasApi.downloadReport(id)
    const url = window.URL.createObjectURL(new Blob([blob]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `rpt_trx_kelas_${id}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    ElNotification({
      title: 'Berhasil',
      message: `Report kelas ${id} berhasil di-download`,
      type: 'success'
    })
  } catch (err: any) {
    console.error('Failed to download report:', err)
    let errorMessage = err.response?.data?.error || err.message || 'Gagal mendownload report'
    if (err.response?.data instanceof Blob) {
      try {
        const text = await err.response.data.text()
        const parsed = JSON.parse(text)
        if (parsed?.error) errorMessage = parsed.error
      } catch (e) {
        // preserve default errorMessage if text is not JSON
      }
    }
    ElMessage.error(errorMessage)
  } finally {
    loadingInstance.close()
    downloadingId.value = null
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

.create-btn {
  font-weight: 600;
  border-radius: 8px;
}

.filter-card {
  border-radius: 10px;
  background-color: var(--el-bg-color-overlay);
}

.filter-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.filter-icon {
  color: var(--el-color-primary);
  font-size: 1.1rem;
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
  border-radius: 10px;
  background-color: var(--el-bg-color-overlay);
}

.font-mono {
  font-family: monospace;
}

.font-semibold {
  font-weight: 600;
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 1.25rem;
}
</style>
