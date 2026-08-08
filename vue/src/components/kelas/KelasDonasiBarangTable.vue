<template>
  <div class="donasi-barang-accordion-container">
    <el-collapse v-model="activeNames" class="custom-accordion" @change="handleAccordionChange">
      <el-collapse-item name="donasi_barang">
        <template #title>
          <div class="accordion-header" @click.stop>
            <div class="header-title">
              <el-icon class="header-icon">
                <Box />
              </el-icon>
              <span>{{ isDesktop ? 'Daftar Donasi Barang Kelas' : isMobile ? 'D.Barang' : 'Donasi Barang' }}</span>
              <el-tag size="small" type="info" class="ml-2">{{ total }} Barang</el-tag>
            </div>
            <div class="header-actions" @click.stop>
              <el-button type="primary" size="small" :icon="Plus" @click.stop="openAddDialog">
                {{ isMobile ? '' : 'Tambah Donasi Barang' }}
              </el-button>
              <el-button :icon="Refresh" circle size="small" title="Refresh Donasi Barang"
                @click.stop="fetchDonasiBarang" />
            </div>
          </div>
        </template>

        <div class="accordion-content">
          <el-table v-loading="loading" :data="donasiBarangList" stripe border max-height="450" style="width: 100%"
            empty-text="Belum ada donasi barang yang terdaftar pada kelas ini">
            <el-table-column v-if="isDesktop" type="index" label="No." width="60" align="center" fixed="left" />

            <el-table-column prop="donatur" label="Nama Donatur" min-width="180">
              <template #default="{ row }">
                <span class="font-semibold">{{ row.donatur || '-' }}</span>
              </template>
            </el-table-column>

            <el-table-column prop="barang" label="Nama Barang / Deskripsi" min-width="220" />

            <el-table-column label="Aksi" width="90" align="center" :fixed="!isDesktop ? false : 'right'">
              <template #default="{ row }">
                <el-button type="primary" circle size="small" :icon="Edit" @click="openEditDialog(row)" />
                <el-popconfirm title="Yakin ingin menghapus donasi barang ini?" confirm-button-text="Ya, Hapus"
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

    <!-- Dialog Popup Add / Edit Donasi Barang -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'add' ? 'Tambah Donasi Barang' : 'Edit Donasi Barang'"
      width="520px" destroy-on-close>
      <el-alert v-if="Object.keys(dialogFieldErrors).length > 0" type="error" show-icon title="Invalid Inputs"
        description="Terdapat kesalahan pengisian form. Silahkan periksa pesan kesalahan berwarna merah di bawah."
        class="mb-4" />

      <el-form ref="formRef" :model="form" :rules="formRules" label-width="130px" size="default" v-loading="submitting">
        <el-form-item label="Nama Donatur" prop="donatur" :error="hasFieldError('donatur') ? ' ' : undefined">
          <el-input v-model="form.donatur" placeholder="Nama donatur" maxlength="100" />
          <FieldErrors :errors="getFieldErrors('donatur')" />
        </el-form-item>

        <el-form-item label="Nama Barang" prop="barang" :error="hasFieldError('barang') ? ' ' : undefined">
          <el-input v-model="form.barang" placeholder="Deskripsi barang donasi" maxlength="100" />
          <FieldErrors :errors="getFieldErrors('barang')" />
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
import { Box, Refresh, Plus, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import axios from 'axios'
import { kelasApi } from '../../api/kelas'
import FieldErrors from '../common/FieldErrors.vue'
import type { KelasDonasiBarang } from '../../types/kelas'

const props = defineProps<{
  kelasId: string | number
}>()

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
// const isTablet = breakpoints.smaller('lg')   // True if width < 1024px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

const activeNames = ref<string[]>([])
const isExpanded = computed(() => activeNames.value.includes('donasi_barang'))

const donasiBarangList = ref<KelasDonasiBarang[]>([])
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

interface DonasiBarangFormState {
  detail_id?: number | string
  trx_id?: number | string
  donatur: string
  barang: string
}

const form = ref<DonasiBarangFormState>({
  donatur: '',
  barang: ''
})

const formRules: FormRules = {
  donatur: [{ required: true, message: 'Harap isi nama donatur', trigger: 'blur' }],
  barang: [{ required: true, message: 'Harap isi nama barang', trigger: 'blur' }]
}

function handleAccordionChange(val: string[] | string) {
  const activeList = Array.isArray(val) ? val : [val]
  if (activeList.includes('donasi_barang')) {
    fetchDonasiBarang()
  }
}

async function fetchDonasiBarang() {
  if (!props.kelasId || !isExpanded.value) return

  if (abortController) abortController.abort()
  abortController = new AbortController()

  loading.value = true
  try {
    const res = await kelasApi.getKelasDonasiBarang(
      props.kelasId,
      { page: currentPage.value, limit: pageSize.value },
      abortController.signal
    )
    donasiBarangList.value = res.data || []
    total.value = res.meta?.total ?? donasiBarangList.value.length
  } catch (err: any) {
    if (err.name === 'AbortError' || err.name === 'CanceledError') return
    console.error('Error fetching kelas donasi barang:', err)
  } finally {
    loading.value = false
  }
}

function handleSizeChange(newSize: number) {
  pageSize.value = newSize
  currentPage.value = 1
  if (isExpanded.value) fetchDonasiBarang()
}

function handleCurrentChange(newPage: number) {
  currentPage.value = newPage
  if (isExpanded.value) fetchDonasiBarang()
}

function openAddDialog() {
  dialogMode.value = 'add'
  dialogFieldErrors.value = {}
  form.value = {
    trx_id: props.kelasId,
    donatur: '',
    barang: ''
  }
  dialogVisible.value = true
}

function openEditDialog(row: KelasDonasiBarang) {
  dialogMode.value = 'edit'
  dialogFieldErrors.value = {}
  const detailId = row.detailid || row.detail_id

  form.value = {
    detail_id: detailId,
    trx_id: props.kelasId,
    donatur: row.donatur || '',
    barang: row.barang || ''
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

    const payload: Partial<KelasDonasiBarang> = {
      trx_id: props.kelasId ? Number(props.kelasId) : undefined,
      donatur: form.value.donatur,
      barang: form.value.barang
    }

    if (dialogMode.value === 'add') {
      await kelasApi.createKelasDonasiBarang(payload)
      ElMessage.success('Donasi barang berhasil ditambahkan')
    } else {
      const detailId = form.value.detail_id
      if (!detailId) return
      await kelasApi.updateKelasDonasiBarang(detailId, payload)
      ElMessage.success('Data donasi barang berhasil diperbarui')
    }

    dialogVisible.value = false
    if (isExpanded.value) fetchDonasiBarang()
  } catch (err: unknown) {
    console.error('Error submitting donasi barang form:', err)
    if (axios.isAxiosError(err) && err.response?.data?.details) {
      dialogFieldErrors.value = err.response.data.details
      ElMessage.error(err.response.data.error || 'Invalid Inputs, Silahkan periksa kolom form')
    } else {
      const msg = (axios.isAxiosError(err) && err.response?.data?.error) || (err instanceof Error ? err.message : 'Gagal menyimpan data donasi barang')
      ElMessage.error(msg)
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: KelasDonasiBarang) {
  const detailId = row.detailid || row.detail_id
  if (!detailId) return

  try {
    await kelasApi.deleteKelasDonasiBarang(detailId, props.kelasId)
    ElMessage.success('Donasi barang berhasil dihapus')
    if (isExpanded.value) fetchDonasiBarang()
  } catch (err: unknown) {
    console.error('Error deleting donasi barang:', err)
    const msg = (axios.isAxiosError(err) && err.response?.data?.error) || (err instanceof Error ? err.message : 'Gagal menghapus data donasi barang')
    ElMessage.error(msg)
  }
}

watch(
  () => props.kelasId,
  (newId) => {
    if (newId) {
      currentPage.value = 1
      if (isExpanded.value) {
        fetchDonasiBarang()
      } else {
        donasiBarangList.value = []
        total.value = 0
      }
    }
  }
)

onMounted(() => {
  if (isExpanded.value) fetchDonasiBarang()
})

onUnmounted(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.donasi-barang-accordion-container {
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
</style>
