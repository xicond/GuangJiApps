<template>
  <div class="kendaraan-accordion-container">
    <el-collapse v-model="activeNames" class="custom-accordion" @change="handleAccordionChange">
      <el-collapse-item name="kendaraan">
        <template #title>
          <div class="accordion-header">
            <div class="header-title">
              <el-icon class="header-icon">
                <Van />
              </el-icon>
              <span>{{ isDesktop ? 'Daftar Kendaraan Kelas' : 'Kendaraan' }}</span>
              <el-tag v-if="isExpanded && !isMobile" size="small" type="info" class="ml-2">{{ total }}
                Kendaraan</el-tag>
            </div>
            <div class="header-actions" @click.stop>
              <el-button v-if="!readonly" type="primary" size="small" :icon="Plus" @click.stop="openAddDialog">
                {{ isMobile ? '' : 'Tambah Kendaraan' }}
              </el-button>
              <el-button :icon="Refresh" circle size="small" title="Refresh Kendaraan" @click.stop="fetchKendaraan" />
            </div>
          </div>
        </template>

        <div class="accordion-content">
          <el-table v-loading="loading" :data="kendaraanList" stripe border max-height="450" style="width: 100%"
            empty-text="Belum ada kendaraan yang terdaftar pada kelas ini">
            <el-table-column v-if="isDesktop" type="index" label="No." width="60" align="center" fixed="left" />

            <el-table-column prop="no_polisi" label="No. Polisi" min-width="130">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.no_polisi || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="pengendara" label="Pengendara" min-width="160">
              <template #default="{ row }">
                <span>{{ row.pengendara || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="tipe_kendaraan" label="Tipe Kendaraan" min-width="140">
              <template #default="{ row }">
                <span>{{ row.tipe_kendaraan || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="fotang" label="Fotang" min-width="120" align="center">
              <template #default="{ row }">
                <span>{{ row.fotang || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="hari" label="Hari" min-width="120" align="center">
              <template #default="{ row }">
                <span>{{ row.hari || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="keterangan" label="Keterangan" min-width="180">
              <template #default="{ row }">
                <span>{{ row.keterangan || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column v-if="!readonly" label="Aksi" width="90" align="center"
              :fixed="!isDesktop ? false : 'right'">
              <template #default="{ row }">
                <el-button type="primary" circle size="small" :icon="Edit" @click="openEditDialog(row)" />
                <el-popconfirm title="Yakin ingin menghapus kendaraan ini?" confirm-button-text="Ya, Hapus"
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

    <!-- Dialog Popup Add / Edit Kendaraan -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'add' ? 'Tambah Kendaraan' : 'Edit Kendaraan'"
      :width="isMobile ? '90%' : '600px'" destroy-on-close>
      <el-alert v-if="Object.keys(dialogFieldErrors).length > 0" type="error" show-icon title="Invalid Inputs"
        description="Terdapat kesalahan pengisian form. Silahkan periksa pesan kesalahan berwarna merah di bawah."
        class="mb-4" />

      <el-form ref="formRef" :model="form" :rules="formRules" label-width="140px" size="default" v-loading="submitting">
        <el-row :gutter="16">
          <el-col :md="12" :sm="24">
            <el-form-item label="No. Polisi" prop="no_polisi" :error="hasFieldError('no_polisi') ? ' ' : undefined">
              <el-input v-model="form.no_polisi" placeholder="Contoh: B 1234 CD" maxlength="15" />
              <FieldErrors :errors="getFieldErrors('no_polisi')" />
            </el-form-item>
          </el-col>
          <el-col :md="12" :sm="24">
            <el-form-item label="Tipe Kendaraan" prop="tipe_kendaraan"
              :error="hasFieldError('tipe_kendaraan') ? ' ' : undefined">
              <el-input v-model="form.tipe_kendaraan" placeholder="Contoh: Mobil, Motor, Bus" maxlength="50" />
              <FieldErrors :errors="getFieldErrors('tipe_kendaraan')" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="Pengendara" prop="pengendara" :error="hasFieldError('pengendara') ? ' ' : undefined">
          <el-input v-model="form.pengendara" placeholder="Nama pengendara" maxlength="100" />
          <FieldErrors :errors="getFieldErrors('pengendara')" />
        </el-form-item>

        <!-- <el-form-item label="Fotang" prop="fotang" :error="hasFieldError('fotang') ? ' ' : undefined">
          <LookupSelect v-model="form.fotang" :fetch-api="lookupApi.getLookupFotang" value-key="lookup_value"
            placeholder="Pilih Fotang..." :clearable="true" />
          <FieldErrors :errors="getFieldErrors('fotang')" />
        </el-form-item> -->

        <el-form-item label="Keterangan" prop="keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
          <el-input v-model="form.keterangan" type="textarea" :rows="2" placeholder="Keterangan kendaraan (opsional)"
            maxlength="200" show-word-limit />
          <FieldErrors :errors="getFieldErrors('keterangan')" />
        </el-form-item>

        <!-- Jadwal / Kehadiran Operasional (Multiply by Number of Days) -->
        <el-divider content-position="left">Jadwal Kehadiran ({{ daysCount }} Hari)</el-divider>
        <div class="hari-table-wrapper mb-4">
          <el-table :data="logistikRows" border size="small" style="width: 100%">
            <el-table-column label="Hari" min-width="140" align="center">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.label }}</span>
              </template>
            </el-table-column>

            <el-table-column label="Operasional" align="center">
              <template #default="{ $index }">
                <el-checkbox v-model="hariDays[$index]" />
              </template>
            </el-table-column>
          </el-table>
        </div>
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
import { Van, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import dayjs from 'dayjs'
import { kelasApi } from '../../api/kelas'
import { lookupApi } from '../../api/lookup'
import FieldErrors from '../common/FieldErrors.vue'
import LookupSelect from '../common/LookupSelect.vue'
import type { KelasKendaraan } from '../../types/kelas'

const props = withDefaults(
  defineProps<{
    kelasId: string | number
    startDate?: string
    endDate?: string
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
const isExpanded = computed(() => activeNames.value.includes('kendaraan'))

const kendaraanList = ref<KelasKendaraan[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
let abortController: AbortController | null = null

const kelasDetail = ref<{ start_date?: string; end_date?: string } | null>(null)

async function fetchKelasDetail() {
  if (!props.kelasId) return
  if (props.startDate && props.endDate) return
  if (kelasDetail.value) return
  try {
    const res = await kelasApi.getKelasById(props.kelasId)
    if (res.data) {
      kelasDetail.value = res.data
    }
  } catch (err) {
    console.error('Error fetching kelas detail for kendaraan days:', err)
  }
}

const daysCount = computed(() => {
  const start = props.startDate || kelasDetail.value?.start_date
  const end = props.endDate || kelasDetail.value?.end_date
  if (!start || !end) return 1
  const dStart = dayjs(start)
  const dEnd = dayjs(end)
  if (!dStart.isValid() || !dEnd.isValid()) return 1
  const diff = dEnd.diff(dStart, 'day') + 1
  return diff > 0 ? diff : 1
})

const logistikRows = computed(() => {
  const count = daysCount.value
  const start = props.startDate || kelasDetail.value?.start_date
  const rows = []
  for (let i = 0; i < count; i++) {
    let label = `Hari ${i + 1}`
    if (start && dayjs(start).isValid()) {
      label += ` (${dayjs(start).add(i, 'day').format('DD/MM')})`
    }
    rows.push({ index: i, label })
  }
  return rows
})

const hariDays = ref<boolean[]>([])

function parseBoolArray(strVal: string | undefined, count: number): boolean[] {
  const result: boolean[] = new Array(count).fill(false)
  if (!strVal) return result
  const parts = strVal.split(',')
  for (let i = 0; i < count; i++) {
    if (i < parts.length && parts[i].trim() !== '') {
      const val = parts[i].trim().toLowerCase()
      result[i] = ['1', 'y', 'true', 'ya'].includes(val)
    }
  }
  return result
}

function resizeBoolArray(arr: boolean[], count: number): boolean[] {
  const newArr = new Array(count).fill(false)
  for (let i = 0; i < count; i++) {
    if (i < arr.length) newArr[i] = arr[i]
  }
  return newArr
}

watch(daysCount, (newCount) => {
  if (dialogVisible.value) {
    hariDays.value = resizeBoolArray(hariDays.value, newCount)
  }
})

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

interface KendaraanFormState {
  detail_id?: number | string
  trx_id?: number | string
  no_polisi: string
  pengendara: string
  tipe_kendaraan: string
  fotang?: string
  keterangan: string
}

const form = ref<KendaraanFormState>({
  no_polisi: '',
  pengendara: '',
  tipe_kendaraan: '',
  fotang: '',
  keterangan: ''
})

const formRules: FormRules = {
  no_polisi: [
    { required: true, message: 'Harap isi nomor polisi', trigger: 'blur' },
    { max: 15, message: 'Nomor polisi maksimal 15 karakter', trigger: 'blur' }
  ],
  pengendara: [
    { required: true, message: 'Harap isi nama pengendara', trigger: 'blur' },
    { max: 100, message: 'Nama pengendara maksimal 100 karakter', trigger: 'blur' }
  ],
  tipe_kendaraan: [
    { max: 50, message: 'Tipe kendaraan maksimal 50 karakter', trigger: 'blur' }
  ],
  keterangan: [
    { max: 200, message: 'Keterangan maksimal 200 karakter', trigger: 'blur' }
  ]
}

function handleAccordionChange(val: string[] | string) {
  const activeList = Array.isArray(val) ? val : [val]
  if (activeList.includes('kendaraan')) {
    fetchKendaraan()
  }
}

async function fetchKendaraan() {
  if (!props.kelasId || !isExpanded.value) return

  if (abortController) abortController.abort()
  abortController = new AbortController()

  loading.value = true
  try {
    const res = await kelasApi.getKelasKendaraan(
      props.kelasId,
      { page: currentPage.value, limit: pageSize.value },
      abortController.signal
    )
    kendaraanList.value = res.data || []
    total.value = res.meta?.total ?? kendaraanList.value.length
  } catch (err: any) {
    if (err.name === 'AbortError' || err.name === 'CanceledError') return
    console.error('Error fetching kelas kendaraan:', err)
  } finally {
    loading.value = false
  }
}

function handleSizeChange(newSize: number) {
  pageSize.value = newSize
  currentPage.value = 1
  if (isExpanded.value) fetchKendaraan()
}

function handleCurrentChange(newPage: number) {
  currentPage.value = newPage
  if (isExpanded.value) fetchKendaraan()
}

async function openAddDialog() {
  await fetchKelasDetail()
  dialogMode.value = 'add'
  dialogFieldErrors.value = {}
  form.value = {
    trx_id: props.kelasId,
    no_polisi: '',
    pengendara: '',
    tipe_kendaraan: '',
    fotang: '',
    keterangan: ''
  }
  const count = daysCount.value
  hariDays.value = new Array(count).fill(false)
  dialogVisible.value = true
}

async function openEditDialog(row: KelasKendaraan) {
  await fetchKelasDetail()
  dialogMode.value = 'edit'
  dialogFieldErrors.value = {}
  const detailId = row.detailid || row.detail_id

  form.value = {
    detail_id: detailId,
    trx_id: props.kelasId,
    no_polisi: row.no_polisi || '',
    pengendara: row.pengendara || '',
    tipe_kendaraan: row.tipe_kendaraan || '',
    fotang: row.fotang || '',
    keterangan: row.keterangan || ''
  }

  const count = daysCount.value
  hariDays.value = parseBoolArray(row.hari, count)
  dialogVisible.value = true
}

async function submitForm() {
  if (!formRef.value || submitting.value) return
  submitting.value = true
  try {
    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return
    dialogFieldErrors.value = {}

    const payload: Partial<KelasKendaraan> = {
      trx_id: props.kelasId ? Number(props.kelasId) : undefined,
      no_polisi: form.value.no_polisi || undefined,
      pengendara: form.value.pengendara || undefined,
      tipe_kendaraan: form.value.tipe_kendaraan || undefined,
      fotang: form.value.fotang || undefined,
      hari: hariDays.value.map(v => v ? '1' : '0').join(','),
      keterangan: form.value.keterangan || undefined
    }

    if (dialogMode.value === 'add') {
      await kelasApi.createKelasKendaraan(payload)
      ElMessage.success('Data kendaraan berhasil ditambahkan')
    } else {
      const detailId = form.value.detail_id
      if (!detailId) return
      await kelasApi.updateKelasKendaraan(detailId, payload)
      ElMessage.success('Data kendaraan berhasil diperbarui')
    }

    dialogVisible.value = false
    if (isExpanded.value) fetchKendaraan()
  } catch (err: any) {
    console.error('Error submitting kendaraan form:', err)
    if (err.response?.data?.details) {
      dialogFieldErrors.value = err.response.data.details
      ElMessage.error(err.response.data.error || 'Invalid Inputs, Silahkan periksa kolom form')
    } else {
      ElMessage.error(err.response?.data?.error || err.message || 'Gagal menyimpan data kendaraan')
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: KelasKendaraan) {
  const detailId = row.detailid || row.detail_id
  if (!detailId) return

  try {
    await kelasApi.deleteKelasKendaraan(detailId, props.kelasId)
    ElMessage.success('Data kendaraan berhasil dihapus')
    if (isExpanded.value) fetchKendaraan()
  } catch (err: any) {
    console.error('Error deleting kendaraan:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal menghapus data kendaraan')
  }
}

watch(
  () => props.kelasId,
  (newId) => {
    if (newId) {
      currentPage.value = 1
      if (isExpanded.value) {
        fetchKendaraan()
      } else {
        kendaraanList.value = []
        total.value = 0
      }
    }
  }
)

onMounted(() => {
  if (isExpanded.value) fetchKendaraan()
})

onUnmounted(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.kendaraan-accordion-container {
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

/* .pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
} */

.hari-table-wrapper {
  margin-top: 0.5rem;
}
</style>
