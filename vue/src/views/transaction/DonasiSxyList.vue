<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Transaksi Donasi Sxy</h2>
        <p class="page-subtitle">Kelola daftar data donasi sxy, pencarian, serta pembaruan profil</p>
      </div>
      <el-button type="primary" size="large" :icon="Plus" class="create-btn" @click="handleCreate">
        Tambah Donasi Sxy Baru
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
        <el-col :xs="24" :sm="12" :md="7">
          <el-form-item label="No. Kwitansi" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.no_kwitansi" placeholder="Cari no kwitansi..." clearable :prefix-icon="Search"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12" :md="7">
          <el-form-item label="Donatur" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.donatur" placeholder="Cari donatur..." clearable :prefix-icon="Search"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12" :md="10">
          <el-form-item label="Periode" :label-position="isMobile ? 'top' : 'right'">
            <el-date-picker v-model="filters.date_range" type="daterange" range-separator="s/d"
              start-placeholder="Start Date" end-placeholder="End Date" value-format="YYYY-MM-DD"
              :disabled-date="disabledDate" @change="fetchData" :clearable="false" style="width: 100%"
              :single-panel="isMobile" />
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
        empty-text="Tidak ada data donasi sxy yang ditemukan">
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="70" align="center"
          fixed="left" />

        <el-table-column prop="no_kwitansi" label="No. Kwitansi" width="140" align="left">
          <template #default="{ row }">
            <el-tag size="small" type="info" class="font-mono">{{ row.no_kwitansi }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="tanggal" label="Tanggal" min-width="120" align="center">
          <template #default="{ row }">
            <span>{{ row.tanggal || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="nama_donatur" label="Donatur" min-width="160">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.nama_donatur || '-' }}</span>
          </template>
        </el-table-column>

        <!-- <el-table-column prop="nama_penggalang" label="Penggalang" min-width="140">
          <template #default="{ row }">
            <span>{{ row.nama_penggalang || '-' }}</span>
          </template>
        </el-table-column> -->

        <el-table-column prop="atas_nama" label="Atas Nama" min-width="140">
          <template #default="{ row }">
            <span>{{ row.atas_nama || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="tipe_sumbangan_desc" label="Tipe Sumbangan" min-width="140">
          <template #default="{ row }">
            <span>{{ row.tipe_sumbangan_desc || row.tipe_sumbangan || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="jumlah" label="Jumlah (Rp)" min-width="150" align="right">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.jumlah ? 'Rp ' + Number(row.jumlah).toLocaleString('id-ID') : '-'
            }}</span>
          </template>
        </el-table-column>

        <!-- <el-table-column prop="status" label="Status" width="100" align="center">
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
              <el-button type="primary" size="small" circle :icon="Edit" title="Edit Donasi Sxy"
                @click="handleEdit(row.id!)" />

              <el-popconfirm title="Apakah Anda yakin ingin menghapus data ini?" confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal" confirm-button-type="danger" @confirm="handleDelete(row.id!)">
                <template #reference>
                  <el-button type="danger" size="small" circle :icon="Delete" title="Hapus Donasi Sxy" />
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
          :layout="isDesktop ? 'total, sizes, prev, pager, next, jumper' : 'total, sizes, prev, pager, next'"
          @size-change="handleSizeChange" @current-change="handlePageChange" />
      </div>
    </el-card>

    <!-- Create / Edit Dialog -->
    <el-dialog v-model="dialogVisible" :title="isEditing ? `Edit Donasi Sxy #${editingId}` : 'Tambah Donasi Sxy Baru'"
      :width="isMobile ? '90%' : '600px'" destroy-on-close @closed="resetForm">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="150px"
        :label-position="isMobile ? 'top' : 'right'">
        <el-form-item label="No. Kwitansi" prop="no_kwitansi">
          <el-input v-model="formData.no_kwitansi" placeholder="Masukkan no. kwitansi" />
        </el-form-item>

        <el-form-item label="Tanggal" prop="tanggal">
          <el-date-picker v-model="formData.tanggal" type="date" placeholder="Pilih tanggal" value-format="YYYY-MM-DD"
            format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>

        <el-form-item label="Donatur" prop="donatur_id">
          <LookupSelect v-model="formData.donatur_id" placeholder="Pilih Donatur..."
            :fetch-api="sxyDonaturApi.getSxyDonaturs" :get-item-api="sxyDonaturApi.getSxyDonaturById" value-key="id"
            label-key="nama" clearable />
        </el-form-item>

        <el-form-item label="Tipe Sumbangan" prop="tipe_sumbangan">
          <LookupSelect v-model="formData.tipe_sumbangan" placeholder="Pilih Tipe Sumbangan..."
            :fetch-api="lookupApi.getLookupTipeSumbangan" value-key="lookup_value" label-key="lookup_description"
            clearable />
        </el-form-item>

        <el-form-item label="Jumlah (Rp)" prop="jumlah">
          <el-input-number v-model="formData.jumlah" :min="1000" :precision="0" controls-position="right"
            style="width: 100%" placeholder="Masukkan nominal sumbangan" />
        </el-form-item>

        <el-form-item label="Penggalang Dana" prop="penggalang_id">
          <LookupSelect v-model="formData.penggalang_id" placeholder="Pilih Penggalang Dana..."
            :fetch-api="penggalangDanaApi.getPenggalangDanas" :get-item-api="penggalangDanaApi.getPenggalangDanaById"
            value-key="id" label-key="nama" clearable />
        </el-form-item>

        <el-form-item label="No. Kupon" prop="no_kupon">
          <el-input v-model="formData.no_kupon" placeholder="Masukkan nomor kupon (opsional)" />
        </el-form-item>

        <el-form-item label="Tanggal Transfer" prop="tanggal_transfer">
          <el-date-picker v-model="formData.tanggal_transfer" type="date"
            placeholder="Pilih tanggal transfer (opsional)" value-format="YYYY-MM-DD" format="YYYY-MM-DD"
            style="width: 100%" />
        </el-form-item>

        <el-form-item label="Atas Nama" prop="atas_nama">
          <el-input v-model="formData.atas_nama" placeholder="Masukkan atas nama rekening..." />
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
import dayjs from 'dayjs'
import { donasiSxyApi } from '../../api/donasiSxy'
import { sxyDonaturApi } from '../../api/sxyDonatur'
import { penggalangDanaApi } from '../../api/penggalangDana'
import { lookupApi } from '../../api/lookup'
import LookupSelect from '../../components/common/LookupSelect.vue'
import type { DonasiSxy, DonasiSxyQueryParams } from '../../types/donasiSxy'

// Breakpoints layout calculation
const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')
const isDesktop = breakpoints.greaterOrEqual('lg')

// Dataset state
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

const disabledDate = (time: Date) => {
  return time.getTime() > Date.now()
}

const today = dayjs().format('YYYY-MM-DD')
type DonasiSxyFilter = Omit<DonasiSxyQueryParams, 'start_date' | 'end_date'> & {
  date_range?: string[] | null
}

// Search Filter state
const filters = reactive<DonasiSxyFilter>({
  no_kwitansi: '',
  date_range: [today, today] as [string, string],
  donatur: ''
})

// Dialog & Form state
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance | null>(null)

const formData = reactive<Partial<DonasiSxy>>({
  no_kwitansi: '',
  tanggal: today,
  donatur_id: undefined,
  penggalang_id: undefined,
  tipe_sumbangan: undefined,
  jumlah: 0,
  no_kupon: '',
  tanggal_transfer: '',
  atas_nama: '',
  keterangan: '',
  // status: true
})

const formRules: FormRules = {
  no_kwitansi: [
    { required: true, message: 'No. Kwitansi wajib diisi', trigger: 'blur' },
    { max: 50, message: 'No. Kwitansi maksimal 50 karakter', trigger: 'blur' }
  ],
  tanggal: [
    { required: true, message: 'Tanggal sumbangan wajib diisi', trigger: 'blur' }
  ],
  donatur_id: [
    { required: true, message: 'Donatur wajib diisi', trigger: 'blur' }
  ],
  tipe_sumbangan: [
    { required: true, message: 'Tipe sumbangan wajib diisi', trigger: 'blur' }
  ],
  jumlah: [
    { required: true, message: 'Jumlah sumbangan wajib diisi', trigger: 'blur' }
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
  const [start_date, end_date] = filters.date_range || [null, null]

  try {
    const res = await donasiSxyApi.getDonasiSxys(
      {
        page: pagination.page,
        limit: pagination.limit,
        no_kwitansi: filters.no_kwitansi?.trim(),
        donatur: filters.donatur?.trim(),
        start_date: start_date || undefined,
        end_date: end_date || undefined
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
  filters.no_kwitansi = ''
  filters.donatur = ''
  filters.date_range = [today, today]
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
  formData.no_kwitansi = ''
  formData.tanggal = today
  formData.donatur_id = undefined
  formData.penggalang_id = undefined
  formData.tipe_sumbangan = undefined
  formData.jumlah = 0
  formData.no_kupon = ''
  formData.tanggal_transfer = ''
  formData.atas_nama = ''
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
    const res = await donasiSxyApi.getDonasiSxyById(id)
    const record = res.data
    if (record) {
      formData.no_kwitansi = record.no_kwitansi
      formData.tanggal = record.tanggal || ''
      formData.donatur_id = record.donatur_id || undefined
      formData.penggalang_id = record.penggalang_id || undefined
      formData.tipe_sumbangan = record.tipe_sumbangan ? Number(record.tipe_sumbangan) : undefined
      formData.jumlah = record.jumlah || 0
      formData.no_kupon = record.no_kupon || ''
      formData.tanggal_transfer = record.tanggal_transfer || ''
      formData.atas_nama = record.atas_nama || ''
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
    const payload: Partial<DonasiSxy> = {
      no_kwitansi: formData.no_kwitansi?.trim(),
      tanggal: formData.tanggal || undefined,
      donatur_id: formData.donatur_id ? Number(formData.donatur_id) : undefined,
      penggalang_id: formData.penggalang_id ? Number(formData.penggalang_id) : undefined,
      tipe_sumbangan: formData.tipe_sumbangan ? Number(formData.tipe_sumbangan) : undefined,
      jumlah: formData.jumlah ? Number(formData.jumlah) : 0,
      no_kupon: formData.no_kupon?.trim() || undefined,
      tanggal_transfer: formData.tanggal_transfer || undefined,
      atas_nama: formData.atas_nama?.trim() || undefined,
      keterangan: formData.keterangan?.trim() || undefined,
      // status: formData.status
    }

    if (isEditing.value && editingId.value) {
      await donasiSxyApi.updateDonasiSxy(editingId.value, payload)
      ElNotification({
        title: 'Berhasil',
        message: `Donasi Sxy #${editingId.value} berhasil diperbarui`,
        type: 'success'
      })
    } else {
      await donasiSxyApi.createDonasiSxy(payload)
      ElNotification({
        title: 'Berhasil',
        message: 'Donasi Sxy baru berhasil ditambahkan',
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
    await donasiSxyApi.deleteDonasiSxy(id)
    ElNotification({
      title: 'Berhasil',
      message: `Data Donasi Sxy #${id} berhasil dihapus`,
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
