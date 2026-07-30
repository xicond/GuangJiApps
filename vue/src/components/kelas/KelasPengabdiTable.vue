<template>
  <div class="pengabdi-accordion-container">
    <el-collapse v-model="activeNames" class="custom-accordion" @change="handleAccordionChange">
      <el-collapse-item name="pengabdi">
        <template #title>
          <div class="accordion-header" @click.stop>
            <div class="header-title">
              <el-icon class="header-icon"><Avatar /></el-icon>
              <span>Daftar Pengabdi Kelas</span>
              <el-tag size="small" type="info" class="ml-2">{{ total }} Pengabdi</el-tag>
            </div>
            <div class="header-actions" @click.stop>
              <el-button
                type="primary"
                size="small"
                :icon="Plus"
                @click.stop="openAddDialog"
              >
                Tambah Pengabdi
              </el-button>
              <el-button
                :icon="Refresh"
                circle
                size="small"
                title="Refresh Pengabdi"
                @click.stop="fetchPengabdi"
              />
            </div>
          </div>
        </template>

        <div class="accordion-content">
          <el-table
            v-loading="loading"
            :data="pengabdiList"
            stripe
            border
            max-height="450"
            style="width: 100%"
            empty-text="Belum ada pengabdi yang terdaftar pada kelas ini"
          >
            <el-table-column type="index" label="No." width="60" align="center" fixed="left" />

            <el-table-column prop="id_pengabdi" label="ID / Pengabdi" min-width="160">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.nama_indonesia || row.id_pengabdi || '-' }}</span>
                <span v-if="row.nama_mandarin" class="sub-text"> ({{ row.nama_mandarin }})</span>
              </template>
            </el-table-column>

            <el-table-column prop="tim_kerja" label="Tim Kerja" width="120" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.tim_kerja" size="small" type="primary">{{ row.tim_kerja }}</el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>

            <el-table-column prop="tim_kerja_report" label="Tim Kerja Report" width="140" align="center" />

            <el-table-column prop="hari" label="Hari" width="120" align="center" />

            <el-table-column prop="keterangan" label="Keterangan" min-width="160" />

            <el-table-column label="Aksi" width="140" align="center" fixed="right">
              <template #default="{ row }">
                <el-button
                  type="primary"
                  link
                  size="small"
                  :icon="Edit"
                  @click="openEditDialog(row)" />
                <el-popconfirm
                  title="Yakin ingin menghapus pengabdi ini?"
                  confirm-button-text="Ya, Hapus"
                  cancel-button-text="Batal"
                  confirm-button-type="danger"
                  @confirm="handleDelete(row)"
                >
                  <template #reference>
                    <el-button type="danger" link size="small" :icon="Delete" />
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>

          <!-- Pagination -->
          <div class="pagination-container">
            <el-pagination
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              :total="total"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </div>
      </el-collapse-item>
    </el-collapse>

    <!-- Dialog Popup Add / Edit Pengabdi -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'add' ? 'Tambah Pengabdi Kelas' : 'Edit Pengabdi Kelas'"
      width="640px"
      destroy-on-close
    >
      <el-alert
        v-if="Object.keys(dialogFieldErrors).length > 0"
        type="error"
        show-icon
        title="Validasi Gagal"
        description="Terdapat kesalahan pengisian form. Silahkan periksa pesan kesalahan berwarna merah di bawah."
        class="mb-4"
      />

      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="140px"
        size="default"
      >
        <el-form-item label="Pengabdi (Umat)" prop="id_pengabdi" :error="hasFieldError('id_pengabdi') ? ' ' : undefined">
          <el-select
            v-model="form.id_pengabdi"
            filterable
            remote
            reserve-keyword
            placeholder="Ketik nama untuk mencari Umat..."
            :remote-method="searchUmat"
            :loading="loadingUmat"
            style="width: 100%"
          >
            <el-option
              v-for="item in umatOptions"
              :key="item.id"
              :label="getUmatOptionLabel(item)"
              :value="item.id"
            />
          </el-select>
          <FieldErrors :errors="getFieldErrors('id_pengabdi')" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Tim Kerja" :error="hasFieldError('tim_kerja') ? ' ' : undefined">
              <el-input v-model="form.tim_kerja" placeholder="Kode tim kerja" maxlength="3" />
              <FieldErrors :errors="getFieldErrors('tim_kerja')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Sub Kerja" :error="hasFieldError('sub_kerja') ? ' ' : undefined">
              <el-input v-model="form.sub_kerja" placeholder="Sub kerja" maxlength="3" />
              <FieldErrors :errors="getFieldErrors('sub_kerja')" />
            </el-form-item>
          </el-col>

          <!-- masuk ke Tim Kerja Selective -->
          <!-- <el-col :span="12">
            <el-form-item label="Tim Kerja Report">
              <el-input v-model="form.tim_kerja_report" placeholder="Tim kerja report" maxlength="3" />
            </el-form-item>
          </el-col> -->
        </el-row>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Hari" :error="hasFieldError('hari') ? ' ' : undefined">
              <el-input v-model="form.hari" placeholder="Hari pengabdian" maxlength="30" />
              <FieldErrors :errors="getFieldErrors('hari')" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Sumbangan" :error="hasFieldError('sumbangan') ? ' ' : undefined">
              <el-input-number
                v-model="form.sumbangan"
                :min="0"
                :precision="2"
                :step="10000"
                controls-position="right"
                style="width: 100%"
              />
              <FieldErrors :errors="getFieldErrors('sumbangan')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Barang" :error="hasFieldError('barang') ? ' ' : undefined">
              <el-input v-model="form.barang" placeholder="Barang sumbangan" maxlength="100" />
              <FieldErrors :errors="getFieldErrors('barang')" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="Keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
          <el-input v-model="form.keterangan" placeholder="Keterangan pengabdi" maxlength="200" />
          <FieldErrors :errors="getFieldErrors('keterangan')" />
        </el-form-item>

        <!-- Logistik Optional Fields -->
        <el-divider content-position="left">Data Logistik / Kehadiran</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Anak" :error="hasFieldError('anak') ? ' ' : undefined">
              <el-input v-model="form.anak" placeholder="Detail anak" maxlength="30" />
              <FieldErrors :errors="getFieldErrors('anak')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Suster" :error="hasFieldError('suster') ? ' ' : undefined">
              <el-input v-model="form.suster" placeholder="Detail suster" maxlength="30" />
              <FieldErrors :errors="getFieldErrors('suster')" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Menginap" :error="hasFieldError('menginap') ? ' ' : undefined">
              <el-input v-model="form.menginap" placeholder="Detail menginap" maxlength="30" />
              <FieldErrors :errors="getFieldErrors('menginap')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Makanan Pagi" :error="hasFieldError('makanan_pagi') ? ' ' : undefined">
              <el-input v-model="form.makanan_pagi" placeholder="Makanan pagi" maxlength="30" />
              <FieldErrors :errors="getFieldErrors('makanan_pagi')" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Makanan Siang" :error="hasFieldError('makanan_siang') ? ' ' : undefined">
              <el-input v-model="form.makanan_siang" placeholder="Makanan siang" maxlength="30" />
              <FieldErrors :errors="getFieldErrors('makanan_siang')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Makanan Malam" :error="hasFieldError('makanan_malam') ? ' ' : undefined">
              <el-input v-model="form.makanan_malam" placeholder="Makanan malam" maxlength="30" />
              <FieldErrors :errors="getFieldErrors('makanan_malam')" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">Batal</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">
          {{ dialogMode === 'add' ? 'Simpan' : 'Perbarui' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Avatar, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { kelasApi } from '../../api/kelas'
import { umatApi } from '../../api/umat'
import FieldErrors from '../common/FieldErrors.vue'
import type { KelasPengabdi } from '../../types/kelas'
import type { Umat } from '../../types/umat'

const props = defineProps<{
  kelasId: string | number
}>()

const activeNames = ref<string[]>([])
const isExpanded = computed(() => activeNames.value.includes('pengabdi'))

const pengabdiList = ref<KelasPengabdi[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
let abortController: AbortController | null = null

// Dialog state
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const submitting = ref(false)
const formRef = ref<FormInstance>()
const dialogFieldErrors = ref<Record<string, string[]>>({})

function getFieldErrors(fieldName: string): string[] {
  return dialogFieldErrors.value[fieldName] || []
}

function hasFieldError(fieldName: string): boolean {
  return getFieldErrors(fieldName).length > 0
}

interface PengabdiFormState {
  detail_id?: number | string
  trx_id?: number | string
  id_pengabdi?: number | string
  tim_kerja: string
  tim_kerja_report: string
  hari: string
  sub_kerja: string
  sumbangan?: number
  barang: string
  keterangan: string
  anak: string
  suster: string
  menginap: string
  makanan_pagi: string
  makanan_siang: string
  makanan_malam: string
}

const form = ref<PengabdiFormState>({
  tim_kerja: '',
  tim_kerja_report: '',
  hari: '',
  sub_kerja: '',
  barang: '',
  keterangan: '',
  anak: '',
  suster: '',
  menginap: '',
  makanan_pagi: '',
  makanan_siang: '',
  makanan_malam: ''
})

const formRules: FormRules = {
  id_pengabdi: [{ required: true, message: 'Harap pilih pengabdi / umat', trigger: 'change' }]
}

const umatOptions = ref<Umat[]>([])
const loadingUmat = ref(false)

async function searchUmat(query: string) {
  if (!query || query.trim() === '') {
    umatOptions.value = []
    return
  }
  loadingUmat.value = true
  try {
    const res = await umatApi.getUmats({ namaindonesia: query.trim(), limit: 20 })
    umatOptions.value = res.data || []
  } catch (err) {
    console.error('Error searching umat:', err)
  } finally {
    loadingUmat.value = false
  }
}

function getUmatOptionLabel(item: Umat): string {
  let label = item.nama_indonesia || item.kode || `ID #${item.id}`
  if (item.nama_mandarin) label += ` (${item.nama_mandarin})`
  if (item.alias) label += ` - ${item.alias}`
  return label
}

function handleAccordionChange(val: string[] | string) {
  const activeList = Array.isArray(val) ? val : [val]
  if (activeList.includes('pengabdi')) {
    fetchPengabdi()
  }
}

async function fetchPengabdi() {
  if (!props.kelasId || !isExpanded.value) return

  if (abortController) abortController.abort()
  abortController = new AbortController()

  loading.value = true
  try {
    const res = await kelasApi.getKelasPengabdi(
      props.kelasId,
      { page: currentPage.value, limit: pageSize.value },
      abortController.signal
    )
    pengabdiList.value = res.data || []
    total.value = res.meta?.total ?? pengabdiList.value.length
  } catch (err: any) {
    if (err.name === 'AbortError' || err.name === 'CanceledError') return
    console.error('Error fetching kelas pengabdi:', err)
  } finally {
    loading.value = false
  }
}

function handleSizeChange(newSize: number) {
  pageSize.value = newSize
  currentPage.value = 1
  if (isExpanded.value) fetchPengabdi()
}

function handleCurrentChange(newPage: number) {
  currentPage.value = newPage
  if (isExpanded.value) fetchPengabdi()
}

function openAddDialog() {
  dialogMode.value = 'add'
  dialogFieldErrors.value = {}
  form.value = {
    trx_id: props.kelasId,
    id_pengabdi: undefined,
    tim_kerja: '',
    tim_kerja_report: '',
    hari: '',
    sub_kerja: '',
    sumbangan: undefined,
    barang: '',
    keterangan: '',
    anak: '',
    suster: '',
    menginap: '',
    makanan_pagi: '',
    makanan_siang: '',
    makanan_malam: ''
  }
  umatOptions.value = []
  dialogVisible.value = true
}

function openEditDialog(row: KelasPengabdi) {
  dialogMode.value = 'edit'
  dialogFieldErrors.value = {}
  const detailId = row.detailid || row.detail_id
  const rawIdPengabdi = row.id_pengabdi ? Number(row.id_pengabdi) : undefined

  form.value = {
    detail_id: detailId,
    trx_id: props.kelasId,
    id_pengabdi: rawIdPengabdi,
    tim_kerja: row.tim_kerja || '',
    tim_kerja_report: row.tim_kerja_report || '',
    hari: row.hari || '',
    sub_kerja: row.sub_kerja || '',
    sumbangan: row.sumbangan,
    barang: row.barang || '',
    keterangan: row.keterangan || '',
    anak: row.anak || '',
    suster: row.suster || '',
    menginap: row.menginap || '',
    makanan_pagi: row.makanan_pagi || '',
    makanan_siang: row.makanan_siang || '',
    makanan_malam: row.makanan_malam || ''
  }

  if (rawIdPengabdi && row.nama_indonesia) {
    umatOptions.value = [
      {
        id: rawIdPengabdi,
        kode: '',
        nama_indonesia: row.nama_indonesia,
        nama_mandarin: row.nama_mandarin
      } as Umat
    ]
  } else {
    umatOptions.value = []
  }

  dialogVisible.value = true
}

async function submitForm() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    dialogFieldErrors.value = {}
    try {
      const payload: Partial<KelasPengabdi> = {
        trx_id: props.kelasId ? Number(props.kelasId) : undefined,
        id_pengabdi: form.value.id_pengabdi ? Number(form.value.id_pengabdi) : undefined,
        tim_kerja: form.value.tim_kerja,
        tim_kerja_report: form.value.tim_kerja_report,
        hari: form.value.hari,
        sub_kerja: form.value.sub_kerja,
        sumbangan: form.value.sumbangan,
        barang: form.value.barang,
        keterangan: form.value.keterangan,
        anak: form.value.anak,
        suster: form.value.suster,
        menginap: form.value.menginap,
        makanan_pagi: form.value.makanan_pagi,
        makanan_siang: form.value.makanan_siang,
        makanan_malam: form.value.makanan_malam
      }

      if (dialogMode.value === 'add') {
        await kelasApi.createKelasPengabdi(payload)
        ElMessage.success('Pengabdi kelas berhasil ditambahkan')
      } else {
        const detailId = form.value.detail_id
        if (!detailId) return
        await kelasApi.updateKelasPengabdi(detailId, payload)
        ElMessage.success('Data pengabdi kelas berhasil diperbarui')
      }

      dialogVisible.value = false
      if (isExpanded.value) fetchPengabdi()
    } catch (err: any) {
      console.error('Error submitting pengabdi form:', err)
      if (err.response?.data?.details) {
        dialogFieldErrors.value = err.response.data.details
        ElMessage.error(err.response.data.error || 'Validasi gagal, Silahkan periksa kolom form')
      } else {
        ElMessage.error(err.response?.data?.error || err.message || 'Gagal menyimpan data pengabdi')
      }
    } finally {
      submitting.value = false
    }
  })
}

