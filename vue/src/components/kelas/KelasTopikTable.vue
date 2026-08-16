<template>
  <div class="topik-accordion-container">
    <el-collapse v-model="activeNames" class="custom-accordion" @change="handleAccordionChange">
      <el-collapse-item name="topik">
        <template #title>
          <div class="accordion-header" @click.stop>
            <div class="header-title">
              <el-icon class="header-icon">
                <Document />
              </el-icon>
              <span>{{ isDesktop ? 'Daftar Topik / Materi Kelas' : 'Topik' }}</span>
              <el-tag size="small" type="info" class="ml-2">{{ total }} Topik</el-tag>
            </div>
            <div class="header-actions" @click.stop>
              <el-button type="primary" size="small" :icon="Plus" @click.stop="openAddDialog">
                {{ isMobile ? '' : 'Tambah Topik' }}
              </el-button>
              <el-button :icon="Refresh" circle size="small" title="Refresh Topik" @click.stop="fetchTopik" />
            </div>
          </div>
        </template>

        <div class="accordion-content">
          <el-table v-loading="loading" :data="topikList" stripe border max-height="450" style="width: 100%"
            empty-text="Belum ada topik yang terdaftar pada kelas ini">
            <el-table-column v-if="isDesktop" prop="urutan" label="Urutan" width="80" align="center" fixed="left" />

            <el-table-column prop="kode_topik" label="Kode Topik" min-width="160">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.kode_topik || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="nama_topik" label="Nama Topik" min-width="160">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.nama_topik || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="topik_date" label="Tgl Topik" width="130" align="center">
              <template #default="{ row }">
                <span>{{ formatDate(row.topik_date) }}</span>
              </template>
            </el-table-column>

            <el-table-column label="Penceramah" min-width="160">
              <template #default="{ row }">
                <span v-if="row.penceramah_ext">{{ row.penceramah_ext }} <el-tag size="small"
                    type="warning">Ext</el-tag></span>
                <span v-else-if="row.penceramah">Umat #{{ row.penceramah }}</span>
                <span v-else>-</span>
              </template>
            </el-table-column>

            <el-table-column prop="topik_date" label="Tanggal" min-width="140" />

            <el-table-column prop="urutan" label="Urutan" min-width="140" />

            <el-table-column prop="durasi" label="Durasi" width="100" align="center">
              <template #default="{ row }">
                <span v-if="row.durasi">{{ row.durasi }} mnt</span>
                <span v-else>-</span>
              </template>
            </el-table-column>

            <el-table-column prop="penterjemah" label="Penterjemah" min-width="140" />

            <el-table-column prop="keterangan" label="Keterangan" min-width="160" />

            <el-table-column label="Aksi" width="90" align="center" :fixed="!isDesktop ? false : 'right'">
              <template #default="{ row }">
                <el-button type="primary" circle size="small" :icon="Edit" @click="openEditDialog(row)" />
                <el-popconfirm title="Yakin ingin menghapus topik ini?" confirm-button-text="Ya, Hapus"
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

    <!-- Dialog Popup Add / Edit Topik -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'add' ? 'Tambah Topik' : 'Edit Topik'"
      :width="isMobile ? '90%' : '600px'" destroy-on-close>
      <el-alert v-if="Object.keys(dialogFieldErrors).length > 0" type="error" show-icon title="Invalid Inputs"
        description="Terdapat kesalahan pengisian form. Silahkan periksa pesan kesalahan berwarna merah di bawah."
        class="mb-4" />

      <el-form ref="formRef" :model="form" :rules="formRules" label-width="140px" size="default" v-loading="submitting">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Nama Topik" prop="kode_topik" :error="hasFieldError('kode_topik') ? ' ' : undefined">
              <LookupSelect v-model="form.kode_topik" placeholder="Pilih Nama Topik..." :fetch-api="fetchTopicLookup"
                value-key="topic_code" label-key="topic_name" :initial-option="initialTopicOption" :clearable="false" />
              <FieldErrors :errors="getFieldErrors('kode_topik')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Kategori">
              <span class="static-text">{{ selectedTopic?.topic_category || '-' }}</span>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Desc">
              <span class="static-text">{{ selectedTopic?.description || form.keterangan || '-' }}</span>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Tanggal Topik" prop="topik_date"
              :error="hasFieldError('topik_date') ? ' ' : undefined">
              <el-date-picker v-model="form.topik_date" type="date" placeholder="Pilih tanggal" format="YYYY-MM-DD"
                value-format="YYYY-MM-DD" style="width: 100%" :clearable="false" :disabled-date="disabledTopikDate" />
              <FieldErrors :errors="getFieldErrors('topik_date')" />
            </el-form-item>
          </el-col>

        </el-row>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Durasi (Menit)" prop="durasi" :error="hasFieldError('durasi') ? ' ' : undefined">
              <el-input-number v-model="form.durasi" :min="1" :step="5" controls-position="right" style="width: 100%" />
              <FieldErrors :errors="getFieldErrors('durasi')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Urutan" prop="urutan" :error="hasFieldError('urutan') ? ' ' : undefined">
              <el-input-number v-model="form.urutan" :min="1" :step="1" controls-position="right" style="width: 100%" />
              <FieldErrors :errors="getFieldErrors('urutan')" />
            </el-form-item>
          </el-col>

        </el-row>

        <el-form-item label="Penceramah Internal" prop="penceramah"
          :error="hasFieldError('penceramah') ? ' ' : undefined">
          <el-select v-model="form.penceramah" filterable remote clearable reserve-keyword
            placeholder="Cari penceramah (Umat)..." :remote-method="searchUmat" :loading="loadingUmat"
            style="width: 100%">
            <el-option v-for="item in umatOptions" :key="item.id" :label="getUmatOptionLabel(item)" :value="item.id" />
          </el-select>
          <FieldErrors :errors="getFieldErrors('penceramah')" />
        </el-form-item>

        <el-form-item label="Penceramah Eksternal" prop="penceramah_ext"
          :error="hasFieldError('penceramah_ext') ? ' ' : undefined">
          <el-input v-model="form.penceramah_ext" placeholder="Nama penceramah luar (bila ada)" maxlength="100" />
          <FieldErrors :errors="getFieldErrors('penceramah_ext')" />
        </el-form-item>

        <el-form-item label="Penterjemah" prop="penterjemah" :error="hasFieldError('penterjemah') ? ' ' : undefined">
          <el-input v-model="form.penterjemah" placeholder="Nama penterjemah" maxlength="200" />
          <FieldErrors :errors="getFieldErrors('penterjemah')" />
        </el-form-item>

        <el-form-item label="Keterangan" prop="keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
          <el-input v-model="form.keterangan" placeholder="Keterangan topik" maxlength="200" />
          <FieldErrors :errors="getFieldErrors('keterangan')" />
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
import { Document, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import dayjs from 'dayjs'
import { kelasApi } from '../../api/kelas'
import { umatApi } from '../../api/umat'
import { topicApi } from '../../api/topic'
import LookupSelect from '../common/LookupSelect.vue'
import FieldErrors from '../common/FieldErrors.vue'
import type { KelasTopik } from '../../types/kelas'
import type { Umat } from '../../types/umat'
import type { Topic } from '../../types/topic'

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
const isExpanded = computed(() => activeNames.value.includes('topik'))

const topikList = ref<KelasTopik[]>([])
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

interface TopikFormState {
  detail_id?: number | string
  trx_id?: number | string
  kode_topik: string
  urutan?: number
  topik_date: string
  penceramah?: number | string
  penceramah_ext: string
  penterjemah: string
  durasi?: number
  keterangan: string
}

const form = ref<TopikFormState>({
  kode_topik: '',
  urutan: 1,
  durasi: 30,
  topik_date: '',
  penceramah_ext: '',
  penterjemah: '',
  keterangan: ''
})

const topicsMap = ref<Map<string, Topic>>(new Map())
const selectedTopic = ref<Topic | null>(null)

async function fetchTopicLookup(params: any, signal?: AbortSignal) {
  const query = params.lookup_description || ''
  const res = await topicApi.getTopikLookup(
    { page: params.page || 1, limit: params.limit || 10, topic_name: query },
    signal
  )
  if (res.data) {
    res.data.forEach((t) => {
      if (t.topic_code) {
        topicsMap.value.set(t.topic_code, t)
      }
    })
  }
  return res
}

watch(
  () => form.value.kode_topik,
  async (newCode) => {
    if (!newCode) {
      selectedTopic.value = null
      return
    }
    if (topicsMap.value.has(newCode)) {
      selectedTopic.value = topicsMap.value.get(newCode) || null
    } else {
      try {
        const res = await topicApi.getTopicById(newCode)
        if (res.data) {
          topicsMap.value.set(res.data.topic_code, res.data)
          selectedTopic.value = res.data
        }
      } catch (err) {
        console.error('Error fetching topic detail:', err)
        selectedTopic.value = null
      }
    }
  },
  { immediate: true }
)

const initialTopicOption = computed(() => {
  if (!form.value.kode_topik) return undefined
  if (selectedTopic.value) return selectedTopic.value
  return { topic_code: form.value.kode_topik, topic_name: form.value.kode_topik }
})

function disabledTopikDate(time: Date): boolean {
  if (props.startDate && dayjs(time).isBefore(dayjs(props.startDate), 'day')) {
    return true
  }
  if (props.endDate && dayjs(time).isAfter(dayjs(props.endDate), 'day')) {
    return true
  }
  return false
}

const validateTopikDateRange = (_rule: any, value: any, callback: any) => {
  if (!value) {
    return callback(new Error('Harap pilih Tanggal Topik'))
  }
  if (props.startDate && dayjs(value).isBefore(dayjs(props.startDate), 'day')) {
    return callback(new Error(`Tanggal Topik tidak boleh sebelum Tanggal Mulai (${props.startDate})`))
  }
  if (props.endDate && dayjs(value).isAfter(dayjs(props.endDate), 'day')) {
    return callback(new Error(`Tanggal Topik tidak boleh setelah Tanggal Selesai (${props.endDate})`))
  }
  callback()
}

const formRules: FormRules = {
  kode_topik: [{ required: true, message: 'Harap pilih Nama Topik', trigger: 'change' }],
  topik_date: [
    { required: true, message: 'Harap pilih Tanggal Topik', trigger: 'change' },
    { validator: validateTopikDateRange, trigger: 'change' }
  ],
  penceramah_ext: [{ required: true, message: 'Harap masukkan Penceramah Eksternal', trigger: 'change' }],
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
  return label
}

function formatDate(val?: string) {
  if (!val) return '-'
  return dayjs(val).format('YYYY-MM-DD')
}

function handleAccordionChange(val: string[] | string) {
  const activeList = Array.isArray(val) ? val : [val]
  if (activeList.includes('topik')) {
    fetchTopik()
  }
}

async function fetchTopik() {
  if (!props.kelasId || !isExpanded.value) return

  if (abortController) abortController.abort()
  abortController = new AbortController()

  loading.value = true
  try {
    const res = await kelasApi.getKelasTopik(
      props.kelasId,
      { page: currentPage.value, limit: pageSize.value },
      abortController.signal
    )
    topikList.value = res.data || []
    total.value = res.meta?.total ?? topikList.value.length
  } catch (err: any) {
    if (err.name === 'AbortError' || err.name === 'CanceledError') return
    console.error('Error fetching kelas topik:', err)
  } finally {
    loading.value = false
  }
}

function handleSizeChange(newSize: number) {
  pageSize.value = newSize
  currentPage.value = 1
  if (isExpanded.value) fetchTopik()
}

function handleCurrentChange(newPage: number) {
  currentPage.value = newPage
  if (isExpanded.value) fetchTopik()
}

function openAddDialog() {
  dialogMode.value = 'add'
  dialogFieldErrors.value = {}

  let defaultDate = dayjs().format('YYYY-MM-DD')
  if (props.startDate) {
    if (dayjs(defaultDate).isBefore(dayjs(props.startDate), 'day') || (props.endDate && dayjs(defaultDate).isAfter(dayjs(props.endDate), 'day'))) {
      defaultDate = props.startDate
    }
  }

  form.value = {
    trx_id: props.kelasId,
    kode_topik: '',
    urutan: (topikList.value.length || 0) + 1,
    topik_date: defaultDate,
    penceramah: undefined,
    penceramah_ext: '',
    penterjemah: '',
    durasi: 30,
    keterangan: ''
  }
  umatOptions.value = []
  dialogVisible.value = true
}

function openEditDialog(row: KelasTopik) {
  dialogMode.value = 'edit'
  dialogFieldErrors.value = {}
  const detailId = row.detailid || row.detail_id
  const rawPenceramah = row.penceramah ? Number(row.penceramah) : undefined

  form.value = {
    detail_id: detailId,
    trx_id: props.kelasId,
    kode_topik: row.kode_topik || '',
    urutan: row.urutan || 1,
    topik_date: row.topik_date ? dayjs(row.topik_date).format('YYYY-MM-DD') : '',
    penceramah: rawPenceramah,
    penceramah_ext: row.penceramah_ext || '',
    penterjemah: row.penterjemah || '',
    durasi: row.durasi,
    keterangan: row.keterangan || ''
  }

  if (rawPenceramah && row.penceramah_nama_indonesia) {
    umatOptions.value = [
      {
        id: rawPenceramah,
        kode: '',
        nama_indonesia: row.penceramah_nama_indonesia
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

    const payload: Partial<KelasTopik> = {
      trx_id: props.kelasId ? Number(props.kelasId) : undefined,
      kode_topik: form.value.kode_topik,
      urutan: form.value.urutan,
      topik_date: form.value.topik_date,
      penceramah: form.value.penceramah ? Number(form.value.penceramah) : undefined,
      penceramah_ext: form.value.penceramah_ext,
      penterjemah: form.value.penterjemah,
      durasi: form.value.durasi,
      keterangan: selectedTopic.value?.description || form.value.keterangan
    }

    if (dialogMode.value === 'add') {
      await kelasApi.createKelasTopik(payload)
      ElMessage.success('Topik kelas berhasil ditambahkan')
    } else {
      const detailId = form.value.detail_id
      if (!detailId) return
      await kelasApi.updateKelasTopik(detailId, payload)
      ElMessage.success('Data topik kelas berhasil diperbarui')
    }

    dialogVisible.value = false
    if (isExpanded.value) fetchTopik()
  } catch (err: any) {
    console.error('Error submitting topik form:', err)
    if (err.response?.data?.details) {
      dialogFieldErrors.value = err.response.data.details
      ElMessage.error(err.response.data.error || 'Invalid Inputs, Silahkan periksa kolom form')
    } else {
      ElMessage.error(err.response?.data?.error || err.message || 'Gagal menyimpan data topik')
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: KelasTopik) {
  const detailId = row.detailid || row.detail_id
  if (!detailId) return

  try {
    await kelasApi.deleteKelasTopik(detailId, props.kelasId)
    ElMessage.success('Topik kelas berhasil dihapus')
    if (isExpanded.value) fetchTopik()
  } catch (err: any) {
    console.error('Error deleting topik:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal menghapus data topik')
  }
}

watch(
  () => props.kelasId,
  (newId) => {
    if (newId) {
      currentPage.value = 1
      if (isExpanded.value) {
        fetchTopik()
      } else {
        topikList.value = []
        total.value = 0
      }
    }
  }
)

onMounted(() => {
  if (isExpanded.value) fetchTopik()
})

onUnmounted(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.topik-accordion-container {
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

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}

.static-text {
  font-size: 14px;
  color: var(--el-text-color-regular);
  line-height: 32px;
}
</style>
