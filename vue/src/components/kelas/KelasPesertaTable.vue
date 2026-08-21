<template>
  <div class="peserta-accordion-container">
    <el-collapse v-model="activeNames" class="custom-accordion" @change="handleAccordionChange">
      <el-collapse-item name="peserta">
        <template #title>
          <div class="accordion-header">
            <div class="header-title">
              <el-icon class="header-icon">
                <UserFilled />
              </el-icon>
              <span>{{ isDesktop ? 'Daftar Peserta Kelas' : 'Peserta' }}</span>
              <el-tag v-if="isExpanded && !isMobile" size="small" type="info" class="ml-2">{{ total }} Peserta</el-tag>
            </div>
            <div class="header-actions" @click.stop>
              <el-button type="primary" size="small" :icon="Plus" @click.stop="openAddDialog">
                {{ isMobile ? '' : 'Tambah Peserta' }}
              </el-button>
              <el-button :icon="Refresh" circle size="small" title="Refresh Peserta" @click.stop="fetchPeserta" />
            </div>
          </div>
        </template>

        <div class="accordion-content">
          <el-table v-loading="loading" :data="pesertaList" stripe border max-height="450" style="width: 100%"
            empty-text="Belum ada peserta yang terdaftar pada kelas ini">
            <el-table-column v-if="isDesktop" type="index" label="No." width="60" align="center" fixed="left" />

            <el-table-column prop="nama_indonesia" label="Nama Ciu Tao" min-width="160">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.nama_indonesia || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="nama_mandarin" label="Nama Lain" min-width="140">
              <template #default="{ row }">
                <span>{{ row.nama_mandarin || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="fotang_ciutao_desc" label="Fotang Ciu Tao" min-width="150" />

            <el-table-column prop="fotang_aktif_desc" label="Fotang Aktif" min-width="150" />

            <el-table-column prop="tanggal_ciu_tao_int" label="Tgl Ciu Tao" width="120" align="center" />

            <el-table-column prop="pengajak" label="Pengajak" min-width="140" />

            <el-table-column prop="penanggung" label="Penanggung" min-width="140" />

            <el-table-column label="Status Lulus" width="140" align="center">
              <template #default="{ row }">
                <el-tag :type="row.lulus ? 'success' : 'info'" size="small">
                  {{ row.lulus ? 'Lulus' : 'Belum' }}
                </el-tag>
                <div v-if="row.keterangan_lulus" class="lulus-note">
                  {{ row.keterangan_lulus }}
                </div>
              </template>
            </el-table-column>

            <el-table-column label="Status Ikrar" min-width="180" align="center">
              <template #default="{ row }">
                <div class="ikrar-tags">
                  <el-tag v-if="row.ikrar1" size="small" type="success">I-1</el-tag>
                  <el-tag v-if="row.ikrar2" size="small" type="success">I-2</el-tag>
                  <el-tag v-if="row.ikrar3" size="small" type="success">I-3</el-tag>
                  <el-tag v-if="row.ikrar4" size="small" type="success">I-4</el-tag>
                  <el-tag v-if="row.ikrar5" size="small" type="success">I-5</el-tag>
                  <el-tag v-if="row.ikrar6" size="small" type="success">I-6</el-tag>
                  <span v-if="!row.ikrar1 && !row.ikrar2 && !row.ikrar3 && !row.ikrar4 && !row.ikrar5 && !row.ikrar6"
                    class="no-ikrar">-</span>
                </div>
              </template>
            </el-table-column>

            <el-table-column label="Aksi" width="90" align="center" :fixed="!isDesktop ? false : 'right'">
              <template #default="{ row }">
                <el-button type="primary" circle size="small" :icon="Edit" @click="openEditDialog(row)" />
                <el-popconfirm title="Yakin ingin menghapus peserta ini?" confirm-button-text="Ya, Hapus"
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
              :layout="'total, ' + (isDesktop ? ', jumper' : '') + ', prev, pager, next' + (isDesktop ? ', jumper' : '')"
              :total="total" @size-change="handleSizeChange" @current-change="handleCurrentChange" />
          </div>
        </div>
      </el-collapse-item>
    </el-collapse>

    <!-- Dialog Popup Add / Edit Peserta -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'add' ? 'Tambah Peserta' : 'Edit Peserta'"
      :width="isMobile ? '90%' : '600px'" destroy-on-close>
      <el-alert v-if="Object.keys(dialogFieldErrors).length > 0" type="error" show-icon title="Invalid Inputs"
        description="Terdapat kesalahan pengisian form. Silahkan periksa pesan kesalahan berwarna merah di bawah."
        class="mb-4" />

      <el-form ref="formRef" :model="form" :rules="formRules" label-width="140px" size="default" v-loading="submitting">
        <el-form-item label="Peserta (Umat)" prop="id_peserta" :error="hasFieldError('id_peserta') ? ' ' : undefined">
          <el-select v-model="form.id_peserta" filterable remote reserve-keyword
            placeholder="Ketik nama untuk mencari Umat..." :remote-method="searchUmat" :loading="loadingUmat"
            style="width: 100%">
            <el-option v-for="item in umatOptions" :key="item.id" :label="getUmatOptionLabel(item)" :value="item.id" />
          </el-select>
          <FieldErrors :errors="getFieldErrors('id_peserta')" />
        </el-form-item>

        <el-form-item v-if="dialogMode !== 'add'" label="Status Lulus"
          :error="hasFieldError('lulus') ? ' ' : undefined">
          <el-switch v-model="form.lulus" active-text="Lulus" inactive-text="Belum Lulus" />
          <FieldErrors :errors="getFieldErrors('lulus')" />
        </el-form-item>

        <el-form-item v-if="dialogMode !== 'add'" label="Keterangan Lulus"
          :error="hasFieldError('keterangan_lulus') ? ' ' : undefined">
          <el-input v-model="form.keterangan_lulus" placeholder="Catatan / keterangan kelulusan" maxlength="200" />
          <FieldErrors :errors="getFieldErrors('keterangan_lulus')" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :md="12" :sm="24">
            <el-form-item label="Keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
              <el-input v-model="form.keterangan" placeholder="Keterangan tambahan" maxlength="200" />
              <FieldErrors :errors="getFieldErrors('keterangan')" />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- Logistik Optional Fields (Multiply by Number of Days) -->
        <el-divider content-position="left">Data Logistik / Kehadiran ({{ daysCount }} Hari)</el-divider>
        <div class="logistik-table-wrapper mb-4">
          <el-table :data="logistikRows" border size="small" style="width: 100%">
            <el-table-column label="Hari" width="120" align="center">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.label }}</span>
              </template>
            </el-table-column>

            <el-table-column label="Anak" width="100" align="center">
              <template #default="{ $index }">
                <el-input-number v-model="anakDays[$index]" :min="0" :max="99" controls-position="right" size="small"
                  style="width: 100%" />
              </template>
            </el-table-column>

            <el-table-column label="Suster" width="100" align="center">
              <template #default="{ $index }">
                <el-input-number v-model="susterDays[$index]" :min="0" :max="99" controls-position="right" size="small"
                  style="width: 100%" />
              </template>
            </el-table-column>

            <el-table-column label="Menginap" align="center">
              <template #default="{ $index }">
                <el-checkbox v-model="menginapDays[$index]" />
              </template>
            </el-table-column>

            <el-table-column label="Makan Pagi" align="center">
              <template #default="{ $index }">
                <el-checkbox v-model="makananPagiDays[$index]" />
              </template>
            </el-table-column>

            <el-table-column label="Makan Siang" align="center">
              <template #default="{ $index }">
                <el-checkbox v-model="makananSiangDays[$index]" />
              </template>
            </el-table-column>

            <el-table-column label="Makan Malam" align="center">
              <template #default="{ $index }">
                <el-checkbox v-model="makananMalamDays[$index]" />
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
import { UserFilled, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import dayjs from 'dayjs'
import { kelasApi } from '../../api/kelas'
import { umatApi } from '../../api/umat'
import FieldErrors from '../common/FieldErrors.vue'
import type { KelasPeserta } from '../../types/kelas'
import type { Umat } from '../../types/umat'
import { scrollToFormError } from '../../utils/scroll'

const props = defineProps<{
  kelasId: string | number
  startDate?: string
  endDate?: string
}>()


// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

// Accordion collapse state (empty array by default -> collapsed)
const activeNames = ref<string[]>([])
const isExpanded = computed(() => activeNames.value.includes('peserta'))

const pesertaList = ref<KelasPeserta[]>([])
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
    console.error('Error fetching kelas detail for logistik days:', err)
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

const anakDays = ref<number[]>([])
const susterDays = ref<number[]>([])
const menginapDays = ref<boolean[]>([])
const makananPagiDays = ref<boolean[]>([])
const makananSiangDays = ref<boolean[]>([])
const makananMalamDays = ref<boolean[]>([])

function parseIntArray(strVal: string | undefined, count: number): number[] {
  const result: number[] = new Array(count).fill(0)
  if (!strVal) return result
  const parts = strVal.split(',')
  for (let i = 0; i < count; i++) {
    if (i < parts.length && parts[i].trim() !== '') {
      const parsed = parseInt(parts[i].trim(), 10)
      result[i] = isNaN(parsed) ? 0 : parsed
    }
  }
  return result
}

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

function resizeIntArray(arr: number[], count: number): number[] {
  const newArr = new Array(count).fill(0)
  for (let i = 0; i < count; i++) {
    if (i < arr.length) newArr[i] = arr[i]
  }
  return newArr
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
    anakDays.value = resizeIntArray(anakDays.value, newCount)
    susterDays.value = resizeIntArray(susterDays.value, newCount)
    menginapDays.value = resizeBoolArray(menginapDays.value, newCount)
    makananPagiDays.value = resizeBoolArray(makananPagiDays.value, newCount)
    makananSiangDays.value = resizeBoolArray(makananSiangDays.value, newCount)
    makananMalamDays.value = resizeBoolArray(makananMalamDays.value, newCount)
  }
})

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

