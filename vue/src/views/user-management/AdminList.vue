<template>
  <div class="admin-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Master Data Admin</h2>
        <p class="page-subtitle">
          Kelola daftar akun administrator, hak akses group, status penggunaan, serta manajemen pengguna
        </p>
      </div>
      <el-button type="primary" size="large" :icon="Plus" class="create-btn" @click="openCreateDialog">
        Tambah Admin Baru
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
        <!-- Filter Username -->
        <el-col :xs="24" :sm="12" :md="10">
          <el-form-item label="Username" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.username" placeholder="Cari berdasarkan username..." clearable
              :prefix-icon="Search" @input="onFilterChange" />
          </el-form-item>
        </el-col>

        <!-- Filter Group Name -->
        <el-col :xs="24" :sm="12" :md="10">
          <el-form-item label="Nama Group" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.group_name" placeholder="Cari berdasarkan nama group..." clearable
              :prefix-icon="UserFilled" @input="onFilterChange" />
          </el-form-item>
        </el-col>
      </el-row>

      <div class="filter-actions">
        <el-button :icon="Refresh" @click="resetFilters">Reset Filter</el-button>
      </div>
    </el-card>

    <!-- Table Card -->
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="adminList" stripe border height="475" style="width: 100%"
        empty-text="Tidak ada data admin yang ditemukan">
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="80" fixed="left" />
        <!-- <el-table-column prop="id" label="ID" width="80" align="center" fixed="left" /> -->

        <el-table-column prop="username" label="User Name" min-width="135" :fixed="isMobile ? false : 'left'"
          align="left">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.username }}</span>
          </template>
        </el-table-column>

        <el-table-column label="Group" min-width="160">
          <template #default="{ row }">
            <el-tag v-if="row.admin_group?.group_name" type="primary" effect="plain" size="small">
              {{ row.admin_group.group_name }}
            </el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>

        <el-table-column label="Dept Name" min-width="160">
          <template #default="{ row }">
            <el-tag v-if="row.department?.department_name" type="primary" effect="plain" size="small">
              {{ row.department.department_code + " - " + row.department.department_name }}
            </el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>

        <!-- <el-table-column prop="email" label="Email" min-width="180">
          <template #default="{ row }">
            <span>{{ row.email || '-' }}</span>
          </template>
        </el-table-column> -->

        <!-- <el-table-column prop="phone_number" label="Telepon" min-width="140">
          <template #default="{ row }">
            <span>{{ row.phone_number || '-' }}</span>
          </template>
        </el-table-column> -->

        <!-- <el-table-column prop="flag_use" label="Status" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="row.flag_use ? 'success' : 'danger'" size="small">
              {{ row.flag_use ? 'Aktif' : 'Nonaktif' }}
            </el-tag>
          </template>
        </el-table-column> -->

        <el-table-column prop="is_warehouse" label="Warehouse" width="120" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_warehouse" type="warning" size="small">Ya</el-tag>
            <span v-else class="text-muted">Tidak</span>
          </template>
        </el-table-column>

        <el-table-column prop="date_end" label="Expired Date" width="120" align="center">
          <template #default="{ row }">
            <span>{{ formatDate(row.date_end) }}</span>
          </template>
        </el-table-column>

        <!-- <el-table-column prop="login_desc" label="Keterangan" min-width="180">
          <template #default="{ row }">
            <span>{{ row.login_desc || '-' }}</span>
          </template>
        </el-table-column> -->

        <el-table-column prop="last_login" label="Login Terakhir" min-width="160" align="center">
          <template #default="{ row }">
            <span>{{ formatDate(row.last_login) }}</span>
          </template>
        </el-table-column>

        <!-- Actions Column -->
        <el-table-column label="Aksi" width="130" align="center" :fixed="!isDesktop ? false : 'right'">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button type="primary" size="small" circle :icon="Edit" title="Edit Admin"
                @click="openEditDialog(row)" />

              <el-popconfirm title="Apakah Anda yakin ingin menghapus admin ini?" confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal" confirm-button-type="danger" @confirm="handleDelete(row.id)">
                <template #reference>
                  <el-button type="danger" size="small" circle :icon="Delete" title="Hapus Admin" />
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
    <el-dialog v-model="dialogVisible" :title="isEditing ? `Edit Admin #${editingId}` : 'Tambah Admin Baru'"
      :width="isMobile ? '90%' : '560px'" destroy-on-close @closed="resetForm">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="130px"
        :label-position="isMobile ? 'top' : 'right'">
        <el-form-item label="Username" prop="username">
          <el-input v-model="formData.username" placeholder="Masukkan username" />
        </el-form-item>

        <el-form-item label="Password" prop="password"
          :rules="isEditing ? [] : [{ required: true, message: 'Password wajib diisi', trigger: 'blur' }]">
          <el-input v-model="formData.password" type="password" show-password
            :placeholder="isEditing ? 'Kosongkan jika tidak ingin mengubah password' : 'Masukkan password'" />
        </el-form-item>

        <el-form-item label="Group Admin" prop="group_id">
          <el-select v-model="formData.group_id" placeholder="Pilih Group Admin" style="width: 100%" filterable>
            <el-option v-for="group in groupOptions" :key="group.group_id" :label="group.group_name"
              :value="group.group_id" />
          </el-select>
        </el-form-item>

        <el-form-item label="Department" prop="department_id">
          <el-select v-model="formData.department_id" placeholder="Pilih Department" style="width: 100%" filterable
            clearable>
            <el-option v-for="department in departmentOptions" :key="department.department_id"
              :label="department.department_name" :value="department.department_id" />
          </el-select>
        </el-form-item>

        <el-form-item label="Email" prop="email">
          <el-input v-model="formData.email" placeholder="contoh@domain.com" type="email" />
        </el-form-item>

        <el-form-item label="Telepon" prop="phone_number">
          <el-input v-model="formData.phone_number" placeholder="Nomor telepon/HP" />
        </el-form-item>

        <el-form-item label="Keterangan" prop="login_desc">
          <el-input v-model="formData.login_desc" type="textarea" :rows="2"
            placeholder="Keterangan atau deskripsi admin" />
        </el-form-item>

        <el-row :gutter="16">
          <!-- <el-col :span="12">
            <el-form-item label="Status Aktif" prop="flag_use">
              <el-switch
                v-model="formData.flag_use"
                active-text="Aktif"
                inactive-text="Nonaktif"
              />
            </el-form-item>
          </el-col> -->
          <el-col :span="12">
            <el-form-item label="Is Warehouse" prop="is_warehouse">
              <el-switch v-model="formData.is_warehouse" active-text="Ya" inactive-text="Tidak" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">Batal</el-button>
          <el-button type="primary" :loading="submitting" @click="submitForm">
            {{ isMobile ? (isEditing ? 'Simpan' : 'Tambah') : (isEditing ? 'Simpan Perubahan' : 'Tambah Admin') }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, reactive, onMounted, onUnmounted } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElNotification } from 'element-plus'
