<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Group Menu Mapping</h2>
        <p class="page-subtitle">Kelola daftar data group menu mapping, pencarian, serta pembaruan profil</p>
      </div>
      <!-- <el-button
        type="primary"
        size="large"
        :icon="Plus"
        class="create-btn"
        @click="handleCreate"
      >
        Tambah Group Menu Mapping Baru
      </el-button> -->
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
          <el-form-item label="Nama Menu" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.menu_name" placeholder="Cari nama menu..." clearable :prefix-icon="Search"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="12">
          <el-form-item label="URL Halaman" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.page_url" placeholder="Cari page url..." clearable :prefix-icon="Search"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>
      </el-row>

      <div class="filter-actions">
        <el-button :icon="Refresh" @click="resetFilters">Reset Filter</el-button>
      </div>
    </el-card>

    <!-- Table Card -->
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="dataList" stripe border height="475" style="width: 100%"
        empty-text="Tidak ada data group menu mapping yang ditemukan">
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="70" align="center"
          fixed="left" />

        <el-table-column prop="menu_name" label="Nama Menu" min-width="180">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.menu_name || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="page_url" label="URL Halaman" min-width="200">
          <template #default="{ row }">
            <span>{{ row.page_url || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="sequence" label="Urutan" width="90" align="center">
          <template #default="{ row }">
            <span>{{ row.sequence || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="menu_desc" label="Keterangan" min-width="200">
          <template #default="{ row }">
            <span>{{ row.menu_desc || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="flag_active" label="Status" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="row.flag_active ? 'success' : 'info'" size="small">
              {{ row.flag_active ? 'Aktif' : 'Nonaktif' }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- Actions Column -->
        <el-table-column label="Aksi" width="150" align="center" :fixed="!isDesktop ? false : 'right'">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button type="primary" size="small" circle :icon="Edit" title="Edit Group Menu Mapping"
                @click="handleEdit(row.menu_id)" />

              <el-popconfirm title="Apakah Anda yakin ingin menghapus data ini?" confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal" confirm-button-type="danger" @confirm="handleDelete(row.menu_id)">
                <template #reference>
                  <el-button type="danger" size="small" circle :icon="Delete" title="Hapus Group Menu Mapping" />
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
          :layout="(!isMobile ? 'total, ->,' : (Math.ceil(pagination.total / pagination.limit) < 6 ? '-> ,' : '')) + 'prev, pager, next' + (isDesktop ? ', jumper' : '')"
          :pager-count="6" @size-change="handleSizeChange" @current-change="handlePageChange" />
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
import { groupMenuMappingApi } from '../../api/groupMenuMapping'
import type { GroupMenuMapping, GroupMenuMappingQueryParams } from '../../types/groupMenuMapping'

const router = useRouter()

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

// Memory Optimization: shallowRef for table dataset
const dataList = shallowRef<GroupMenuMapping[]>([])
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
const filters = reactive<GroupMenuMappingQueryParams>({
  menu_name: '',
  page_url: ''
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
    const res = await groupMenuMappingApi.getGroupMenuMappings(
      {
        page: pagination.page,
        limit: pagination.limit,
        menu_name: filters.menu_name?.trim(),
        page_url: filters.page_url?.trim()
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
  filters.menu_name = ''
  filters.page_url = ''
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
    message: 'Tambah group menu mapping baru',
    type: 'info'
  })
}

function handleEdit(id: number) {
  ElNotification({
    title: 'Informasi',
    message: `Edit group menu mapping ID/Kode: ${id}`,
    type: 'info'
  })
}

async function handleDelete(id: number) {
  try {
    await groupMenuMappingApi.deleteGroupMenuMapping(id)
    ElNotification({
      title: 'Berhasil',
      message: `Data group menu mapping ${id} berhasil dihapus`,
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
  margin-top: 1.25rem;
}
</style>