interface PesertaFormState {
  detail_id?: number | string
  trx_id?: number | string
  id_peserta?: number | string
  lulus: boolean
  keterangan_lulus: string
  sumbangan?: number
  barang: string
  tim_kerja: string
  keterangan: string
}

const form = ref<PesertaFormState>({
  lulus: false,
  keterangan_lulus: '',
  sumbangan: 0,
  barang: '',
  tim_kerja: '',
  keterangan: ''
})

const formRules: FormRules = {
  id_peserta: [
    { required: true, message: 'Peserta (Umat) wajib dipilih', trigger: 'change' }
  ]
}

// Search Umat dropdown
const umatOptions = ref<Umat[]>([])
const loadingUmat = ref(false)

async function searchUmat(query: string) {
  if (!query || query.trim().length < 1) {
    umatOptions.value = []
    return
  }
  loadingUmat.value = true
  try {
    const res = await umatApi.getUmats({ page: 1, limit: 20, namaindonesia: query })
    umatOptions.value = res.data || []
  } catch (err) {
    console.error('Failed searching umat:', err)
  } finally {
    loadingUmat.value = false
  }
}

function getUmatOptionLabel(item: Umat): string {
  const parts = []
  if (item.nama_indonesia) parts.push(item.nama_indonesia)
  if (item.alias) parts.push(item.alias)
  if (item.nama_mandarin) parts.push(item.nama_mandarin)
  const names = parts.join(' / ')
  return names ? `${names} (${item.kode || item.id || ''})` : String(item.kode || item.id || '')
}

