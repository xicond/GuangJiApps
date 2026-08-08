<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Master Data Sxy Donatur</h2>
        <p class="page-subtitle">Kelola daftar data sxy donatur, pencarian, serta pembaruan profil</p>
      </div>
      <el-button type="primary" size="large" :icon="Plus" class="create-btn" @click="handleCreate">
        Tambah Sxy Donatur Baru
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
        <el-col :xs="24" :sm="12" :md="6">
          <el-form-item label="Nama Indo" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.nama" placeholder="Cari nama donatur..." clearable :prefix-icon="Search"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="6">
          <el-form-item label="Mandarin" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.mandarin" placeholder="Cari nama mandarin..." clearable :prefix-icon="Search"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="6">
          <el-form-item label="Fotang" :label-position="isMobile ? 'top' : 'right'">
            <LookupSelect v-model="filters.fotang" placeholder="Pilih fotang..." :fetch-api="fotangApi.getFotangLookup"
              clearable @change="onFilterChange" />
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
        empty-text="Tidak ada data sxy donatur yang ditemukan">
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="70" align="center"
          fixed="left" />

        <el-table-column prop="nama" label="Nama Indonesia" min-width="180">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.nama || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="mandarin" label="Nama Mandarin" min-width="140">
          <template #default="{ row }">
            <span>{{ row.mandarin || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="fotang" label="Fotang" min-width="140">
          <template #default="{ row }">
            <span>{{ row.fotang || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="alamat" label="Alamat" min-width="180">
          <template #default="{ row }">
            <span>{{ row.alamat || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="mobile" label="Phone" min-width="140">
          <template #default="{ row }">
            <span>{{ row.mobile !== '-' ? row.mobile : row.telepon || '-' }}</span>
          </template>
        </el-table-column>

        <!-- <el-table-column prop="email" label="Email" min-width="160">
          <template #default="{ row }">
            <span>{{ row.email || '-' }}</span>
          </template>
        </el-table-column> -->

        <el-table-column prop="keterangan" label="Keterangan" min-width="200">
          <template #default="{ row }">
            <span>{{ row.keterangan || '-' }}</span>
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
              <el-button type="primary" size="small" circle :icon="Edit" title="Edit Sxy Donatur"
                @click="handleEdit(row.id!)" />

              <el-popconfirm title="Apakah Anda yakin ingin menghapus data ini?" confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal" confirm-button-type="danger" @confirm="handleDelete(row.id!)">
                <template #reference>
                  <el-button type="danger" size="small" circle :icon="Delete" title="Hapus Sxy Donatur" />
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
    <el-dialog v-model="dialogVisible" :title="isEditing ? `Edit Sxy Donatur #${editingId}` : 'Tambah Sxy Donatur Baru'"
      :width="isMobile ? '90%' : '600px'" destroy-on-close @closed="resetForm">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="140px"
        :label-position="isMobile ? 'top' : 'right'">
        <el-form-item label="Nama Indonesia" prop="nama">
          <el-input v-model="formData.nama" placeholder="Masukkan nama Indonesia" />
        </el-form-item>

        <el-form-item label="Nama Mandarin" prop="mandarin">
          <el-input v-model="formData.mandarin" placeholder="Masukkan nama Mandarin" />
        </el-form-item>

        <el-form-item label="Fotang" prop="lookup_fothang">
          <LookupSelect v-model="formData.lookup_fothang" placeholder="Pilih Fotang..."
            :fetch-api="fotangApi.getFotangLookupSxy" clearable />
        </el-form-item>

        <!-- <el-form-item label="No. Donatur" prop="no">
          <el-input
            v-model="formData.no"
            placeholder="Masukkan nomor donatur"
          />
        </el-form-item> -->

        <el-form-item label="Alamat" prop="alamat">
          <el-input v-model="formData.alamat" type="textarea" :rows="2" placeholder="Masukkan alamat lengkap..." />
        </el-form-item>

        <el-form-item label="Telepon" prop="telepon">
          <el-input v-model="formData.telepon" placeholder="Masukkan nomor telepon rumah/kantor" />
        </el-form-item>

        <el-form-item label="Mobile Phone" prop="mobile">
          <el-input v-model="formData.mobile" placeholder="Masukkan nomor handphone" />
        </el-form-item>

        <el-form-item label="Email" prop="email">
          <el-input v-model="formData.email" placeholder="contoh@domain.com" type="email" />
        </el-form-item>

        <el-form-item label="Keterangan" prop="keterangan">
          <el-input v-model="formData.keterangan" type="textarea" :rows="3" placeholder="Masukkan keterangan..." />
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
import { sxyDonaturApi } from '../../api/sxyDonatur'
import { fotangApi } from '../../api/fotang'
import LookupSelect from '../../components/common/LookupSelect.vue'
import type { SxyDonatur, SxyDonaturQueryParams } from '../../types/sxyDonatur'

// Breakpoints layout calculation
const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')
const isDesktop = breakpoints.greaterOrEqual('lg')

// Dataset state
const dataList = shallowRef<SxyDonatur[]>([])
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
const filters = reactive<SxyDonaturQueryParams>({
  nama: '',
  mandarin: '',
  fotang: ''
})

// Dialog & Form state
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance | null>(null)

const formData = reactive<Partial<SxyDonatur>>({
  no: '',
  nama: '',
  mandarin: '',
  lookup_fothang: undefined,
  alamat: '',
  telepon: '',
  mobile: '',
  email: '',
  keterangan: '',
  // status: true
})

const formRules: FormRules = {
  nama: [
    { required: true, message: 'Nama Indonesia wajib diisi', trigger: 'blur' },
    { max: 50, message: 'Nama Indonesia maksimal 50 karakter', trigger: 'blur' }
  ],
  mandarin: [
    { max: 50, message: 'Nama Mandarin maksimal 50 karakter', trigger: 'blur' }
  ],
  no: [
    { max: 10, message: 'Nomor donatur maksimal 10 karakter', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: 'Format email tidak valid', trigger: ['blur', 'change'] },
    { max: 100, message: 'Email maksimal 100 karakter', trigger: 'blur' }
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
    const res = await sxyDonaturApi.getSxyDonaturs(
      {
        page: pagination.page,
        limit: pagination.limit,
        nama: filters.nama?.trim(),
        mandarin: filters.mandarin?.trim(),
        fotang: filters.fotang
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
  filters.nama = ''
  filters.mandarin = ''
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

function resetForm() {
  formData.no = ''
  formData.nama = ''
  formData.mandarin = ''
  formData.lookup_fothang = undefined
  formData.alamat = ''
  formData.telepon = ''
  formData.mobile = ''
  formData.email = ''
  formData.keterangan = ''
  // formData.status = true
  isEditing.value = false
  editingId.value = null
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

function handleCreate() {
  resetForm()
  isEditing.value = false
  dialogVisible.value = true
}

async function handleEdit(id: number) {
  resetForm()
  isEditing.value = true
  editingId.value = id

  try {
    const res = await sxyDonaturApi.getSxyDonaturById(id)
    const record = res.data
    if (record) {
      formData.no = record.no || ''
      formData.nama = record.nama
      formData.mandarin = record.mandarin || ''
      formData.lookup_fothang = record.lookup_fothang || undefined
      formData.alamat = record.alamat || ''
      formData.telepon = record.telepon || ''
      formData.mobile = record.mobile || ''
      formData.email = record.email || ''
      formData.keterangan = record.keterangan || ''
      // formData.status = record.status !== undefined ? record.status : true
      dialogVisible.value = true
    }
  } catch (err: unknown) {
    console.error('Error fetching record by ID:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal memuat detail data')
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
    const payload: Partial<SxyDonatur> = {
      no: formData.no?.trim() || undefined,
      nama: formData.nama?.trim(),
      mandarin: formData.mandarin?.trim() || undefined,
      lookup_fothang: formData.lookup_fothang ? Number(formData.lookup_fothang) : undefined,
      alamat: formData.alamat?.trim() || undefined,
      telepon: formData.telepon?.trim() || undefined,
      mobile: formData.mobile?.trim() || undefined,
      email: formData.email?.trim() || undefined,
      keterangan: formData.keterangan?.trim() || undefined,
      // status: formData.status
    }

    if (isEditing.value && editingId.value) {
      await sxyDonaturApi.updateSxyDonatur(editingId.value, payload)
      ElNotification({
        title: 'Berhasil',
        message: `Sxy Donatur #${editingId.value} berhasil diperbarui`,
        type: 'success'
      })
    } else {
      await sxyDonaturApi.createSxyDonatur(payload)
      ElNotification({
        title: 'Berhasil',
        message: 'Sxy Donatur baru berhasil ditambahkan',
        type: 'success'
      })
    }

    dialogVisible.value = false
    fetchData()
  } catch (err: unknown) {
    console.error('Failed to submit form:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal menyimpan data')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await sxyDonaturApi.deleteSxyDonatur(id)
    ElNotification({
      title: 'Berhasil',
      message: `Data Sxy Donatur #${id} berhasil dihapus`,
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
