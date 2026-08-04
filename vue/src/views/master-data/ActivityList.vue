<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Master Data Activity</h2>
        <p class="page-subtitle">Kelola daftar data activity, pencarian, serta pembaruan profil</p>
      </div>
      <el-button
        type="primary"
        size="large"
        :icon="Plus"
        class="create-btn"
        @click="handleCreate"
      >
        Tambah Activity Baru
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
          <el-form-item label="Kode Event" :label-position="isMobile ? 'top' : 'right'">
            <el-input
              v-model="filters.event_code"
              placeholder="Cari kode event..."
              clearable
              :prefix-icon="Search"
              @input="onFilterChange"
            />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="6">
          <el-form-item label="Nama Event" :label-position="isMobile ? 'top' : 'right'">
            <el-input
              v-model="filters.event_name"
              placeholder="Cari nama event..."
              clearable
              :prefix-icon="Search"
              @input="onFilterChange"
            />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="6">
          <el-form-item label="Kategori" :label-position="isMobile ? 'top' : 'right'">
            <LookupSelect
              v-model="filters.event_category"
              placeholder="Pilih Kategori Event..."
              :fetch-api="lookupApi.getLookupKategoriEvent"
              value-key="lookup_value"
              label-key="lookup_description"
              clearable
              @change="onFilterChange"
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
        empty-text="Tidak ada data activity yang ditemukan"
      >
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="70" align="center" fixed="left" />

        <el-table-column prop="event_code" label="Kode Event" width="130" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info" class="font-mono">{{ row.event_code }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="event_name" label="Nama Event" min-width="200">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.event_name || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="event_category" label="Kategori" min-width="150">
          <template #default="{ row }">
            <span>{{ row.event_category_info?.lookup_description || row.event_category || '-' }}</span>
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
              <el-button
                type="primary"
                size="small"
                circle
                :icon="Edit"
                title="Edit Activity"
                @click="handleEdit(row.event_code)"
              />

              <el-popconfirm
                title="Apakah Anda yakin ingin menghapus data ini?"
                confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal"
                confirm-button-type="danger"
                @confirm="handleDelete(row.event_code)"
              >
                <template #reference>
                  <el-button
                    type="danger"
                    size="small"
                    circle
                    :icon="Delete"
                    title="Hapus Activity"
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

    <!-- Create / Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? `Edit Activity #${editingCode}` : 'Tambah Activity Baru'"
      :width="isMobile ? '90%' : '560px'"
      destroy-on-close
      @closed="resetForm"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="130px"
        :label-position="isMobile ? 'top' : 'right'"
      >
        <el-form-item label="Kode Event" prop="event_code">
          <el-input
            v-model="formData.event_code"
            placeholder="Masukkan kode event (misal: EV01)"
            :disabled="isEditing"
          />
        </el-form-item>

        <el-form-item label="Nama Event" prop="event_name">
          <el-input
            v-model="formData.event_name"
            placeholder="Masukkan nama event"
          />
        </el-form-item>

        <el-form-item label="Kategori Event" prop="event_category">
          <LookupSelect
            v-model="formData.event_category"
            placeholder="Pilih Kategori Event..."
            :fetch-api="lookupApi.getLookupKategoriEvent"
            value-key="lookup_value"
            label-key="lookup_description"
            clearable
          />
        </el-form-item>

        <el-form-item label="Keterangan" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="Masukkan keterangan tambahan..."
          />
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
import { activityApi } from '../../api/activity'
import { lookupApi } from '../../api/lookup'
import type { Activity, ActivityQueryParams } from '../../types/activity'
import LookupSelect from '../../components/common/LookupSelect.vue'

// Breakpoints layout calculation
const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')
const isDesktop = breakpoints.greaterOrEqual('lg')

// Dataset state
const dataList = shallowRef<Activity[]>([])
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
const filters = reactive<ActivityQueryParams>({
  event_code: '',
  event_name: '',
  event_category: ''
})

// Dialog & Form state
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingCode = ref('')
const submitting = ref(false)
const formRef = ref<FormInstance | null>(null)

const formData = reactive<Partial<Activity>>({
  event_code: '',
  event_name: '',
  event_category: '',
  description: '',
  // status: true
})

const formRules: FormRules = {
  event_code: [
    { required: true, message: 'Kode event wajib diisi', trigger: 'blur' },
    { max: 10, message: 'Kode event maksimal 10 karakter', trigger: 'blur' }
  ],
  event_name: [
    { required: true, message: 'Nama event wajib diisi', trigger: 'blur' },
    { max: 100, message: 'Nama event maksimal 100 karakter', trigger: 'blur' }
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
    const res = await activityApi.getActivitys(
      {
        page: pagination.page,
        limit: pagination.limit,
        event_code: filters.event_code?.trim(),
        event_name: filters.event_name?.trim(),
        event_category: filters.event_category?.trim()
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
  filters.event_code = ''
  filters.event_name = ''
  filters.event_category = ''
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
  formData.event_code = ''
  formData.event_name = ''
  formData.event_category = ''
  formData.description = ''
  // formData.status = true
  isEditing.value = false
  editingCode.value = ''
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

function handleCreate() {
  resetForm()
  isEditing.value = false
  dialogVisible.value = true
}

async function handleEdit(code: string) {
  resetForm()
  isEditing.value = true
  editingCode.value = code

  try {
    const res = await activityApi.getActivityById(code)
    const activity = res.data
    if (activity) {
      formData.event_code = activity.event_code
      formData.event_name = activity.event_name
      formData.event_category = activity.event_category || ''
      formData.description = activity.description || ''
      // formData.status = activity.status !== undefined ? activity.status : true
      dialogVisible.value = true
    }
  } catch (err: unknown) {
    console.error('Error fetching activity by code:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal memuat detail activity')
  }
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitting.value = true
  try {
    const payload: Partial<Activity> = {
      event_code: formData.event_code?.trim(),
      event_name: formData.event_name?.trim(),
      event_category: formData.event_category || undefined,
      description: formData.description?.trim() || undefined,
      // status: formData.status
    }

    if (isEditing.value && editingCode.value) {
      await activityApi.updateActivity(editingCode.value, payload)
      ElNotification({
        title: 'Berhasil',
        message: `Activity ${editingCode.value} berhasil diperbarui`,
        type: 'success'
      })
    } else {
      await activityApi.createActivity(payload)
      ElNotification({
        title: 'Berhasil',
        message: 'Activity baru berhasil ditambahkan',
        type: 'success'
      })
    }

    dialogVisible.value = false
    fetchData()
  } catch (err: unknown) {
    console.error('Failed to submit activity form:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal menyimpan data activity')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(code: string) {
  try {
    await activityApi.deleteActivity(code)
    ElNotification({
      title: 'Berhasil',
      message: `Data activity ${code} berhasil dihapus`,
      type: 'success'
    })
    fetchData()
  } catch (err: unknown) {
    console.error('Failed to delete item:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal menghapus data activity')
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