// Accordion change listener: fetch data ONLY when expanded
function handleAccordionChange(val: string[] | string) {
  const activeList = Array.isArray(val) ? val : [val]
  if (activeList.includes('peserta')) {
    fetchPeserta()
  }
}

async function fetchPeserta() {
  if (!props.kelasId) return
  if (!isExpanded.value) return

  if (abortController) abortController.abort()
  abortController = new AbortController()

  loading.value = true
  try {
    const res = await kelasApi.getKelasPeserta(
      props.kelasId,
      { page: currentPage.value, limit: pageSize.value },
      abortController.signal
    )
    pesertaList.value = res.data || []
    total.value = res.meta?.total ?? pesertaList.value.length
  } catch (err: any) {
    if (err.name === 'AbortError' || err.name === 'CanceledError') return
    console.error('Error fetching kelas peserta:', err)
  } finally {
    loading.value = false
  }
}

function handleSizeChange(newSize: number) {
  pageSize.value = newSize
  currentPage.value = 1
  if (isExpanded.value) fetchPeserta()
}

function handleCurrentChange(newPage: number) {
  currentPage.value = newPage
  if (isExpanded.value) fetchPeserta()
}

// Open Dialog in Add Mode
function openAddDialog() {
  dialogMode.value = 'add'
  dialogFieldErrors.value = {}
  form.value = {
    lulus: false,
    keterangan_lulus: '',
    sumbangan: 0,
    barang: '',
    tim_kerja: '',
    keterangan: ''
  }

  const count = daysCount.value
  anakDays.value = new Array(count).fill(0)
  susterDays.value = new Array(count).fill(0)
  menginapDays.value = new Array(count).fill(false)
  makananPagiDays.value = new Array(count).fill(false)
  makananSiangDays.value = new Array(count).fill(false)
  makananMalamDays.value = new Array(count).fill(false)

  umatOptions.value = []
  dialogVisible.value = true
}

