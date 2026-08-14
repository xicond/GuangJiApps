<template>
  <div class="pengabdi-accordion-container">
    <el-collapse v-model="activeNames" class="custom-accordion" @change="handleAccordionChange">
      <el-collapse-item name="pengabdi">
        <template #title>
          <div class="accordion-header" @click.stop>
            <div class="header-title">
              <el-icon class="header-icon">
                <Avatar />
              </el-icon>
              <span>{{ isDesktop ? 'Daftar Pengabdi' : 'Pengabdi' }}</span>
              <el-tag size="small" type="info" class="ml-2">{{ total }} Pengabdi</el-tag>
            </div>
            <div class="header-actions" @click.stop>
              <el-button type="primary" size="small" :icon="Plus" @click.stop="openAddDialog">
                {{ isMobile ? '' : 'Tambah Pengabdi' }}
              </el-button>
              <el-button :icon="Refresh" circle size="small" title="Refresh Pengabdi" @click.stop="fetchPengabdi" />
            </div>
          </div>
        </template>

        <div class="accordion-content">
          <el-table v-loading="loading" :data="pengabdiList" stripe border max-height="450" style="width: 100%"
            empty-text="Belum ada pengabdi yang terdaftar pada kelas ini">
            <el-table-column v-if="isDesktop" type="index" label="No." width="60" align="center" fixed="left" />

            <el-table-column prop="nama_indonesia" label="Nama Ciu Tao" min-width="160">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.nama_indonesia || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="nama_mandarin" label="Nama Lain" min-width="160">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.nama_mandarin || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="hari" label="Hari" width="120" align="center" />

            <el-table-column prop="tim_kerja" label="Tim Kerja" min-width="130" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.tim_kerja_desc || row.tim_kerja" size="small" type="primary">{{ row.tim_kerja_desc ||
                  row.tim_kerja }}</el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>

            <el-table-column prop="sub_kerja" label="Sub Kerja" min-width="130" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.sub_kerja_desc || row.sub_kerja" size="small" type="info">{{ row.sub_kerja_desc ||
                  row.sub_kerja }}</el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>

            <el-table-column prop="tim_kerja_report" label="Tim Kerja Report" width="140" align="center" />

            <el-table-column prop="keterangan" label="Keterangan" min-width="160" />

            <el-table-column label="Aksi" width="90" align="center" :fixed="!isDesktop ? false : 'right'">
              <template #default="{ row }">
                <el-button type="primary" circle size="small" :icon="Edit" @click="openEditDialog(row)" />
                <el-popconfirm title="Yakin ingin menghapus pengabdi ini?" confirm-button-text="Ya, Hapus"
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

    <!-- Dialog Popup Add / Edit Pengabdi -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'add' ? 'Tambah Pengabdi' : 'Edit Pengabdi'"
      :width="isMobile ? '90%' : '600px'" destroy-on-close>
      <el-alert v-if="Object.keys(dialogFieldErrors).length > 0" type="error" show-icon title="Invalid Inputs"
        description="Terdapat kesalahan pengisian form. Silahkan periksa pesan kesalahan berwarna merah di bawah."
        class="mb-4" />

      <el-form ref="formRef" :model="form" :rules="formRules" label-width="140px" size="default" v-loading="submitting">
        <el-form-item label="Pengabdi (Umat)" prop="id_pengabdi"
          :error="hasFieldError('id_pengabdi') ? ' ' : undefined">
          <el-select v-model="form.id_pengabdi" filterable remote reserve-keyword
            placeholder="Ketik nama untuk mencari Umat..." :remote-method="searchUmat" :loading="loadingUmat"
            style="width: 100%">
            <el-option v-for="item in umatOptions" :key="item.id" :label="getUmatOptionLabel(item)" :value="item.id" />
          </el-select>
          <FieldErrors :errors="getFieldErrors('id_pengabdi')" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Tim Kerja" :error="hasFieldError('tim_kerja') ? ' ' : undefined">
              <LookupSelect v-model="form.tim_kerja" :fetch-api="lookupApi.getLookupTimKerja" value-key="lookup_value"
                :initial-option="initialTimKerjaOption" placeholder="Pilih Tim Kerja..." :clearable="false"
                @change="onTimKerjaChange" />
              <FieldErrors :errors="getFieldErrors('tim_kerja')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Sub Kerja" :error="hasFieldError('sub_kerja') ? ' ' : undefined">
              <LookupSelect ref="subKerjaSelectRef" :key="String(form.tim_kerja)" v-model="form.sub_kerja"
                :fetch-api="fetchSubKerjaApi" :disabled="!form.tim_kerja" value-key="lookup_value"
                :initial-option="initialSubKerjaOption" placeholder="Pilih Sub Kerja..." />
              <FieldErrors :errors="getFieldErrors('sub_kerja')" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="Keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
          <el-input v-model="form.keterangan" placeholder="Keterangan pengabdi" maxlength="200" />
          <FieldErrors :errors="getFieldErrors('keterangan')" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Sumbangan" :error="hasFieldError('sumbangan') ? ' ' : undefined">
              <el-input-number v-model="form.sumbangan" :min="100" :precision="0" :step="1000" controls-position="right"
                style="width: 100%" />
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

        <!-- Logistik Optional Fields (Multiply by Number of Days) -->
        <el-divider content-position="left">Data Logistik / Kehadiran ({{ daysCount }} Hari)</el-divider>
        <div class="logistik-table-wrapper mb-4">
          <el-table :data="logistikRows" border size="small" style="width: 100%">
            <el-table-column label="Hari" width="120" align="center">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.label }}</span>
              </template>
            </el-table-column>

            <el-table-column label="Tugas" align="center">
              <template #default="{ $index }">
                <el-checkbox v-model="hariDays[$index]" />
              </template>
            </el-table-column>

            <el-table-column label="Anak" width="90" align="center">
              <template #default="{ $index }">
                <el-input-number v-model="anakDays[$index]" :min="0" :max="99" controls-position="right" size="small"
                  style="width: 100%" />
              </template>
            </el-table-column>

            <el-table-column label="Suster" width="90" align="center">
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
import { Avatar, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import dayjs from 'dayjs'
import { kelasApi } from '../../api/kelas'
import { umatApi } from '../../api/umat'
import lookupApi from '../../api/lookup'
import LookupSelect from '../common/LookupSelect.vue'
import FieldErrors from '../common/FieldErrors.vue'
import type { KelasPengabdi } from '../../types/kelas'
import type { Umat } from '../../types/umat'
import type { LookupQueryParams } from '../../types/lookup'

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

const activeNames = ref<string[]>([])
const isExpanded = computed(() => activeNames.value.includes('pengabdi'))

const pengabdiList = ref<KelasPengabdi[]>([])
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

const hariDays = ref<boolean[]>([])
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
    hariDays.value = resizeBoolArray(hariDays.value, newCount)
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

interface PengabdiFormState {
  detail_id?: number | string
  trx_id?: number | string
  id_pengabdi?: number | string
  tim_kerja: string | number
  tim_kerja_report: string
  hari: string
  sub_kerja: string
  sumbangan?: number
  barang: string
  keterangan: string
}

const form = ref<PengabdiFormState>({
  tim_kerja: '',
  tim_kerja_report: '',
  hari: '',
  sub_kerja: '',
  barang: '',
  keterangan: ''
})

const formRules: FormRules = {
  id_pengabdi: [{ required: true, message: 'Harap pilih pengabdi / umat', trigger: 'change' }]
}

const subKerjaSelectRef = ref<InstanceType<typeof LookupSelect> | null>(null)

function fetchSubKerjaApi(params: LookupQueryParams = {}, signal?: AbortSignal) {
  if (!form.value.tim_kerja) {
    return Promise.resolve({ data: [], meta: { page: 1, limit: 10, total: 0 } })
  }
  return lookupApi.getLookupSubKerja(form.value.tim_kerja, params, signal)
}

function onTimKerjaChange() {
  form.value.sub_kerja = ''
  if (subKerjaSelectRef.value) {
    subKerjaSelectRef.value.loadData(1, '')
  }
}

watch(
  () => form.value.tim_kerja,
  (newVal, oldVal) => {
    if (newVal !== oldVal && oldVal !== undefined) {
      form.value.sub_kerja = ''
      if (subKerjaSelectRef.value) {
        subKerjaSelectRef.value.loadData(1, '')
      }
    }
  }
)

const initialTimKerjaOption = computed(() => {
  if (!form.value.tim_kerja) return undefined
  return {
    lookup_value: String(form.value.tim_kerja),
    lookup_description: String(form.value.tim_kerja)
  }
})

const initialSubKerjaOption = computed(() => {
  if (!form.value.sub_kerja) return undefined
  return {
    // lookup_id: Number(form.value.sub_kerja),
    lookup_value: String(form.value.sub_kerja),
    // lookup_description: String(form.value.sub_kerja)
  }
})

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
    keterangan: ''
  }

  const count = daysCount.value
  hariDays.value = new Array(count).fill(false)
  anakDays.value = new Array(count).fill(0)
  susterDays.value = new Array(count).fill(0)
  menginapDays.value = new Array(count).fill(false)
  makananPagiDays.value = new Array(count).fill(false)
  makananSiangDays.value = new Array(count).fill(false)
  makananMalamDays.value = new Array(count).fill(false)

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
    keterangan: row.keterangan || ''
  }

  const count = daysCount.value
  hariDays.value = parseBoolArray(row.hari, count)
  anakDays.value = parseIntArray(row.anak, count)
  susterDays.value = parseIntArray(row.suster, count)
  menginapDays.value = parseBoolArray(row.menginap, count)
  makananPagiDays.value = parseBoolArray(row.makanan_pagi, count)
  makananSiangDays.value = parseBoolArray(row.makanan_siang, count)
  makananMalamDays.value = parseBoolArray(row.makanan_malam, count)

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
  if (!formRef.value || submitting.value) return
  submitting.value = true
  try {
    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return
    dialogFieldErrors.value = {}

    const payload: Partial<KelasPengabdi> = {
      trx_id: props.kelasId ? Number(props.kelasId) : undefined,
      id_pengabdi: form.value.id_pengabdi ? Number(form.value.id_pengabdi) : undefined,
      tim_kerja: String(form.value.tim_kerja || ''),
      tim_kerja_report: form.value.tim_kerja_report,
      hari: hariDays.value.map(v => v ? '1' : '0').join(','),
      sub_kerja: form.value.sub_kerja,
      sumbangan: form.value.sumbangan,
      barang: form.value.barang,
      keterangan: form.value.keterangan,
      anak: anakDays.value.map(v => v || 0).join(','),
      suster: susterDays.value.map(v => v || 0).join(','),
      menginap: menginapDays.value.map(v => v ? '1' : '0').join(','),
      makanan_pagi: makananPagiDays.value.map(v => v ? '1' : '0').join(','),
      makanan_siang: makananSiangDays.value.map(v => v ? '1' : '0').join(','),
      makanan_malam: makananMalamDays.value.map(v => v ? '1' : '0').join(',')
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
      ElMessage.error(err.response.data.error || 'Invalid Inputs, Silahkan periksa kolom form')
    } else {
      ElMessage.error(err.response?.data?.error || err.message || 'Gagal menyimpan data pengabdi')
    }
  } finally {
    submitting.value = false
  }
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
      if (!props.startDate || !props.endDate) {
        fetchKelasDetail()
      }
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
  if (!props.startDate || !props.endDate) {
    fetchKelasDetail()
  }
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
