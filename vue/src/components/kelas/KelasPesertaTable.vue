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
            <div class="header-actions">
              <template v-if="!readonly">
                <el-button type="success" size="small" :icon="Timer" @click.stop="openHistoryDialog">
                  {{ isMobile ? '' : 'Add From History' }}
                </el-button>
                <el-button type="primary" size="small" :icon="Plus" @click.stop="openAddDialog">
                  {{ isMobile ? '' : 'Tambah Peserta' }}
                </el-button>
              </template>
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

            <!-- <el-table-column prop="fotang_ciutao_desc" label="Fotang Ciu Tao" min-width="150" /> -->

            <el-table-column prop="fotang_aktif_desc" label="Fotang Aktif" min-width="150" />

            <el-table-column prop="tanggal_ciu_tao_int" label="Tgl Ciu Tao" width="120" align="center" />

            <!-- <el-table-column prop="pengajak" label="Pengajak" min-width="140" /> -->

            <!-- <el-table-column prop="penanggung" label="Penanggung" min-width="140" /> -->

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

            <el-table-column v-if="isSuaiSingPan" label="Ikrar Yg Dipilih" min-width="240" align="center">
              <template #default="{ row }">
                <div class="ikrar-checkboxes-cell">
                  <el-checkbox :model-value="Boolean(row.ikrar1)" disabled title="Ikrar 1 (重聖輕凡)">1</el-checkbox>
                  <el-checkbox :model-value="Boolean(row.ikrar2)" disabled title="Ikrar 2 (財法雙施)">2</el-checkbox>
                  <el-checkbox :model-value="Boolean(row.ikrar3)" disabled title="Ikrar 3 (清口茹素)">3</el-checkbox>
                  <el-checkbox :model-value="Boolean(row.ikrar4)" disabled title="Ikrar 4 (捨身辦道)">4</el-checkbox>
                  <el-checkbox :model-value="Boolean(row.ikrar5)" disabled title="Ikrar 5 (開設佛堂)">5</el-checkbox>
                  <el-checkbox :model-value="Boolean(row.ikrar6)" disabled title="Ikrar 6 (開荒下種)">6</el-checkbox>
                </div>
              </template>
            </el-table-column>

            <el-table-column v-if="!readonly" label="Aksi" width="90" align="center"
              :fixed="!isDesktop ? false : 'right'">
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
              :layout="(!isMobile ? 'total, ->,' : (Math.ceil(total / pageSize) < 6 ? '-> ,' : '')) + 'prev, pager, next' + (isDesktop ? ', jumper' : '')"
              :pager-count="6" :total="total" @size-change="handleSizeChange" @current-change="handleCurrentChange" />
          </div>
        </div>
      </el-collapse-item>
    </el-collapse>

    <!-- Dialog Popup Add / Edit / History Peserta -->
    <el-dialog v-model="dialogVisible"
      :title="dialogMode === 'add' ? 'Tambah Peserta' : (dialogMode === 'history' ? 'Add From History' : 'Edit Peserta')"
      :width="isMobile ? '90%' : '650px'" destroy-on-close>
      <el-alert v-if="Object.keys(dialogFieldErrors).length > 0" type="error" show-icon title="Invalid Inputs"
        description="Terdapat kesalahan pengisian form. Silahkan periksa pesan kesalahan berwarna merah di bawah."
        class="mb-4" />

      <el-form ref="formRef" :model="form" :rules="formRules" label-width="140px" size="default" v-loading="submitting">
        <el-form-item label="Peserta (Umat)" prop="id_peserta" :error="hasFieldError('id_peserta') ? ' ' : undefined">
          <!-- History Mode: Multi-select UmatPopupSelector trigger -->
          <div v-if="dialogMode === 'history'" class="history-umat-selector">
            <div class="history-tags-wrapper">
              <template v-if="selectedHistoryUmats.length > 0">
                <div class="selected-tags-container" @click="openHistoryPopup">
                  <el-tag v-for="item in selectedHistoryUmats" :key="item.id_peserta || item.id" closable size="small"
                    class="mr-1 mb-1" :title="item.nama_indonesia || item.kode || `ID #${item.id_peserta || item.id}`"
                    @close.stop="removeHistoryUmat(item.id_peserta || item.id)">
                    {{ item.nama_indonesia || item.kode || `ID #${item.id_peserta || item.id}` }}
                  </el-tag>
                </div>
                <div class="history-tags-summary">
                  <el-tag type="success" size="small">{{ selectedHistoryUmats.length }} Peserta</el-tag>
                  <el-button size="small" text type="primary" :icon="MoreFilled" class="ml-2" @click="openHistoryPopup">
                    {{ selectedHistoryUmats.length ? "Ubah" : "Pilih" }}
                  </el-button>
                </div>
              </template>
              <template v-else>
                <el-input placeholder="Klik di sini untuk memilih peserta dari riwayat kelas..." readonly
                  class="clickable-umat-input" @click="openHistoryPopup">
                  <template #append>
                    <el-button :icon="MoreFilled" title="Buka Riwayat Peserta" @click="openHistoryPopup" />
                  </template>
                </el-input>
              </template>
            </div>
            <FieldErrors :errors="getFieldErrors('id_peserta')" />
          </div>

          <!-- Add / Edit Single Mode: Mobile view searchable el-select dropdown -->
          <el-select v-else-if="isMobile" v-model="form.id_peserta" filterable remote reserve-keyword
            placeholder="Ketik nama untuk mencari Umat..." :remote-method="searchUmat" :loading="loadingUmat"
            style="width: 100%" @change="handleUmatSelectChange">
            <el-option v-for="item in umatOptions" :key="item.id" :label="getUmatOptionLabel(item)" :value="item.id" />
          </el-select>

          <!-- Add / Edit Single Mode: Desktop & Tablet view custom popup dialog trigger -->
          <div v-else class="desktop-umat-selector">
            <el-input :model-value="selectedUmatLabel" placeholder="Cari Umat..." readonly class="clickable-umat-input"
              @click="openUmatPopup">
              <template #append>
                <el-button :icon="MoreFilled" title="Buka Pencarian Umat" @click="openUmatPopup">
                </el-button>
              </template>
            </el-input>
          </div>
          <FieldErrors v-if="dialogMode !== 'history'" :errors="getFieldErrors('id_peserta')" />
        </el-form-item>

        <!-- Ikrar Checkboxes (Only when KodeKelas == '004') -->
        <el-form-item v-if="isSuaiSingPan" label="Ikrar Yg Dipilih">
          <div class="ikrar-dialog-checkbox-group">
            <el-checkbox v-model="form.ikrar_1" label="Ikrar 1 (重聖輕凡)" />
            <el-checkbox v-model="form.ikrar_2" label="Ikrar 2 (財法雙施)" />
            <el-checkbox v-model="form.ikrar_3" label="Ikrar 3 (清口茹素)" />
            <el-checkbox v-model="form.ikrar_4" label="Ikrar 4 (捨身辦道)" />
            <el-checkbox v-model="form.ikrar_5" label="Ikrar 5 (開設佛堂)" />
            <el-checkbox v-model="form.ikrar_6" label="Ikrar 6 (開荒下種)" />
          </div>
        </el-form-item>

        <el-form-item v-if="dialogMode === 'edit'" label="Status Lulus"
          :error="hasFieldError('lulus') ? ' ' : undefined">
          <el-switch v-model="form.lulus" active-text="Lulus" inactive-text="Belum Lulus" />
          <FieldErrors :errors="getFieldErrors('lulus')" />
        </el-form-item>

        <el-form-item v-if="dialogMode === 'edit'" label="Keterangan Lulus"
          :error="hasFieldError('keterangan_lulus') ? ' ' : undefined">
          <el-input v-model="form.keterangan_lulus" placeholder="Catatan / keterangan kelulusan" maxlength="200" />
          <FieldErrors :errors="getFieldErrors('keterangan_lulus')" />
        </el-form-item>

        <el-form-item label="Keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
          <el-input v-model="form.keterangan" placeholder="Keterangan tambahan" maxlength="200" />
          <FieldErrors :errors="getFieldErrors('keterangan')" />
        </el-form-item>

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
          {{ dialogMode === 'add' ? 'Simpan' : (dialogMode === 'history' ? `Simpan` : 'Perbarui') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Custom Popup Dialog Selector for Umat (Desktop/Tablet) -->
    <UmatPopupSelector v-model="umatPopupVisible" :fetch-api="umatApi.getUmats"
      :fetch-fotang-api="lookupApi.getLookupFotang" :filter-name="{
        namaindonesia: 'Nama Chiu Tao',
        namamandarin: 'Nama Lain',
        alias: 'Alias / Pin Yin'
      }" :filter-fotang="{
        fotang_chiutao: 'Fotang Ciu Tao'
      }" :Columns="{
        kode: 'Kode',
        nama_indonesia: 'Nama Chiu Tao',
        alias: 'Alias / Pin Yin',
        nama_mandarin: 'Nama Lain',
        fotang_aktif_desc: 'Fotang Aktif',
        pengajak: 'Pengajak',
        penanggung: 'Penanggung',
        alamat: 'Alamat'
      }" :multiple="false" @select="handleUmatSelected" :width="isMobile ? '90%' : '650px'" />

    <!-- Custom Popup Dialog Selector for History Peserta (Multiple Selection) -->
    <UmatPopupSelector v-model="historyPopupVisible" title="Pilih dari Riwayat" :fetch-api="fetchPreviousPesertaApi"
      :selected="selectedHistoryUmats" :Columns="{
        kode: 'Kode',
        nama_indonesia: 'Nama Chiu Tao',
        nama_mandarin: 'Nama Lain',
        alias: 'Alias / Pin Yin',
        fotang_aktif_desc: 'Fotang Aktif',
        fotang_ciu_tao_desc: 'Fotang Ciu Tao',
        pengajak: 'Pengajak',
        penanggung: 'Penanggung',
        alamat: 'Alamat'
      }" :width="isMobile ? '90%' : '650px'" :filter-name="false" :filter-fotang="false" :multiple="true"
      @select="handleHistorySelected" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import { UserFilled, Refresh, Plus, Edit, Delete, MoreFilled, Timer } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import dayjs from 'dayjs'
