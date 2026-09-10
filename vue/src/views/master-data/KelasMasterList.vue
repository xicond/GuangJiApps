<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Master Data Kelas</h2>
        <p class="page-subtitle">Kelola data master kelas khusus</p>
      </div>
      <el-button type="primary" size="large" :icon="Plus" class="create-btn" @click="handleCreate">
        Tambah Kelas Baru
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
        <!-- <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Kode Kelas" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.lookup_value" placeholder="Cari kode kelas (contoh: 001)..." clearable
              :prefix-icon="Search" @input="onFilterChange" />
          </el-form-item>
        </el-col> -->

        <el-col :xs="24" :sm="12" :md="10">
          <el-form-item label="Nama Kelas" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.lookup_description" placeholder="Cari nama kelas..." clearable
              :prefix-icon="Search" @input="onFilterChange" />
          </el-form-item>
        </el-col>

        <!-- <el-col :xs="24" :sm="12" :md="6">
          <el-form-item label="Status" :label-position="isMobile ? 'top' : 'right'">
            <el-select v-model="filters.status" placeholder="Semua Status" clearable @change="onFilterChange"
              style="width: 100%">
              <el-option label="Aktif" value="1" />
              <el-option label="Nonaktif" value="0" />
            </el-select>
          </el-form-item>
        </el-col> -->
      </el-row>

      <div class="filter-actions">
        <el-button :icon="Refresh" @click="resetFilters">Reset Filter</el-button>
      </div>
    </el-card>

    <!-- Table Card -->
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="dataList" stripe border height="475" style="width: 100%"
        empty-text="Tidak ada data kelas yang ditemukan">
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="70" align="center"
          fixed="left" />

        <el-table-column prop="lookup_value" label="Kode Kelas" width="130" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="primary" class="font-mono">{{ row.lookup_value || '-' }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="lookup_id" label="Lookup ID" width="190">
          <template #default="{ row }">
            <span class="font-mono text-muted">{{ row.lookup_id || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="lookup_description" label="Nama Kelas" min-width="280">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.lookup_description || '-' }}</span>
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
        <el-table-column label="Aksi" width="140" align="center" :fixed="!isDesktop ? false : 'right'">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button type="primary" size="small" circle :icon="Edit" title="Edit Kelas"
                @click="handleEdit(row.lookup_id)" />

              <!-- <el-popconfirm title="Apakah Anda yakin ingin menonaktifkan data ini?"
                confirm-button-text="Ya, Nonaktifkan" cancel-button-text="Batal" confirm-button-type="danger"
                @confirm="handleDelete(row.lookup_id)">
                <template #reference>
                  <el-button type="danger" size="small" circle :icon="Delete" title="Hapus Kelas" />
                </template>
  </el-popconfirm> -->
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

    <!-- Create / Edit Dialog -->
    <el-dialog v-model="dialogVisible" :title="isEditing ? `Edit Kelas #${editingId}` : 'Tambah Kelas Baru'"
      :width="isMobile ? '90%' : '560px'" destroy-on-close @closed="resetForm">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="140px"
        :label-position="isMobile ? 'top' : 'right'" :disabled="submitting">
        <el-form-item label="Kode Kelas" prop="lookup_value">
          <el-input v-model="formData.lookup_value" placeholder="Otomatis dibuat oleh sistem jika dikosongkan"
            :disabled="isEditing" clearable />
          <div v-if="!isEditing" class="form-hint">
            *Biarkan kosong untuk penomoran otomatis (contoh: 023)
          </div>
          <FieldErrors :errors="fieldErrors.lookup_value" />
        </el-form-item>

        <el-form-item label="Nama Kelas" prop="lookup_description">
          <el-autocomplete v-if="!isEditing" v-model="formData.lookup_description"
            :fetch-suggestions="queryKelasSuggestions" :trigger-on-focus="false" placeholder="Ketik nama kelas"
            clearable style="width: 100%" @select="handleSelectSuggestion">
            <template #default="{ item }">
              <div class="kelas-suggest-item">
                <div class="kelas-suggest-row">
                  <span class="kelas-suggest-name font-semibold">{{ item.lookup_description }}</span>
                  <el-tag size="small" type="primary" class="font-mono">{{ item.lookup_value }}</el-tag>
                </div>
                <div class="kelas-suggest-desc">
                  <span class="font-mono text-muted">{{ item.lookup_id }}</span>
                </div>
              </div>
            </template>
          </el-autocomplete>
          <el-input v-else v-model="formData.lookup_description" placeholder="Masukkan nama kelas" />

          <div v-if="fieldErrors.lookup_description?.length" class="field-errors-list">
            <div v-for="(errMsg, idx) in fieldErrors.lookup_description" :key="idx" class="field-error-item">
              <span class="error-bullet">&bull;</span>
              <span>
                {{ errMsg }}<template v-if="duplicateLookupId && errMsg.includes('sudah ada')">,
                  <el-link type="primary" :underline="true" class="duplicate-edit-link"
                    @click="handleEdit(duplicateLookupId)">
                    update here instead
                  </el-link>
                </template>
              </span>
            </div>
          </div>
        </el-form-item>

        <el-form-item v-if="isEditing" label="Status" prop="status">
          <el-switch v-model="formData.status" active-text="Aktif" inactive-text="Nonaktif" />
        </el-form-item>
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
import { ref, shallowRef, reactive, onMounted, onUnmounted } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import { ElMessage, ElNotification, type FormInstance, type FormRules } from 'element-plus'
import {
  Search,
  Refresh,
  Plus,
  Edit,
  Delete
} from '@element-plus/icons-vue'
import { kelasMasterApi } from '../../api/kelasMaster'
import type { KelasMaster, KelasMasterQueryParams } from '../../types/kelasMaster'
import FieldErrors from '../../components/common/FieldErrors.vue'
import { scrollToFormError } from '../../utils/scroll'

// Responsive breakpoints
const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')
const isDesktop = breakpoints.greaterOrEqual('lg')

// Table dataset state
const dataList = shallowRef<KelasMaster[]>([])
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
const filters = reactive<KelasMasterQueryParams>({
  lookup_value: '',
  lookup_description: '',
  status: ''
})

// Dialog & Form state
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref('')
const submitting = ref(false)
const formRef = ref<FormInstance | null>(null)
const fieldErrors = ref<Record<string, string[]>>({})
const duplicateLookupId = ref('')

interface KelasSuggestion {
  value: string
  lookup_id: string
  lookup_value?: string
  lookup_description: string
}

const formData = reactive<Partial<KelasMaster>>({
  lookup_id: '',
  lookup_value: '',
  lookup_description: '',
  status: true
})

const formRules: FormRules = {
  lookup_description: [
    { required: true, message: 'Nama kelas wajib diisi', trigger: 'blur' },
    { max: 150, message: 'Nama kelas maksimal 150 karakter', trigger: 'blur' }
  ]
}

let currentAbortController: AbortController | null = null
let suggestAbortController: AbortController | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const queryKelasSuggestions = async (
  queryString: string,
  cb: (suggestions: KelasSuggestion[]) => void
) => {
  const trimmed = queryString ? queryString.trim() : ''
  const words = trimmed.split(/\s+/).filter(Boolean)

  if (words.length < 2) {
    cb([])
    return
  }

  if (suggestAbortController) {
    suggestAbortController.abort()
  }
  suggestAbortController = new AbortController()

  try {
    const res = await kelasMasterApi.getKelasMasters(
      {
        lookup_description: trimmed,
        limit: 10
      },
      suggestAbortController.signal
    )
    const items: KelasSuggestion[] = (res.data || []).map((k) => ({
      value: k.lookup_description || '',
      lookup_id: k.lookup_id,
      lookup_value: k.lookup_value,
      lookup_description: k.lookup_description || ''
    }))
    cb(items)
  } catch (err: unknown) {
    if (err instanceof Error && (err.name === 'CanceledError' || err.name === 'AbortError')) return
    cb([])
  }
}

function handleSelectSuggestion(item: KelasSuggestion) {
  if (item?.lookup_id) {
    handleEdit(item.lookup_id)
  }
}

async function fetchData() {
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()
  loading.value = true

  try {
    const res = await kelasMasterApi.getKelasMasters(
      {
        page: pagination.page,
        limit: pagination.limit,
        lookup_value: filters.lookup_value?.trim(),
        lookup_description: filters.lookup_description?.trim(),
        status: filters.status
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
  filters.lookup_value = ''
  filters.lookup_description = ''
  filters.status = ''
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
  formData.lookup_id = ''
  formData.lookup_value = ''
  formData.lookup_description = ''
  formData.status = true
  isEditing.value = false
  editingId.value = ''
  fieldErrors.value = {}
  duplicateLookupId.value = ''
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
    const res = await kelasMasterApi.getKelasMasterById(id)
    const item = res.data
    if (item) {
      formData.lookup_id = item.lookup_id
      formData.lookup_value = item.lookup_value || ''
      formData.lookup_description = item.lookup_description || ''
      formData.status = item.status !== undefined ? item.status : true
      dialogVisible.value = true
    }
  } catch (err: unknown) {
    console.error('Error fetching kelas master by ID:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal memuat detail kelas')
  }
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    scrollToFormError()
    return
  }

  fieldErrors.value = {}
  duplicateLookupId.value = ''
  submitting.value = true

  try {
    const payload: Partial<KelasMaster> = {
      lookup_value: formData.lookup_value?.trim() || undefined,
      lookup_description: formData.lookup_description?.trim() || undefined,
      status: formData.status
    }

    if (isEditing.value && editingId.value) {
      await kelasMasterApi.updateKelasMaster(editingId.value, payload)
      ElNotification({
        title: 'Berhasil',
        message: `Kelas ${editingId.value} berhasil diperbarui`,
        type: 'success'
      })
    } else {
      await kelasMasterApi.createKelasMaster(payload)
      ElNotification({
        title: 'Berhasil',
        message: 'Kelas baru berhasil ditambahkan',
        type: 'success'
      })
    }

    dialogVisible.value = false
    fetchData()
  } catch (err: unknown) {
    console.error('Failed to submit kelas master form:', err)
    const errObj = err as {
      response?: {
        data?: {
          error?: string
          details?: Record<string, string[]>
        }
      }
      message?: string
    }

    if (errObj.response?.data?.details) {
      fieldErrors.value = errObj.response.data.details

      const descErrors = fieldErrors.value.lookup_description || []
      const isDuplicateError = descErrors.some((msg) => msg.includes('sudah ada'))

      if (isDuplicateError && formData.lookup_description) {
        try {
          const searchRes = await kelasMasterApi.getKelasMasters({
            lookup_description: formData.lookup_description.trim(),
            limit: 5
          })
          const matched =
            searchRes.data?.find(
              (k) => k.lookup_description?.trim().toLowerCase() === formData.lookup_description!.trim().toLowerCase()
            ) || searchRes.data?.[0]

          if (matched) {
            duplicateLookupId.value = matched.lookup_id
          }
        } catch (searchErr) {
          console.warn('Failed to resolve duplicate kelas id:', searchErr)
        }
      }

      ElMessage.error(errObj.response.data.error || 'Terdapat kesalahan validasi')
      scrollToFormError()
    } else {
      ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal menyimpan data')
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: string) {
  try {
    await kelasMasterApi.deleteKelasMaster(id)
    ElNotification({
      title: 'Berhasil',
      message: `Kelas ${id} berhasil dinonaktifkan`,
      type: 'success'
    })
    fetchData()
  } catch (err: unknown) {
    console.error('Error deleting kelas master:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal menghapus data')
  }
}

onMounted(() => {
  fetchData()
})

onUnmounted(() => {
  if (currentAbortController) currentAbortController.abort()
  if (suggestAbortController) suggestAbortController.abort()
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

.text-muted {
  color: var(--el-text-color-secondary);
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

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.form-hint {
  font-size: 0.75rem;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.kelas-suggest-item {
  display: flex;
  flex-direction: column;
  padding: 4px 0;
  line-height: 1.3;
}

.kelas-suggest-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.kelas-suggest-name {
  color: var(--el-text-color-primary);
  font-size: 0.875rem;
}

.kelas-suggest-desc {
  font-size: 0.75rem;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.field-errors-list {
  margin-top: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 100%;
}

.field-error-item {
  color: var(--el-color-danger, #f56c6c);
  font-size: 0.75rem;
  line-height: 1.25;
  display: flex;
  align-items: flex-start;
  gap: 4px;
}

.error-bullet {
  font-weight: bold;
}

.duplicate-edit-link {
  font-size: 0.75rem;
  font-weight: 600;
  vertical-align: baseline;
  text-decoration: underline;
  margin-left: 4px;
}
</style>
