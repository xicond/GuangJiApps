<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Master Data Tahun Ciu Tao</h2>
        <p class="page-subtitle">Kelola daftar data tahun ciu tao, pencarian, serta manajemen data</p>
      </div>
      <el-button
        type="primary"
        size="large"
        :icon="Plus"
        class="create-btn"
        @click="handleCreate"
      >
        Tambah Tahun Ciu Tao Baru
      </el-button>
    </div>

    <!-- Filter Card -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-header">
        <el-icon class="filter-icon"><Search /></el-icon>
        <span class="filter-title">Filter & Pencarian Data</span>
      </div>

      <el-row :gutter="16" class="filter-row">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Kata Kunci">
            <el-input
              v-model="filters.keyword"
              placeholder="Cari berdasarkan nama atau kata kunci..."
              clearable
              :prefix-icon="Search"
              @input="onFilterChange"
            />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Status">
            <el-select
              v-model="filters.status"
              placeholder="Semua Status"
              clearable
              @change="onFilterChange"
            >
              <el-option label="Aktif" value="active" />
              <el-option label="Nonaktif" value="inactive" />
            </el-select>
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
        empty-text="Tidak ada data tahun ciu tao yang ditemukan"
      >
        <el-table-column prop="id" label="ID" width="80" align="center" sortable />

        <el-table-column prop="nama" label="Nama Tahun Ciu Tao" min-width="180">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.nama || row.name || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="description" label="Keterangan" min-width="200">
          <template #default="{ row }">
            <span>{{ row.description || row.keterangan || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="status" label="Status" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? 'Aktif' : 'Nonaktif' }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- Actions Column -->
        <el-table-column label="Aksi" width="150" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                type="primary"
                size="small"
                circle
                :icon="Edit"
                title="Edit Tahun Ciu Tao"
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
                    title="Hapus Tahun Ciu Tao"
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
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import {
  Search,
  Refresh,
  Plus,
  Edit,
  Delete
} from '@element-plus/icons-vue'

const router = useRouter()

const dataList = shallowRef<any[]>([])
const loading = ref(false)

const pagination = reactive({
  page: 1,
  limit: 10,
  total: 0
})

const filters = reactive({
  keyword: '',
  status: ''
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
    dataList.value = []
    pagination.total = 0
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
  filters.keyword = ''
  filters.status = ''
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

function handleCreate() {
  ElNotification({
    title: 'Informasi',
    message: 'Tambah tahun ciu tao baru',
    type: 'info'
  })
}

function handleEdit(id: number | string) {
  ElNotification({
    title: 'Informasi',
    message: `Edit tahun ciu tao ID: ${id}`,
    type: 'info'
  })
}

async function handleDelete(id: number | string) {
  try {
    ElNotification({
      title: 'Berhasil',
      message: `Data tahun ciu tao ID: ${id} berhasil dihapus`,
      type: 'success'
    })
    fetchData()
  } catch (err: any) {
    ElMessage.error('Gagal menghapus data')
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

.font-semibold {
  font-weight: 600;
}
</style>