import { kelasApi } from '../../api/kelas'
import { umatApi } from '../../api/umat'
import lookupApi from '../../api/lookup'
import FieldErrors from '../common/FieldErrors.vue'
import UmatPopupSelector from '../common/UmatPopupSelector.vue'
import type { KelasPeserta, KelasPesertaPrevious, KelasPesertaBulkPayload } from '../../types/kelas'
import type { Umat } from '../../types/umat'
import { scrollToFormError } from '../../utils/scroll'

const props = withDefaults(
  defineProps<{
    kelasId: string | number
    kodeKelas?: string
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

// Accordion collapse state (empty array by default -> collapsed)
const activeNames = ref<string[]>([])
const isExpanded = computed(() => activeNames.value.includes('peserta'))

const pesertaList = ref<KelasPeserta[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
let abortController: AbortController | null = null

const kelasDetail = ref<{ start_date?: string; end_date?: string; kode_kelas?: string } | null>(null)

const currentKodeKelas = computed(() => {
  return props.kodeKelas || kelasDetail.value?.kode_kelas || ''
})

const isSuaiSingPan = computed(() => {
  return currentKodeKelas.value === '004'
})

async function fetchKelasDetail() {
  if (!props.kelasId) return
  if (props.startDate && props.endDate && props.kodeKelas) return
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
const dialogMode = ref<'add' | 'edit' | 'history'>('add')
const submitting = ref(false)
const formRef = ref<FormInstance>()
const dialogFieldErrors = ref<Record<string, string[]>>({})

const historyPopupVisible = ref(false)
const selectedHistoryUmats = ref<KelasPesertaPrevious[]>([])

function getFieldErrors(fieldName: string): string[] {
  return dialogFieldErrors.value[fieldName] || []
}

function hasFieldError(fieldName: string): boolean {
  return getFieldErrors(fieldName).length > 0
}

interface PesertaFormState {
  detail_id?: number | string
  trx_id?: number | string
  id_peserta?: number | string | (number | string)[]
  lulus: boolean
  keterangan_lulus: string
  sumbangan?: number
  barang: string
  tim_kerja: string
  keterangan: string
  ikrar_1: boolean
  ikrar_2: boolean
  ikrar_3: boolean
  ikrar_4: boolean
  ikrar_5: boolean
  ikrar_6: boolean
}

const form = ref<PesertaFormState>({
  lulus: false,
  keterangan_lulus: '',
  sumbangan: 0,
  barang: '',
  tim_kerja: '',
  keterangan: '',
  ikrar_1: false,
  ikrar_2: false,
  ikrar_3: false,
  ikrar_4: false,
  ikrar_5: false,
  ikrar_6: false
})

const formRules: FormRules = {
  id_peserta: [
    {
      validator: (_rule: any, value: any, callback: any) => {
        if (dialogMode.value === 'history') {
          if (!Array.isArray(value) || value.length === 0) {
            return callback(new Error('Minimal satu peserta dari riwayat wajib dipilih'))
          }
          return callback()
        }
        if (!value) {
          return callback(new Error('Peserta (Umat) wajib dipilih'))
        }
        return callback()
      },
      trigger: 'change'
    }
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

async function handleUmatSelectChange(val: number | string | undefined) {
  if (!val) return
  if (isSuaiSingPan.value) {
    try {
      const res = await umatApi.getUmatById(val)
      if (res.data) {
        form.value.ikrar_1 = Boolean(res.data.ikrar_1)
        form.value.ikrar_2 = Boolean(res.data.ikrar_2)
        form.value.ikrar_3 = Boolean(res.data.ikrar_3)
        form.value.ikrar_4 = Boolean(res.data.ikrar_4)
        form.value.ikrar_5 = Boolean(res.data.ikrar_5)
        form.value.ikrar_6 = Boolean(res.data.ikrar_6)
      }
    } catch (err) {
      console.error('Error loading selected umat ikrar:', err)
    }
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

const umatPopupVisible = ref(false)

const selectedUmatLabel = computed(() => {
  if (!form.value.id_peserta) return ''
  const found = umatOptions.value.find((u) => String(u.id) === String(form.value.id_peserta))
  if (found) return getUmatOptionLabel(found)
  return `ID #${form.value.id_peserta}`
})

function openUmatPopup() {
  umatPopupVisible.value = true
}

function handleUmatSelected(selected: Umat) {
  if (!selected) return
  const exists = umatOptions.value.some((u) => u.id === selected.id)
  if (!exists) {
    umatOptions.value.push(selected)
  }
  form.value.id_peserta = selected.id
  handleUmatSelectChange(selected.id)
  if (formRef.value) {
    formRef.value.validateField('id_peserta').catch(() => { })
  }
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
  selectedHistoryUmats.value = []
  form.value = {
    lulus: false,
    keterangan_lulus: '',
    sumbangan: 0,
    barang: '',
    tim_kerja: '',
    keterangan: '',
    ikrar_1: false,
    ikrar_2: false,
    ikrar_3: false,
    ikrar_4: false,
    ikrar_5: false,
    ikrar_6: false
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

// Open Dialog in Add From History Mode
function openHistoryDialog() {
  dialogMode.value = 'history'
  dialogFieldErrors.value = {}
  selectedHistoryUmats.value = []
  form.value = {
    lulus: false,
    keterangan_lulus: '',
    sumbangan: 0,
    barang: '',
    tim_kerja: '',
    keterangan: '',
    id_peserta: [],
    ikrar_1: false,
    ikrar_2: false,
    ikrar_3: false,
    ikrar_4: false,
    ikrar_5: false,
    ikrar_6: false
  }

  const count = daysCount.value
  anakDays.value = new Array(count).fill(0)
  susterDays.value = new Array(count).fill(0)
  menginapDays.value = new Array(count).fill(false)
  makananPagiDays.value = new Array(count).fill(false)
  makananSiangDays.value = new Array(count).fill(false)
  makananMalamDays.value = new Array(count).fill(false)

  dialogVisible.value = true
  historyPopupVisible.value = true
}

function openHistoryPopup() {
  historyPopupVisible.value = true
}

function handleHistorySelected(rows: KelasPesertaPrevious[] | KelasPesertaPrevious) {
  const list = Array.isArray(rows) ? rows : rows ? [rows] : []
  selectedHistoryUmats.value = list
  form.value.id_peserta = list.map((r) => r.id_peserta || (r.id as number))
  if (formRef.value) {
    formRef.value.validateField('id_peserta').catch(() => { })
  }
}

function removeHistoryUmat(id: number | string | undefined) {
  if (!id) return
  selectedHistoryUmats.value = selectedHistoryUmats.value.filter(
    (u) => (u.id_peserta || u.id) !== id
  )
  form.value.id_peserta = selectedHistoryUmats.value.map((u) => u.id_peserta || (u.id as number))
  if (formRef.value) {
    formRef.value.validateField('id_peserta').catch(() => { })
  }
}

async function fetchPreviousPesertaApi(params: Record<string, any>, signal?: AbortSignal) {
  if (!props.kelasId) {
    return { data: [], meta: { total: 0, page: 1, limit: 10 } }
  }
  return await kelasApi.loadPreviousKelasPeserta(props.kelasId, params, signal)
}

// Open Dialog in Edit Mode
function openEditDialog(row: KelasPeserta) {
  dialogMode.value = 'edit'
  dialogFieldErrors.value = {}
  selectedHistoryUmats.value = []
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
    keterangan: row.keterangan || '',
    ikrar_1: Boolean(row.ikrar1 ?? row.umat?.ikrar_1),
    ikrar_2: Boolean(row.ikrar2 ?? row.umat?.ikrar_2),
    ikrar_3: Boolean(row.ikrar3 ?? row.umat?.ikrar_3),
    ikrar_4: Boolean(row.ikrar4 ?? row.umat?.ikrar_4),
    ikrar_5: Boolean(row.ikrar5 ?? row.umat?.ikrar_5),
    ikrar_6: Boolean(row.ikrar6 ?? row.umat?.ikrar_6)
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

// Submit Create, Bulk or Edit Form
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

    if (dialogMode.value === 'history') {
      const idList = (Array.isArray(form.value.id_peserta) ? form.value.id_peserta : [form.value.id_peserta])
        .map((v) => Number(v))
        .filter((v) => !isNaN(v) && v > 0)

      if (idList.length === 0) {
        dialogFieldErrors.value = { id_peserta: ['Minimal satu peserta dari riwayat wajib dipilih'] }
        ElMessage.warning('Pilih minimal satu peserta dari riwayat kelas')
        return
      }

      const payloadBulk: KelasPesertaBulkPayload = {
        trx_id: Number(props.kelasId),
        id_peserta: idList,
        sumbangan: form.value.sumbangan || undefined,
        barang: form.value.barang || undefined,
        tim_kerja: form.value.tim_kerja || undefined,
        keterangan: form.value.keterangan || undefined,
        anak: anakDays.value.map((v) => v || 0).join(','),
        suster: susterDays.value.map((v) => v || 0).join(','),
        menginap: menginapDays.value.map((v) => (v ? '1' : '0')).join(','),
        makanan_pagi: makananPagiDays.value.map((v) => (v ? '1' : '0')).join(','),
        makanan_siang: makananSiangDays.value.map((v) => (v ? '1' : '0')).join(','),
        makanan_malam: makananMalamDays.value.map((v) => (v ? '1' : '0')).join(',')
      }

      if (isSuaiSingPan.value) {
        payloadBulk.umat = {
          ikrar_1: form.value.ikrar_1,
          ikrar_2: form.value.ikrar_2,
          ikrar_3: form.value.ikrar_3,
          ikrar_4: form.value.ikrar_4,
          ikrar_5: form.value.ikrar_5,
          ikrar_6: form.value.ikrar_6
        }
      }

      await kelasApi.createKelasPesertaBulk(payloadBulk)
      ElMessage.success(`${idList.length} peserta dari riwayat berhasil ditambahkan`)
    } else {
      const payload: Partial<KelasPeserta> = {
        trx_id: props.kelasId ? Number(props.kelasId) : undefined,
        id_peserta: form.value.id_peserta ? Number(form.value.id_peserta) : undefined,
        lulus: form.value.lulus,
        keterangan_lulus: form.value.keterangan_lulus,
        sumbangan: form.value.sumbangan,
        barang: form.value.barang,
        tim_kerja: form.value.tim_kerja,
        keterangan: form.value.keterangan,
        anak: anakDays.value.map((v) => v || 0).join(','),
        suster: susterDays.value.map((v) => v || 0).join(','),
        menginap: menginapDays.value.map((v) => (v ? '1' : '0')).join(','),
        makanan_pagi: makananPagiDays.value.map((v) => (v ? '1' : '0')).join(','),
        makanan_siang: makananSiangDays.value.map((v) => (v ? '1' : '0')).join(','),
        makanan_malam: makananMalamDays.value.map((v) => (v ? '1' : '0')).join(',')
      }

      if (isSuaiSingPan.value) {
        payload.umat = {
          ikrar_1: form.value.ikrar_1,
          ikrar_2: form.value.ikrar_2,
          ikrar_3: form.value.ikrar_3,
          ikrar_4: form.value.ikrar_4,
          ikrar_5: form.value.ikrar_5,
          ikrar_6: form.value.ikrar_6
        }
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

.ikrar-checkboxes-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: center;
  align-items: center;
}

:deep(.ikrar-checkboxes-cell .el-checkbox) {
  margin-right: 0;
  height: auto;
}

.ikrar-dialog-checkbox-group {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 8px 16px;
  width: 100%;
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

/* .pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
} */

.desktop-umat-selector {
  width: 100%;
}

.clickable-umat-input :deep(.el-input__inner) {
  cursor: pointer;
}

.history-umat-selector {
  width: 100%;
}

.history-tags-wrapper {
  width: 100%;
}

.selected-tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 0 6px;
  background-color: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color);
  border-radius: var(--el-border-radius-base);
  min-height: 32px;
  max-height: 140px;
  overflow-y: auto;
  align-items: center;
  cursor: pointer;
}

.selected-tags-container :deep(.el-tag) {
  max-width: 85px;
  display: inline-flex;
  align-items: center;
}

.selected-tags-container :deep(.el-tag .el-tag__content) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-tags-summary {
  display: flex;
  align-items: center;
  margin-top: 6px;
}
</style>