import {
  Search,
  Refresh,
  Plus,
  Edit,
  Delete,
  UserFilled
} from '@element-plus/icons-vue'
import { adminApi } from '../../api/admin'
import type { Admin, AdminGroup, AdminQueryParams, AdminDepartment } from '../../types/admin'

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px


// Data state
const adminList = shallowRef<Admin[]>([])
const groupOptions = ref<AdminGroup[]>([])
const departmentOptions = ref<AdminDepartment[]>([])
const loading = ref(false)
const submitting = ref(false)

// Dialog state
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()

// Pagination state
const pagination = reactive({
  page: 1,
  limit: 10,
  total: 0
})

const getRowIndex = (index: number) => (pagination.page - 1) * pagination.limit + index + 1

// Search Filter state
const filters = reactive<AdminQueryParams>({
  username: '',
  group_name: ''
})

// Form Data state
const formData = reactive<Partial<Admin>>({
  username: '',
  password: '',
  group_id: undefined,
  department_id: undefined,
  email: '',
  phone_number: '',
  login_desc: '',
  flag_use: true,
  is_warehouse: false
})

// Form validation rules
const formRules: FormRules = {
  username: [
    { required: true, message: 'Username wajib diisi', trigger: 'blur' },
    { min: 3, message: 'Username minimal 3 karakter', trigger: 'blur' }
  ],
  group_id: [
    { required: true, message: 'Group Admin wajib dipilih', trigger: 'change' }
  ]
}

// Request cancellation & debouncing
let currentAbortController: AbortController | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null

/**
 * Fetch Admin list with filters & pagination.
 */
async function fetchAdmins() {
  if (currentAbortController) {
    currentAbortController.abort()
  }

  currentAbortController = new AbortController()
  loading.value = true

  try {
    const res = await adminApi.getAdmins(
      {
        page: pagination.page,
        limit: pagination.limit,
        username: filters.username?.trim(),
        group_name: filters.group_name?.trim()
      },
      currentAbortController.signal
    )

    adminList.value = res.data || []
    pagination.total = res.meta?.total || 0
  } catch (err: any) {
    if (err.name === 'CanceledError' || err.name === 'AbortError') return
    console.error('Failed to fetch admin list:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal memuat data admin')
  } finally {
    loading.value = false
  }
}