// Open Dialog in Edit Mode
function openEditDialog(row: KelasPeserta) {
  dialogMode.value = 'edit'
  dialogFieldErrors.value = {}
  const rawIdPeserta = row.id_peserta ? Number(row.id_peserta) : undefined

  form.value = {
    detail_id: row.detail_id || row.detailid,
    trx_id: row.trx_id,
    id_peserta: rawIdPeserta,
    lulus: Boolean(row.lulus),
    keterangan_lulus: row.keterangan_lulus || '',
    sumbangan: row.sumbangan || 0,
    barang: row.barang || '',
    tim_kerja: row.tim_kerja || '',
    keterangan: row.keterangan || ''
  }

  const count = daysCount.value
  anakDays.value = parseIntArray(row.anak, count)
  susterDays.value = parseIntArray(row.suster, count)
  menginapDays.value = parseBoolArray(row.menginap, count)
  makananPagiDays.value = parseBoolArray(row.makanan_pagi, count)
  makananSiangDays.value = parseBoolArray(row.makanan_siang, count)
  makananMalamDays.value = parseBoolArray(row.makanan_malam, count)

  if (rawIdPeserta && row.nama_indonesia) {
    umatOptions.value = [
      {
        id: rawIdPeserta,
        nama_indonesia: row.nama_indonesia,
        nama_mandarin: row.nama_mandarin
      } as Umat
    ]
  }
  dialogVisible.value = true
}

// Submit Create or Edit Form
async function submitForm() {
  if (!formRef.value || submitting.value) return
  submitting.value = true
  try {
    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) {
      scrollToFormError()
      return
    }
    dialogFieldErrors.value = {}

    const payload: Partial<KelasPeserta> = {
      trx_id: props.kelasId ? Number(props.kelasId) : undefined,
      id_peserta: form.value.id_peserta ? Number(form.value.id_peserta) : undefined,
      lulus: form.value.lulus,
      keterangan_lulus: form.value.keterangan_lulus,
      sumbangan: form.value.sumbangan,
      barang: form.value.barang,
      tim_kerja: form.value.tim_kerja,
      keterangan: form.value.keterangan,
      anak: anakDays.value.map(v => v || 0).join(','),
      suster: susterDays.value.map(v => v || 0).join(','),
      menginap: menginapDays.value.map(v => v ? '1' : '0').join(','),
      makanan_pagi: makananPagiDays.value.map(v => v ? '1' : '0').join(','),
      makanan_siang: makananSiangDays.value.map(v => v ? '1' : '0').join(','),
      makanan_malam: makananMalamDays.value.map(v => v ? '1' : '0').join(',')
    }

    if (dialogMode.value === 'add') {
      await kelasApi.createKelasPeserta(payload)
      ElMessage.success('Peserta kelas berhasil ditambahkan')
    } else {
      const detailId = form.value.detail_id
      if (!detailId) return
      await kelasApi.updateKelasPeserta(detailId, payload)
      ElMessage.success('Data peserta kelas berhasil diperbarui')
    }

    dialogVisible.value = false
    if (isExpanded.value) fetchPeserta()
  } catch (err: any) {
    console.error('Failed to save peserta:', err)
    if (err.response?.data?.details) {
      dialogFieldErrors.value = err.response.data.details
      ElMessage.error(err.response.data.error || 'Invalid Inputs, Silahkan periksa kolom form')
    } else {
      ElMessage.error(err.response?.data?.error || err.message || 'Gagal menyimpan data peserta')
    }
    scrollToFormError()
  } finally {
    submitting.value = false
  }
}

// Handle Delete Participant
async function handleDelete(row: KelasPeserta) {
  const detailId = row.detailid || row.detail_id
  if (!detailId) return

  try {
    await kelasApi.deleteKelasPeserta(detailId, props.kelasId)
    ElMessage.success('Peserta kelas berhasil dihapus')
    if (isExpanded.value) fetchPeserta()
  } catch (err: any) {
    console.error('Error deleting peserta:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal menghapus data peserta')
  }
}

watch(
  () => props.kelasId,
  (newId) => {
    if (newId) {
      if (!props.startDate || !props.endDate) {
        fetchKelasDetail()
      }
      currentPage.value = 1
      if (isExpanded.value) {
        fetchPeserta()
      } else {
        pesertaList.value = []
        total.value = 0
      }
    }
  }
)

onMounted(() => {
  if (!props.startDate || !props.endDate) {
    fetchKelasDetail()
  }
  // Only fetch if accordion is expanded on mount
  if (isExpanded.value) {
    fetchPeserta()
  }
})

onUnmounted(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.peserta-accordion-container {
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

.font-mono {
  font-family: monospace;
}

.font-semibold {
  font-weight: 600;
}

.lulus-note {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.ikrar-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  justify-content: center;
}

.no-ikrar {
  color: var(--el-text-color-placeholder);
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}
</style>
