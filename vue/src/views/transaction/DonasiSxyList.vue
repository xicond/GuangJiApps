<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Transaksi Donasi Sxy</h2>
        <p class="page-subtitle">Kelola daftar data donasi sxy, pencarian, serta pembaruan profil</p>
      </div>
      <el-button
        type="primary"
        size="large"
        :icon="Plus"
        class="create-btn"
        @click="handleCreate"
      >
        Tambah Donasi Sxy Baru
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
          <el-form-item label="No. Kwitansi" :label-position="isMobile? 'top' : 'right'">
            <el-input
              v-model="filters.no_kwitansi"
              placeholder="Cari no kwitansi..."
              clearable
              :prefix-icon="Search"
              @input="onFilterChange"
            />
          </el-form-item>
        </el-col>
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
        empty-text="Tidak ada data donasi sxy yang ditemukan"
      >
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="70" align="center" fixed="left" />

        <el-table-column prop="no_kwitansi" label="No. Kwitansi" width="130" align="left">
          <template #default="{ row }">
            <el-tag size="small" type="info" class="font-mono">{{ row.no_kwitansi }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="tanggal" label="Tanggal"  min-width="140"  >
          <template #default="{ row }">
            <span>{{ row.tanggal || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="jumlah" label="Jumlah (Rp)" min-width="150">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.jumlah ? 'Rp ' + Number(row.jumlah).toLocaleString('id-ID') : '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="no_kupon" label="No. Kupon"  min-width="140"  >
          <template #default="{ row }">
            <span>{{ row.no_kupon || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="keterangan" label="Keterangan"  min-width="200"  >
          <template #default="{ row }">
            <span>{{ row.keterangan || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="status" label="Status" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status ? 'success' : 'info'" size="small">
              {{ row.status ? 'Aktif' : 'Nonaktif' }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- Actions Column -->
        <el-table-column label="Aksi" width="150" align="center" :fixed="!isDesktop ? false : 'right'">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                type="primary"
                size="small"
                circle
                :icon="Edit"
                title="Edit Donasi Sxy"
                @click="handleEdit(row.id)"
              />

              <el-popconfirm
                title="Apakah Anda yakin ingin menghapus data ini?"
                confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal"
                confirm-button-type="danger"
                @confirm="handleDelete(row.id)"
              >
                <template #reference>
                  <el-button
                    type="danger"
                    size="small"
                    circle
                    :icon="Delete"
                    title="Hapus Donasi Sxy"
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
import { ElMessage, ElNotification } from 'element-plus'
import {
  Search,
  Refresh,
  Plus,
  Edit,
  Delete
} from '@element-plus/icons-vue'
import { donasiSxyApi } from '../../api/donasiSxy'
import type { DonasiSxy, DonasiSxyQueryParams } from '../../types/donasiSxy'

const router = useRouter()

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

// Memory Optimization: shallowRef for table dataset
const dataList = shallowRef<DonasiSxy[]>([])
const loading = ref(false)

// Pagination state
const pagination = reactive({
  page: 1,
  limit: 10,
  total: 0
})

const getRowIndex = (index: number) => {
  return (pagination.page - 1) * pagination.limit + index + 1
}

// Search Filter state
const filters = reactive<DonasiSxyQueryParams>({
  no_kwitansi: ''
})

let currentAbortController: AbortController | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null

async function fetchData() {
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()
  loading.value = true

  try {
    const res = await donasiSxyApi.getDonasiSxys(
      {
        page: pagination.page,
        limit: pagination.limit,
        no_kwitansi: filters.no_kwitansi?.trim()
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
  filters.no_kwitansi = ''
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

function handleCreate() {
  ElNotification({
    title: 'Informasi',
    message: 'Tambah donasi sxy baru',
    type: 'info'
  })
}

function handleEdit(id: number) {
  ElNotification({
    title: 'Informasi',
    message: `Edit donasi sxy ID/Kode: ${id}`,
    type: 'info'
  })
}

async function handleDelete(id: number) {
  try {
    await donasiSxyApi.deleteDonasiSxy(id)
    ElNotification({
      title: 'Berhasil',
      message: `Data donasi sxy ${id} berhasil dihapus`,
      type: 'success'
    })
    fetchData()
  } catch (err: any) {
    console.error('Failed to delete item:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal menghapus data')
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