/**
 * Fetch Admin Groups for dropdown choices.
 */
async function fetchGroups() {
  try {
    const res = await adminApi.getAdminGroups()
    groupOptions.value = res.data || []
  } catch (err: any) {
    console.error('Failed to fetch admin groups:', err)
  }
}

async function fetchDepartments() {
  try {
    const res = await adminApi.getDepartments()
    departmentOptions.value = res.data || []
  } catch (err: any) {
    console.error('Failed to fetch admin departments:', err)
  }
}

/**
 * Debounced search input handler.
 */
function onFilterChange() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    pagination.page = 1
    fetchAdmins()
  }, 300)
}

/**
 * Reset search filters.
 */
function resetFilters() {
  filters.username = ''
  filters.group_name = ''
  pagination.page = 1
  fetchAdmins()
}

/**
 * Handle pagination changes.
 */
function handlePageChange(newPage: number) {
  pagination.page = newPage
  fetchAdmins()
}

function handleSizeChange(newLimit: number) {
  pagination.limit = newLimit
  pagination.page = 1
  fetchAdmins()
}

/**
 * Open dialog to create a new Admin.
 */
function openCreateDialog() {
  isEditing.value = false
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

/**
 * Open dialog to edit an existing Admin.
 */
function openEditDialog(row: Admin) {
  isEditing.value = true
  editingId.value = row.id ?? null
  formData.username = row.username
  formData.password = ''
  formData.group_id = row.group_id
  formData.department_id = row.department_id
  formData.email = row.email || ''
  formData.phone_number = row.phone_number || ''
  formData.login_desc = row.login_desc || ''
  formData.flag_use = row.flag_use
  formData.is_warehouse = row.is_warehouse
  dialogVisible.value = true
}

/**
 * Reset dialog form fields.
 */
function resetForm() {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  formData.username = ''
  formData.password = ''
  formData.group_id = undefined
  formData.department_id = undefined
  formData.email = ''
  formData.phone_number = ''
  formData.login_desc = ''
  formData.flag_use = true
  formData.is_warehouse = false
}

import { scrollToFormError } from '../../utils/scroll'

/**
 * Submit Form (Create or Update).
 */
async function submitForm() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) {
      scrollToFormError()
      return
    }

    submitting.value = true
    try {
      const payload: Partial<Admin> = {
        username: formData.username,
        group_id: formData.group_id,
        department_id: formData.department_id,
        email: formData.email,
        phone_number: formData.phone_number,
        login_desc: formData.login_desc,
        flag_use: formData.flag_use,
        is_warehouse: formData.is_warehouse
      }

      // Include password if provided
      if (formData.password?.trim()) {
        payload.password = formData.password.trim()
      }

      if (isEditing.value && editingId.value) {
        await adminApi.updateAdmin(editingId.value, payload)
        ElNotification({
          title: 'Berhasil',
          message: `Data admin ${formData.username} berhasil diperbarui`,
          type: 'success'
        })
      } else {
        await adminApi.createAdmin(payload)
        ElNotification({
          title: 'Berhasil',
          message: `Admin ${formData.username} berhasil ditambahkan`,
          type: 'success'
        })
      }

      dialogVisible.value = false
      fetchAdmins()
    } catch (err: any) {
      console.error('Failed to save admin:', err)
      ElMessage.error(err.response?.data?.error || err.message || 'Gagal menyimpan data admin')
      scrollToFormError()
    } finally {
      submitting.value = false
    }
  })
}

/**
 * Handle Admin deletion.
 */
async function handleDelete(id: number) {
  try {
    await adminApi.deleteAdmin(id)
    ElNotification({
      title: 'Berhasil',
      message: 'Data admin berhasil dihapus',
      type: 'success'
    })
    fetchAdmins()
  } catch (err: any) {
    console.error('Failed to delete admin:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal menghapus data admin')
  }
}

/**
 * Helper to format date strings cleanly.
 */
function formatDate(dateStr?: string): string {
  if (!dateStr || dateStr.startsWith('0001-01-01')) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return dateStr
  }
}

onMounted(() => {
  fetchAdmins()
  fetchGroups()
  fetchDepartments()
})

onUnmounted(() => {
  if (currentAbortController) currentAbortController.abort()
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<style scoped>
.admin-container {
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

.font-semibold {
  font-weight: 600;
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

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}
</style>
