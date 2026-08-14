<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Master Data Tahun Ciu Tao</h2>
        <p class="page-subtitle">Kelola daftar data tahun ciu tao, pencarian, serta pembaruan profil</p>
      </div>
      <el-button type="primary" size="large" :icon="Plus" class="create-btn" @click="handleCreate">
        Tambah Tahun Ciu Tao Baru
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
        <el-col :xs="24" :sm="12" :md="10">
          <el-form-item label="Tahun Mandarin" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.tahun_mandarin" placeholder="Cari tahun (e.g. 2024)..." clearable
              :prefix-icon="Search" @input="onFilterChange" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="10">
          <el-form-item label="Keterangan" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.description" placeholder="Cari deskripsi..." clearable :prefix-icon="Search"
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
        empty-text="Tidak ada data tahun ciu tao yang ditemukan">
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="70" align="center"
          fixed="left" />

        <el-table-column prop="tahun_mandarin" label="Tahun Mandarin" min-width="160">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.tahun_mandarin || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="start_date" label="Tanggal Mulai" min-width="140" align="center">
          <template #default="{ row }">
            <span>{{ row.start_date || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="end_date" label="Tanggal Selesai" min-width="140" align="center">
          <template #default="{ row }">
            <span>{{ row.end_date || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="description" label="Keterangan" min-width="220">
          <template #default="{ row }">
            <span>{{ row.description || '-' }}</span>
          </template>
        </el-table-column>

        <!-- <el-table-column prop="status" label="Status" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status ? 'success' : 'info'" size="small">
              {{ row.status ? 'Aktif' : 'Nonaktif' }}
            </el-tag>
          </template>
        </el-table-column> -->

        <!-- Actions Column -->
        <el-table-column label="Aksi" width="150" align="center" :fixed="!isDesktop ? false : 'right'">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button type="primary" size="small" circle :icon="Edit" title="Edit Tahun Ciu Tao"
                @click="handleEdit(row.tahun_mandarin)" />

              <el-popconfirm title="Apakah Anda yakin ingin menghapus data ini?" confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal" confirm-button-type="danger" @confirm="handleDelete(row.tahun_mandarin)">
                <template #reference>
                  <el-button type="danger" size="small" circle :icon="Delete" title="Hapus Tahun Ciu Tao" />
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

    <!-- Create / Edit Dialog -->
    <el-dialog v-model="dialogVisible"
      :title="isEditing ? `Edit Tahun Ciu Tao #${editingId}` : 'Tambah Tahun Ciu Tao Baru'"
      :width="isMobile ? '90%' : '560px'" destroy-on-close @closed="resetForm">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="140px"
        :label-position="isMobile ? 'top' : 'right'">
        <el-form-item label="Tahun Mandarin" prop="tahun_mandarin">
          <el-input v-model="formData.tahun_mandarin" placeholder="Masukkan tahun (e.g. 2024)" :disabled="isEditing" />
        </el-form-item>

        <el-form-item label="Periode Tanggal">
          <el-date-picker v-model="dateRange" type="daterange" range-separator="s/d" start-placeholder="Tanggal Mulai"
            end-placeholder="Tanggal Selesai" value-format="YYYY-MM-DD" format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>

        <el-form-item label="Keterangan" prop="description">
          <el-input v-model="formData.description" type="textarea" :rows="3" placeholder="Masukkan keterangan..." />
        </el-form-item>

        <!-- <el-form-item label="Status" prop="status">
          <el-switch
            v-model="formData.status"
            active-text="Aktif"
            inactive-text="Nonaktif"
          />
        </el-form-item> -->
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button :disabled="submitting" @click="dialogVisible = false">Batal</el-button>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">
            Simpan
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import { ElMessage, ElNotification, type FormInstance, type FormRules } from 'element-plus'
import {
  Search,
  Refresh,
  Plus,
  Edit,
  Delete
} from '@element-plus/icons-vue'
import { tahunCiuTaoApi } from '../../api/tahunCiuTao'
import type { TahunCiuTao, TahunCiuTaoQueryParams } from '../../types/tahunCiuTao'

// Breakpoints layout calculation
const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')
const isDesktop = breakpoints.greaterOrEqual('lg')

// Dataset state
const dataList = shallowRef<TahunCiuTao[]>([])
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
const filters = reactive<TahunCiuTaoQueryParams>({
  tahun_mandarin: '',
  description: ''
})

// Dialog & Form state
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref('')
const submitting = ref(false)
const formRef = ref<FormInstance | null>(null)

const formData = reactive<Partial<TahunCiuTao>>({
  tahun_mandarin: '',
  start_date: '',
  end_date: '',
  description: '',
  // status: true
})

const dateRange = computed({
  get(): [string, string] | [] {
    if (formData.start_date || formData.end_date) {
      return [formData.start_date || '', formData.end_date || '']
    }
    return []
  },
  set(val: [string, string] | null) {
    if (val && val.length === 2) {
      formData.start_date = val[0]
      formData.end_date = val[1]
    } else {
      formData.start_date = ''
      formData.end_date = ''
    }
  }
})

const formRules: FormRules = {
  tahun_mandarin: [
    { required: true, message: 'Tahun mandarin wajib diisi', trigger: 'blur' },
    { max: 20, message: 'Tahun mandarin maksimal 20 karakter', trigger: 'blur' }
  ]
}

let currentAbortController: AbortController | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null

async function fetchData() {
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()
  loading.value = true

  try {
    const res = await tahunCiuTaoApi.getTahunCiuTaos(
      {
        page: pagination.page,
        limit: pagination.limit,
        tahun_mandarin: filters.tahun_mandarin?.trim(),
        description: filters.description?.trim()
      },
      currentAbortController.signal
    )

    dataList.value = res.data || []
    pagination.total = res.meta?.total || 0
  } catch (err: unknown) {
    if (err instanceof Error && (err.name === 'CanceledError' || err.name === 'AbortError')) return
    console.error('Error fetching data:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal memuat data')
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
  filters.tahun_mandarin = ''
  filters.description = ''
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

function resetForm() {
  formData.tahun_mandarin = ''
  formData.start_date = ''
  formData.end_date = ''
  formData.description = ''
  // formData.status = true
  isEditing.value = false
  editingId.value = ''
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

function handleCreate() {
  resetForm()
  isEditing.value = false
  dialogVisible.value = true
}

async function handleEdit(id: string) {
  resetForm()
  isEditing.value = true
  editingId.value = id

  try {
    const res = await tahunCiuTaoApi.getTahunCiuTaoById(id)
    const record = res.data
    if (record) {
      formData.tahun_mandarin = record.tahun_mandarin
      formData.start_date = record.start_date || ''
      formData.end_date = record.end_date || ''
      formData.description = record.description || ''
      // formData.status = record.status !== undefined ? record.status : true
      dialogVisible.value = true
    }
  } catch (err: unknown) {
    console.error('Error fetching record by ID:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal memuat detail data')
  }
}

import { scrollToFormError } from '../../utils/scroll'

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    scrollToFormError()
    return
  }

  submitting.value = true
  try {
    const payload: Partial<TahunCiuTao> = {
      tahun_mandarin: formData.tahun_mandarin?.trim(),
      start_date: formData.start_date || undefined,
      end_date: formData.end_date || undefined,
      description: formData.description?.trim() || undefined,
      // status: formData.status
    }

    if (isEditing.value && editingId.value) {
      await tahunCiuTaoApi.updateTahunCiuTao(editingId.value, payload)
      ElNotification({
        title: 'Berhasil',
        message: `Tahun Ciu Tao ${editingId.value} berhasil diperbarui`,
        type: 'success'
      })
    } else {
      await tahunCiuTaoApi.createTahunCiuTao(payload)
      ElNotification({
        title: 'Berhasil',
        message: 'Tahun Ciu Tao baru berhasil ditambahkan',
        type: 'success'
      })
    }

    dialogVisible.value = false
    fetchData()
  } catch (err: unknown) {
    console.error('Failed to submit form:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal menyimpan data')
    scrollToFormError()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: string) {
  try {
    await tahunCiuTaoApi.deleteTahunCiuTao(id)
    ElNotification({
      title: 'Berhasil',
      message: `Data Tahun Ciu Tao ${id} berhasil dihapus`,
      type: 'success'
    })
    fetchData()
  } catch (err: unknown) {
    console.error('Failed to delete item:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal menghapus data')
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

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
