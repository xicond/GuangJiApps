<template>
  <div class="pengeluaran-accordion-container">
    <el-collapse v-model="activeNames" class="custom-accordion" @change="handleAccordionChange">
      <el-collapse-item name="pengeluaran">
        <template #title>
          <div class="accordion-header">
            <div class="header-title">
              <el-icon class="header-icon">
                <Wallet />
              </el-icon>
              <span>{{ isDesktop ? 'Daftar Pengeluaran Kelas' : 'Pengeluaran' }}</span>
              <el-tag v-if="isExpanded && !isMobile" size="small" type="info" class="ml-2">{{ total }}
                Pengeluaran</el-tag>
            </div>
            <div class="header-actions" @click.stop>
              <el-button v-if="!readonly" type="primary" size="small" :icon="Plus" @click.stop="openAddDialog">
                {{ isMobile ? '' : 'Tambah Pengeluaran' }}
              </el-button>
              <el-button :icon="Refresh" circle size="small" title="Refresh Pengeluaran"
                @click.stop="fetchPengeluaran" />
            </div>
          </div>
        </template>

        <div class="accordion-content">
          <el-table v-loading="loading" :data="pengeluaranList" stripe border max-height="450" style="width: 100%"
            empty-text="Belum ada pengeluaran yang terdaftar pada kelas ini">
            <el-table-column v-if="isDesktop" type="index" label="No." width="60" align="center" fixed="left" />

            <el-table-column prop="tim_kerja" label="Tim Kerja" min-width="140">
              <template #default="{ row }">
                <span>{{ row.tim_kerja || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="keterangan" label="Keterangan" min-width="220">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.keterangan || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="biaya" label="Biaya (Rp)" min-width="160" align="right">
              <template #default="{ row }">
                <span class="font-mono text-danger">{{ formatCurrency(row.biaya) }}</span>
              </template>
            </el-table-column>

            <el-table-column v-if="!readonly" label="Aksi" width="90" align="center"
              :fixed="!isDesktop ? false : 'right'">
              <template #default="{ row }">
                <el-button type="primary" circle size="small" :icon="Edit" @click="openEditDialog(row)" />
                <el-popconfirm title="Yakin ingin menghapus pengeluaran ini?" confirm-button-text="Ya, Hapus"
                  cancel-button-text="Batal" confirm-button-type="danger" @confirm="handleDelete(row)">
                  <template #reference>
                    <el-button type="danger" circle size="small" :icon="Delete" />
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>

          <!-- Pagination -->
          <div class="pagination-container">
            <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :layout="(!isMobile ? 'total, ->,' : (Math.ceil(total / pageSize) < 6 ? '-> ,' : '')) + 'prev, pager, next' + (isDesktop ? ', jumper' : '')"
              :pager-count="6" :total="total" @size-change="handleSizeChange" @current-change="handleCurrentChange" />
          </div>
        </div>
      </el-collapse-item>
    </el-collapse>

    <!-- Dialog Popup Add / Edit Pengeluaran -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'add' ? 'Tambah Pengeluaran' : 'Edit Pengeluaran'"
      :width="isMobile ? '90%' : '600px'" destroy-on-close>
      <el-alert v-if="Object.keys(dialogFieldErrors).length > 0" type="error" show-icon title="Invalid Inputs"
        description="Terdapat kesalahan pengisian form. Silahkan periksa pesan kesalahan berwarna merah di bawah."
        class="mb-4" />

      <el-form ref="formRef" :model="form" :rules="formRules" label-width="130px" size="default" v-loading="submitting">
        <el-form-item label="Tim Kerja" prop="tim_kerja" :error="hasFieldError('tim_kerja') ? ' ' : undefined">
          <LookupSelect v-model="form.tim_kerja" :fetch-api="lookupApi.getLookupTimKerja" value-key="lookup_value"
            placeholder="Pilih Tim Kerja..." :clearable="true" auto-populate />
          <FieldErrors :errors="getFieldErrors('tim_kerja')" />
        </el-form-item>

        <el-form-item label="Keterangan" prop="keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
          <el-input v-model="form.keterangan" type="textarea" :rows="2" placeholder="Keterangan pengeluaran"
            maxlength="200" show-word-limit />
          <FieldErrors :errors="getFieldErrors('keterangan')" />
        </el-form-item>

        <el-form-item label="Jumlah Biaya" prop="biaya" :error="hasFieldError('biaya') ? ' ' : undefined">
          <el-input-number v-model="form.biaya" :min="0" :precision="2" :step="1000" controls-position="right"
            style="width: 100%" />
          <FieldErrors :errors="getFieldErrors('biaya')" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button :disabled="submitting" @click="dialogVisible = false">Batal</el-button>
        <el-button type="primary" :loading="submitting" :disabled="submitting" @click="submitForm">
          {{ dialogMode === 'add' ? 'Simpan' : 'Perbarui' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import { Wallet, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { kelasApi } from '../../api/kelas'
import { lookupApi } from '../../api/lookup'
import FieldErrors from '../common/FieldErrors.vue'
import LookupSelect from '../common/LookupSelect.vue'
import type { KelasPengeluaran } from '../../types/kelas'

const props = withDefaults(
  defineProps<{
    kelasId: string | number
    readonly?: boolean
  }>(),
  {
    readonly: false
  }
)

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

const activeNames = ref<string[]>([])
const isExpanded = computed(() => activeNames.value.includes('pengeluaran'))

const pengeluaranList = ref<KelasPengeluaran[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
let abortController: AbortController | null = null

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

interface PengeluaranFormState {
  detail_id?: number | string
  trx_id?: number | string
  tim_kerja?: string
  keterangan: string
  biaya: number
}

const form = ref<PengeluaranFormState>({
  tim_kerja: '',
  keterangan: '',
  biaya: 0
})

const formRules: FormRules = {
  tim_kerja: [
    { max: 3, message: 'Kode tim kerja maksimal 3 karakter', trigger: 'blur' }
  ],
  keterangan: [
    { max: 200, message: 'Keterangan maksimal 200 karakter', trigger: 'blur' }
  ],
  biaya: [
    { type: 'number', min: 0, message: 'Biaya harus berupa nilai positif atau 0', trigger: 'blur' }
  ]
}

function formatCurrency(val?: number) {
  if (val === undefined || val === null) return '0'
  return new Intl.NumberFormat('id-ID').format(val)
}

function handleAccordionChange(val: string[] | string) {
  const activeList = Array.isArray(val) ? val : [val]
  if (activeList.includes('pengeluaran')) {
    fetchPengeluaran()
  }
}

async function fetchPengeluaran() {
  if (!props.kelasId || !isExpanded.value) return

  if (abortController) abortController.abort()
  abortController = new AbortController()

  loading.value = true
  try {
    const res = await kelasApi.getKelasPengeluaran(
      props.kelasId,
      { page: currentPage.value, limit: pageSize.value },
      abortController.signal
    )
    pengeluaranList.value = res.data || []
    total.value = res.meta?.total ?? pengeluaranList.value.length
  } catch (err: any) {
    if (err.name === 'AbortError' || err.name === 'CanceledError') return
    console.error('Error fetching kelas pengeluaran:', err)
  } finally {
    loading.value = false
  }
}

function handleSizeChange(newSize: number) {
  pageSize.value = newSize
  currentPage.value = 1
  if (isExpanded.value) fetchPengeluaran()
}

function handleCurrentChange(newPage: number) {
  currentPage.value = newPage
  if (isExpanded.value) fetchPengeluaran()
}

function openAddDialog() {
  dialogMode.value = 'add'
  dialogFieldErrors.value = {}
  form.value = {
    trx_id: props.kelasId,
    tim_kerja: '',
    keterangan: '',
    biaya: 0
  }
  dialogVisible.value = true
}

function openEditDialog(row: KelasPengeluaran) {
  dialogMode.value = 'edit'
  dialogFieldErrors.value = {}
  const detailId = row.detailid || row.detail_id

  form.value = {
    detail_id: detailId,
    trx_id: props.kelasId,
    tim_kerja: row.tim_kerja || '',
    keterangan: row.keterangan || '',
    biaya: row.biaya || 0
  }
  dialogVisible.value = true
}

async function submitForm() {
  if (!formRef.value || submitting.value) return
  submitting.value = true
  try {
    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return
    dialogFieldErrors.value = {}

    const payload: Partial<KelasPengeluaran> = {
      trx_id: props.kelasId ? Number(props.kelasId) : undefined,
      tim_kerja: form.value.tim_kerja || undefined,
      keterangan: form.value.keterangan || undefined,
      biaya: form.value.biaya !== undefined ? form.value.biaya : undefined
    }

    if (dialogMode.value === 'add') {
      await kelasApi.createKelasPengeluaran(payload)
      ElMessage.success('Data pengeluaran berhasil ditambahkan')
    } else {
      const detailId = form.value.detail_id
      if (!detailId) return
      await kelasApi.updateKelasPengeluaran(detailId, payload)
      ElMessage.success('Data pengeluaran berhasil diperbarui')
    }

    dialogVisible.value = false
    if (isExpanded.value) fetchPengeluaran()
  } catch (err: any) {
    console.error('Error submitting pengeluaran form:', err)
    if (err.response?.data?.details) {
      dialogFieldErrors.value = err.response.data.details
      ElMessage.error(err.response.data.error || 'Invalid Inputs, Silahkan periksa kolom form')
    } else {
      ElMessage.error(err.response?.data?.error || err.message || 'Gagal menyimpan data pengeluaran')
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: KelasPengeluaran) {
  const detailId = row.detailid || row.detail_id
  if (!detailId) return

  try {
    await kelasApi.deleteKelasPengeluaran(detailId, props.kelasId)
    ElMessage.success('Data pengeluaran berhasil dihapus')
    if (isExpanded.value) fetchPengeluaran()
  } catch (err: any) {
    console.error('Error deleting pengeluaran:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal menghapus data pengeluaran')
  }
}

watch(
  () => props.kelasId,
  (newId) => {
    if (newId) {
      currentPage.value = 1
      if (isExpanded.value) {
        fetchPengeluaran()
      } else {
        pengeluaranList.value = []
        total.value = 0
      }
    }
  }
)

onMounted(() => {
  if (isExpanded.value) fetchPengeluaran()
})

onUnmounted(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.pengeluaran-accordion-container {
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
  margin-right: 0.75rem;
}

.font-semibold {
  font-weight: 600;
}

.font-mono {
  font-family: monospace;
}

.text-danger {
  color: var(--el-color-danger);
}

.pagination-container {
  display: flex;
  /* justify-content: flex-end;
  margin-top: 1rem; */
}
</style>
