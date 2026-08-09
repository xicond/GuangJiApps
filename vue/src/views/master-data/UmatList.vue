<template>
  <div class="umat-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Master Data Umat</h2>
        <p class="page-subtitle">Kelola daftar data umat, pencarian, serta pembaruan profil</p>
      </div>
      <el-button type="primary" size="large" :icon="Plus" class="create-btn" @click="handleCreate">
        Tambah Umat Baru
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
        <!-- Filter Nama Ciu Tao -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Ciu Tao" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.namaindonesia" placeholder="Cari Nama Ciu Tao..." clearable :prefix-icon="User"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>

        <!-- Filter Nama Lain -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Lain" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.namamandarin" placeholder="Cari Nama Lain..." clearable :prefix-icon="Reading"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>

        <!-- Filter Alias / Pin Yin -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Alias / Pin Yin" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.alias" placeholder="Cari berdasarkan alias..." clearable :prefix-icon="Search"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>

      </el-row>
      <el-row :gutter="16">

        <!-- Filter Tahun Ciu Tao Mandarin -->
        <el-col :xs="24" :sm="12" :md="12">
          <el-form-item label="Tahun Internasional Chiu Tao" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.tahunchiutaomandarin" placeholder="Exact match tahun (e.g. 2024)..." clearable
              :prefix-icon="Calendar" @input="onFilterChange" />
          </el-form-item>
        </el-col>
      </el-row>

      <div class="filter-actions">
        <el-button :icon="Refresh" @click="resetFilters">Reset Filter</el-button>
      </div>
    </el-card>

    <!-- Table Card -->
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="umatList" stripe border height="475" style="width: 100%"
        empty-text="Tidak ada data umat yang ditemukan">
        <!-- <el-table-column prop="id" label="ID" width="80" align="center" /> -->
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="80" fixed="left" />

        <el-table-column prop="kode" label="Kode" width="150" :fixed="isMobile ? false : 'left'">
          <template #default="{ row }">
            <el-tag size="small" type="info" class="font-mono">{{ row.kode }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="nama_indonesia" label="Nama Ciu Tao" min-width="160">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.nama_indonesia }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="alias" label="Alias / Pin Yin" min-width="120">
          <template #default="{ row }">
            <span>{{ row.alias || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="nama_mandarin" label="Nama Lain" min-width="140">
          <template #default="{ row }">
            <span>{{ row.nama_mandarin || '-' }}</span>
          </template>
        </el-table-column>

        <!-- alamat -->
        <el-table-column prop="alamat" label="Alamat" min-width="250">
          <template #default="{ row }">
            <span>{{ row.alamat || '-' }}</span>
          </template>
        </el-table-column>

        <!-- pengajak -->
        <el-table-column prop="pengajak" label="Pengajak" min-width="150">
          <template #default="{ row }">
            <span>{{ row.pengajak_manual || '-' }}</span>
          </template>
        </el-table-column>

        <!-- penanggung -->
        <el-table-column prop="penanggung" label="Penanggung" min-width="150">
          <template #default="{ row }">
            <span>{{ row.penanggung_manual || '-' }}</span>
          </template>
        </el-table-column>

        <!-- usia -->
        <el-table-column prop="usia" label="Usia" min-width="150">
          <template #default="{ row }">
            <span>{{ row.usia || '-' }}</span>
          </template>
        </el-table-column>

        <!-- jenis kelamin -->
        <el-table-column prop="jenis_kelamin" label="Jenis Kelamin" min-width="150">
          <template #default="{ row }">
            <el-tag v-if="row.jenis_kelamin" size="small" type="warning" effect="plain">
              <span>{{ row.jenis_kelamin }}</span>
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>

        <!-- <el-table-column prop="tahun_chiutao_mandarin" label="Tahun Ciu Tao" min-width="140" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.tahun_chiutao_mandarin" size="small" type="warning" effect="plain">
              {{ row.tahun_chiutao_mandarin }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column> -->

        <el-table-column label="Kontak" min-width="150">
          <template #default="{ row }">
            <div class="contact-info">
              <span v-if="row.mobile">📱 {{ row.mobile }}</span>
              <span v-else-if="row.telepon">☎️ {{ row.telepon }}</span>
              <span v-else class="text-muted">-</span>
            </div>
          </template>
        </el-table-column>


        <!-- Actions Column -->
        <el-table-column label="Aksi" width="150" align="center" :fixed="!isDesktop ? false : 'right'">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button type="primary" size="small" circle :icon="Edit" title="Edit Umat"
                @click="handleEdit(row.id)" />

              <el-popconfirm title="Apakah Anda yakin ingin menghapus data umat ini?" confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal" confirm-button-type="danger" @confirm="handleDelete(row.id)">
                <template #reference>
                  <el-button type="danger" size="small" circle :icon="Delete" title="Hapus Umat" />
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-container">
        <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.limit"
          :page-sizes="[10, 20, 50, 100]" :total="pagination.total"
          :layout="'total, ' + (isDesktop ? ', jumper' : '') + ', prev, pager, next' + (isDesktop ? ', jumper' : '')"
          @size-change="handleSizeChange" @current-change="handlePageChange" />
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
  User,
  Reading,
  Calendar,
  Refresh,
  Plus,
  Edit,
  Delete
} from '@element-plus/icons-vue'
import { umatApi } from '../../api/umat'
import type { Umat, UmatQueryParams } from '../../types/umat'

const router = useRouter()

// Memory Optimization: Use shallowRef for data list to avoid deep reactivity overhead on large arrays
const umatList = shallowRef<Umat[]>([])
const loading = ref(false)

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

// Pagination State
const pagination = reactive({
  page: 1,
  limit: 10,
  total: 0
})

const getRowIndex = (index: number) => (pagination.page - 1) * pagination.limit + index + 1

// Search Filter State
const filters = reactive<UmatQueryParams>({
  alias: '',
  namaindonesia: '',
  namamandarin: '',
  tahunchiutaomandarin: ''
})

// Latency & Efficiency: AbortController to cancel stale in-flight requests
let currentAbortController: AbortController | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null

/**
 * Fetch Umats from API with current filters and pagination parameters.
 */
async function fetchUmats() {
  // Cancel previous running request if any
  if (currentAbortController) {
    currentAbortController.abort()
  }

  currentAbortController = new AbortController()
  loading.value = true

  try {
    const res = await umatApi.getUmats(
      {
        page: pagination.page,
        limit: pagination.limit,
        alias: filters.alias?.trim(),
        namaindonesia: filters.namaindonesia?.trim(),
        namamandarin: filters.namamandarin?.trim(),
        tahunchiutaomandarin: filters.tahunchiutaomandarin?.trim()
      },
      currentAbortController.signal
    )

    umatList.value = res.data || []
    pagination.total = res.meta?.total || 0
  } catch (err: any) {
    if (err.name === 'CanceledError' || err.name === 'AbortError') {
      // Ignored intentional aborts for latency optimization
      return
    }
    console.error('Error fetching umat list:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal memuat data umat')
  } finally {
    loading.value = false
  }
}

/**
 * Debounced search input handler to minimize latency and server load.
 */
function onFilterChange() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    pagination.page = 1
    fetchUmats()
  }, 300)
}

