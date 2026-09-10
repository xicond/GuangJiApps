<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">Master Data Topic</h2>
        <p class="page-subtitle">Kelola daftar data topic, pencarian, serta pembaruan profil</p>
      </div>
      <el-button type="primary" size="large" :icon="Plus" class="create-btn" @click="handleCreate">
        Tambah Topic Baru
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
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Kode Topik" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.topic_code" placeholder="Cari kode topik..." clearable :prefix-icon="Search"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Topik" :label-position="isMobile ? 'top' : 'right'">
            <el-input v-model="filters.topic_name" placeholder="Cari nama topik..." clearable :prefix-icon="Search"
              @input="onFilterChange" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Kategori" :label-position="isMobile ? 'top' : 'right'">
            <LookupSelect v-model="filters.topic_category" placeholder="Pilih Kategori Topik..."
              :fetch-api="lookupApi.getLookupKategoriTopic" value-key="lookup_value" label-key="lookup_description"
              clearable @change="onFilterChange" auto-populate />
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
        empty-text="Tidak ada data topic yang ditemukan">
        <el-table-column v-if="isDesktop" :index="getRowIndex" type="index" label="No." width="70" align="center"
          fixed="left" />

        <el-table-column prop="topic_code" label="Kode Topik" width="130" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info" class="font-mono">{{ row.topic_code }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="topic_name" label="Nama Topik" min-width="200">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.topic_name || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="topic_category" label="Kategori" min-width="150">
          <template #default="{ row }">
            <span>{{ row.topic_category_info?.lookup_description || row.topic_category || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="description" label="Keterangan" min-width="220">
          <template #default="{ row }">
            <span>{{ row.description || '-' }}</span>
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
              <el-button type="primary" size="small" circle :icon="Edit" title="Edit Topic"
                @click="handleEdit(row.topic_code)" />

              <el-popconfirm title="Apakah Anda yakin ingin menghapus data ini?" confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal" confirm-button-type="danger" @confirm="handleDelete(row.topic_code)">
                <template #reference>
                  <el-button type="danger" size="small" circle :icon="Delete" title="Hapus Topic" />
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
          :layout="(!isMobile ? 'total, ->,' : (Math.ceil(pagination.total / pagination.limit) < 6 ? '-> ,' : '')) + 'prev, pager, next' + (isDesktop ? ', jumper' : '')"
          :pager-count="6" @size-change="handleSizeChange" @current-change="handlePageChange" />
      </div>
    </el-card>

    <!-- Create / Edit Dialog -->
    <el-dialog v-model="dialogVisible" :title="isEditing ? `Edit Topic #${editingCode}` : 'Tambah Topic Baru'"
      :width="isMobile ? '90%' : '560px'" destroy-on-close @closed="resetForm">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="130px"
        :label-position="isMobile ? 'top' : 'right'" :disabled="submitting">
        <el-form-item label="Kode Topik" prop="topic_code">
          <el-input v-model="formData.topic_code" placeholder="Masukkan kode topik (misal: TP01)"
            :disabled="isEditing" />
          <FieldErrors :errors="fieldErrors.topic_code" />
        </el-form-item>

        <el-form-item label="Nama Topik" prop="topic_name">
          <el-autocomplete v-if="!isEditing" v-model="formData.topic_name" :fetch-suggestions="queryTopicSuggestions"
            :trigger-on-focus="false" placeholder="Masukkan nama topik" clearable style="width: 100%"
            @select="handleSelectSuggestion">
            <template #default="{ item }">
              <div class="topic-suggest-item">
                <div class="topic-suggest-row">
                  <span class="topic-suggest-name font-semibold">{{ item.topic_name }}</span>
                  <el-tag size="small" type="info" class="font-mono">{{ item.topic_code }}</el-tag>
                </div>
                <div v-if="item.topic_category || item.description" class="topic-suggest-desc">
                  <span v-if="item.topic_category" class="topic-suggest-cat">{{ item.topic_category }}</span>
                  <span v-if="item.topic_category && item.description"> &bull; </span>
                  <span v-if="item.description" class="topic-suggest-text">{{ item.description }}</span>
                </div>
              </div>
            </template>
          </el-autocomplete>
          <el-input v-else v-model="formData.topic_name" placeholder="Masukkan nama topik" />

          <div v-if="fieldErrors.topic_name?.length" class="field-errors-list">
            <div v-for="(errMsg, idx) in fieldErrors.topic_name" :key="idx" class="field-error-item">
              <span class="error-bullet">&bull;</span>
              <span>
                {{ errMsg }}<template v-if="duplicateTopicCode && errMsg.includes('sudah ada')">,
                  <el-link type="primary" :underline="true" class="duplicate-edit-link"
                    @click="handleEdit(duplicateTopicCode)">
                    update here instead
                  </el-link>
                </template>
              </span>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="Kategori Topik" prop="topic_category">
          <LookupSelect v-model="formData.topic_category" placeholder="Pilih Kategori Topik..."
            :fetch-api="lookupApi.getLookupKategoriTopic" value-key="lookup_value" label-key="lookup_description"
            clearable auto-populate />
          <FieldErrors :errors="fieldErrors.topic_category" />
        </el-form-item>

        <el-form-item label="Keterangan" prop="description">
          <el-input v-model="formData.description" type="textarea" :rows="3"
            placeholder="Masukkan keterangan tambahan..." />
          <FieldErrors :errors="fieldErrors.description" />
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
import { topicApi } from '../../api/topic'
import { lookupApi } from '../../api/lookup'
import type { Topic, TopicQueryParams } from '../../types/topic'
import LookupSelect from '../../components/common/LookupSelect.vue'
import FieldErrors from '../../components/common/FieldErrors.vue'
import { scrollToFormError } from '../../utils/scroll'

// Breakpoints layout calculation
const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')
const isDesktop = breakpoints.greaterOrEqual('lg')

// Dataset state
const dataList = shallowRef<Topic[]>([])
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
const filters = reactive<TopicQueryParams>({
  topic_code: '',
  topic_name: '',
  topic_category: ''
})

// Dialog & Form state
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingCode = ref('')
const submitting = ref(false)
const formRef = ref<FormInstance | null>(null)
const fieldErrors = ref<Record<string, string[]>>({})
const duplicateTopicCode = ref('')

interface TopicSuggestion {
  value: string
  topic_code: string
  topic_name: string
  topic_category?: string
  description?: string
}

const formData = reactive<Partial<Topic>>({
  topic_code: '',
  topic_name: '',
  topic_category: '',
  description: '',
  // status: true
})

const formRules: FormRules = {
  topic_code: [
    { required: true, message: 'Kode topik wajib diisi', trigger: 'blur' },
    { max: 20, message: 'Kode topik maksimal 20 karakter', trigger: 'blur' }
  ],
  topic_name: [
    { required: true, message: 'Nama topik wajib diisi', trigger: 'blur' },
    { max: 300, message: 'Nama topik maksimal 300 karakter', trigger: 'blur' }
  ]
}

let currentAbortController: AbortController | null = null
let suggestAbortController: AbortController | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const queryTopicSuggestions = async (
  queryString: string,
  cb: (suggestions: TopicSuggestion[]) => void
) => {
  const trimmed = queryString ? queryString.trim() : ''
  const words = trimmed.split(/\s+/).filter(Boolean)

  if (words.length < 2) {
    cb([])
    return
  }

  if (suggestAbortController) {
    suggestAbortController.abort()
  }
  suggestAbortController = new AbortController()

  try {
    const res = await topicApi.getTopics(
      {
        topic_name: trimmed,
        limit: 10
      },
      suggestAbortController.signal
    )
    const items: TopicSuggestion[] = (res.data || []).map((t) => ({
      value: t.topic_name,
      topic_code: t.topic_code,
      topic_name: t.topic_name,
      topic_category: t.topic_category_info?.lookup_description || t.topic_category,
      description: t.description
    }))
    cb(items)
  } catch (err: unknown) {
    if (err instanceof Error && (err.name === 'CanceledError' || err.name === 'AbortError')) return
    cb([])
  }
}

function handleSelectSuggestion(item: TopicSuggestion) {
  if (item?.topic_code) {
    handleEdit(item.topic_code)
  }
}

async function fetchData() {
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()
  loading.value = true

  try {
    const res = await topicApi.getTopics(
      {
        page: pagination.page,
        limit: pagination.limit,
        topic_code: filters.topic_code?.trim(),
        topic_name: filters.topic_name?.trim(),
        topic_category: filters.topic_category?.trim()
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
  filters.topic_code = ''
  filters.topic_name = ''
  filters.topic_category = ''
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
  formData.topic_code = ''
  formData.topic_name = ''
  formData.topic_category = ''
  formData.description = ''
  // formData.status = true
  isEditing.value = false
  editingCode.value = ''
  fieldErrors.value = {}
  duplicateTopicCode.value = ''
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

function handleCreate() {
  resetForm()
  isEditing.value = false
  dialogVisible.value = true
}

async function handleEdit(code: string) {
  resetForm()
  isEditing.value = true
  editingCode.value = code

  try {
    const res = await topicApi.getTopicById(code)
    const topic = res.data
    if (topic) {
      formData.topic_code = topic.topic_code
      formData.topic_name = topic.topic_name
      formData.topic_category = topic.topic_category || ''
      formData.description = topic.description || ''
      // formData.status = topic.status !== undefined ? topic.status : true
      dialogVisible.value = true
    }
  } catch (err: unknown) {
    console.error('Error fetching topic by code:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal memuat detail topic')
  }
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    scrollToFormError()
    return
  }

  fieldErrors.value = {}
  duplicateTopicCode.value = ''
  submitting.value = true
  try {
    const payload: Partial<Topic> = {
      topic_code: formData.topic_code?.trim(),
      topic_name: formData.topic_name?.trim(),
      topic_category: formData.topic_category || undefined,
      description: formData.description?.trim() || undefined
      // status: formData.status
    }

    if (isEditing.value && editingCode.value) {
      await topicApi.updateTopic(editingCode.value, payload)
      ElNotification({
        title: 'Berhasil',
        message: `Topic ${editingCode.value} berhasil diperbarui`,
        type: 'success'
      })
    } else {
      await topicApi.createTopic(payload)
      ElNotification({
        title: 'Berhasil',
        message: 'Topic baru berhasil ditambahkan',
        type: 'success'
      })
    }

    dialogVisible.value = false
    fetchData()
  } catch (err: unknown) {
    console.error('Failed to submit topic form:', err)
    const errObj = err as {
      response?: {
        data?: {
          error?: string
          details?: Record<string, string[]>
        }
      }
      message?: string
    }

    if (errObj.response?.data?.details) {
      fieldErrors.value = errObj.response.data.details

      const topicNameErrors = fieldErrors.value.topic_name || []
      const isDuplicateError = topicNameErrors.some((msg) => msg.includes('sudah ada'))

      if (isDuplicateError && formData.topic_name) {
        try {
          const searchRes = await topicApi.getTopics({
            topic_name: formData.topic_name.trim(),
            limit: 5
          })
          const matched =
            searchRes.data?.find(
              (t) => t.topic_name.trim().toLowerCase() === formData.topic_name!.trim().toLowerCase()
            ) || searchRes.data?.[0]

          if (matched) {
            duplicateTopicCode.value = matched.topic_code
          }
        } catch (searchErr) {
          console.warn('Failed to resolve duplicate topic code:', searchErr)
        }
      }
    }

    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal menyimpan data topic')
    scrollToFormError()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(code: string) {
  try {
    await topicApi.deleteTopic(code)
    ElNotification({
      title: 'Berhasil',
      message: `Data topic ${code} berhasil dihapus`,
      type: 'success'
    })
    fetchData()
  } catch (err: unknown) {
    console.error('Failed to delete item:', err)
    const errObj = err as { response?: { data?: { error?: string } }; message?: string }
    ElMessage.error(errObj.response?.data?.error || errObj.message || 'Gagal menghapus data topic')
  }
}

onMounted(() => {
  fetchData()
})

onUnmounted(() => {
  if (currentAbortController) currentAbortController.abort()
  if (suggestAbortController) suggestAbortController.abort()
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
  margin-top: 1.25rem;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.topic-suggest-item {
  display: flex;
  flex-direction: column;
  padding: 4px 0;
  line-height: 1.3;
}

.topic-suggest-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.topic-suggest-name {
  color: var(--el-text-color-primary);
  font-size: 0.875rem;
}

.topic-suggest-desc {
  font-size: 0.75rem;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.topic-suggest-cat {
  color: var(--el-color-primary);
  font-weight: 500;
}

.topic-suggest-text {
  color: var(--el-text-color-secondary);
}

.field-errors-list {
  margin-top: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 100%;
}

.field-error-item {
  color: var(--el-color-danger, #f56c6c);
  font-size: 0.75rem;
  line-height: 1.25;
  display: flex;
  align-items: flex-start;
  gap: 4px;
}

.error-bullet {
  font-weight: bold;
}

.duplicate-edit-link {
  font-size: 0.75rem;
  font-weight: 600;
  vertical-align: baseline;
  text-decoration: underline;
  margin-left: 4px;
}
</style>