async function handleDelete(row: KelasPengabdi) {
  const detailId = row.detailid || row.detail_id
  if (!detailId) return

  try {
    await kelasApi.deleteKelasPengabdi(detailId, props.kelasId)
    ElMessage.success('Pengabdi kelas berhasil dihapus')
    if (isExpanded.value) fetchPengabdi()
  } catch (err: any) {
    console.error('Error deleting pengabdi:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal menghapus data pengabdi')
  }
}

watch(
  () => props.kelasId,
  (newId) => {
    if (newId) {
      currentPage.value = 1
      if (isExpanded.value) {
        fetchPengabdi()
      } else {
        pengabdiList.value = []
        total.value = 0
      }
    }
  }
)

onMounted(() => {
  if (isExpanded.value) fetchPengabdi()
})

onUnmounted(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.pengabdi-accordion-container {
  margin-top: 1.25rem;
}

.custom-accordion {
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--el-border-color-light);
  background-color: var(--el-bg-color-overlay);
}

:deep(.el-collapse-item__header) {
  padding: 0.75rem 1rem;
  height: auto;
  min-height: 48px;
  background-color: var(--el-bg-color-overlay);
}

:deep(.el-collapse-item__wrap) {
  background-color: var(--el-bg-color-overlay);
  border-bottom: none;
}

:deep(.el-collapse-item__content) {
  padding: 1rem;
}

.accordion-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding-right: 0.5rem;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  font-size: 1rem;
  color: var(--el-text-color-primary);
}

.header-icon {
  color: var(--el-color-primary);
  font-size: 1.2rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.font-semibold {
  font-weight: 600;
}

.sub-text {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}
</style>