/**
 * Reset search filters back to empty defaults.
 */
function resetFilters() {
  filters.alias = ''
  filters.namaindonesia = ''
  filters.namamandarin = ''
  filters.tahunchiutaomandarin = ''
  pagination.page = 1
  fetchUmats()
}

/**
 * Handle page change event.
 */
function handlePageChange(newPage: number) {
  pagination.page = newPage
  fetchUmats()
}

/**
 * Handle page limit size change event.
 */
function handleSizeChange(newLimit: number) {
  pagination.limit = newLimit
  pagination.page = 1
  fetchUmats()
}

/**
 * Navigate to separate Create Umat page.
 */
function handleCreate() {
  router.push('/master-data/umat/create')
}

/**
 * Navigate to separate Edit Umat page.
 */
function handleEdit(id: number) {
  router.push(`/master-data/umat/edit/${id}`)
}

/**
 * Handle Umat deletion.
 */
async function handleDelete(id: number) {
  try {
    await umatApi.deleteUmat(id)
    ElNotification({
      title: 'Berhasil',
      message: 'Data umat telah dihapus',
      type: 'success'
    })
    fetchUmats()
  } catch (err: any) {
    console.error('Failed to delete umat:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal menghapus data umat')
  }
}

onMounted(() => {
  fetchUmats()
})

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
  if (currentAbortController) currentAbortController.abort()
})
</script>

<style scoped>
.umat-container {
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
  margin-top: 0.25rem;
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
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 1rem;
}

.filter-icon {
  color: var(--el-color-primary);
  font-size: 1.1rem;
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

.contact-info {
  font-size: 0.85rem;
  color: var(--el-text-color-regular);
}

.text-muted {
  color: var(--el-text-color-placeholder);
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  justify-content: center;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 1.25rem;
}
</style>
