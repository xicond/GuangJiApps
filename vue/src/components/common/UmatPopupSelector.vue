<template>
  <el-dialog v-model="internalVisible" :title="title" :width="dialogWidth" :append-to-body="appendToBody"
    destroy-on-close class="umat-popup-dialog" @open="onDialogOpen" @close="onDialogClose">
    <!-- Filter Panel -->
    <div v-if="hasFilters" class="filter-panel">
      <el-form :inline="true" size="default" class="filter-form" @submit.prevent="handleSearch">
        <!-- Text filters (from filterName) -->
        <template v-if="resolvedFilterNames">
          <el-form-item v-for="filter in resolvedFilterNames" :key="filter.key" :label="filter.label"
            class="filter-item">
            <el-input v-model="filterValues[filter.key]" :placeholder="'Ketik ' + filter.label + '...'" clearable
              @keyup.enter="handleSearch" />
          </el-form-item>
        </template>

        <!-- Fotang dropdown filters (from filterFotang) -->
        <template v-if="resolvedFilterFotang">
          <el-form-item v-for="fotang in resolvedFilterFotang" :key="fotang.key" :label="fotang.label"
            class="filter-item">
            <el-select v-model="filterValues[fotang.key]" :placeholder="'Pilih ' + fotang.label + '...'" clearable
              filterable :loading="fotangLoading" style="min-width: 180px;">
              <el-option v-for="opt in internalFotangOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </el-form-item>
        </template>

        <!-- Action buttons in filter row -->
        <el-form-item class="filter-actions">
          <el-button type="primary" :icon="Search" :loading="loading" @click="handleSearch">
            Search
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- Action Toolbar (Add Content / Pilih) -->
    <div class="action-toolbar">
      <div class="right-title">
        <span class="grid-title">List Umat</span>
      </div>
    </div>

    <!-- Data Table -->
    <el-table ref="tableRef" :row-key="getRowKey" v-loading="loading" :data="tableData" border stripe
      highlight-current-row max-height="420" style="width: 100%" empty-text="Tidak ada data ditemukan"
      @selection-change="handleSelectionChange" @row-click="handleRowClick" @row-dblclick="handleRowDblClick">
      <!-- Selection Column -->
      <template v-if="multiple">
        <el-table-column type="selection" :reserve-selection="true" width="50" align="center" fixed="left" />
      </template>
      <template v-else>
        <el-table-column width="60" align="center" fixed="left">
          <template #default="{ row }">
            <el-radio :model-value="selectedSingleId" :label="row.id" class="single-radio"
              @change="selectSingleRow(row)">
              <span style="display: none;"></span>
            </el-radio>
          </template>
        </el-table-column>
      </template>

      <!-- Row Number Column -->
      <el-table-column v-if="!isMobile" label="No" width="60" align="center" fixed="left">
        <template #default="{ $index }">
          {{ (currentPage - 1) * pageSizeInternal + $index + 1 }}
        </template>
      </el-table-column>

      <!-- Dynamic Columns -->
      <el-table-column v-for="col in resolvedColumnsList" :key="col.prop" :prop="col.prop" :label="col.label"
        :min-width="col.minWidth || 130" :width="col.width" :align="col.align || 'left'" show-overflow-tooltip>
        <template #default="{ row }">
          <span>{{ formatCellValue(row, col.prop) }}</span>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination Footer -->
    <div class="pagination-footer">
      <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSizeInternal"
        :page-sizes="[10, 20, 50, 100]"
        :layout="(!isMobile ? 'total,' : '->,') + 'prev, pager, next' + (isDesktop ? ', jumper' : '')" :total="total"
        @size-change="handleSizeChange" @current-change="handleCurrentChange" />
    </div>

    <!-- Dialog Footer -->
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="closeDialog">Batal</el-button>
        <el-button type="primary" :disabled="isConfirmDisabled" @click="handleConfirm">
          {{ multiple && selectedMultipleRows.length > 0 ? `Pilih (${selectedMultipleRows.length})` : 'Pilih' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessage, type TableInstance } from 'element-plus'

export interface ColumnDefinition {
  prop: string
  label: string
  minWidth?: number | string
  width?: number | string
  align?: 'left' | 'center' | 'right'
}


// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

export interface FotangOption {
  label: string
  value: string | number
}

const props = withDefaults(
  defineProps<{
    modelValue?: boolean
    visible?: boolean
    title?: string
    /**
     * Data fetch function. Must not contain direct API imports inside component.
     */
    fetchApi: (
      params: Record<string, any>,
      signal?: AbortSignal
    ) => Promise<{ data?: any[]; meta?: { total?: number; page?: number; limit?: number }; total?: number } | any>
    /**
     * Optional Fotang fetcher function for dropdown options.
     */
    fetchFotangApi?: ((params?: any, signal?: AbortSignal) => Promise<any>) | null
    /**
     * Optional pre-populated Fotang options array.
     */
    fotangOptions?: FotangOption[]
    /**
     * Text filter configuration.
     * Example: { namaindonesia: 'Nama Chiu Tao', namamandarin: 'Nama Lain', alias: 'Alias / Pin Yin' }
     * Pass null or false to disable.
     */
    filterName?: Record<string, string> | false | null
    /**
     * Fotang dropdown filter configuration.
     * Example: { fotang_chiutao: 'Fotang Ciu Tao', fotang_aktif: 'Fotang Aktif' }
     * Pass null or false to disable.
     */
    filterFotang?: Record<string, string> | false | null
    /**
     * Columns definition object or array.
     * Example: { kode: 'Kode', nama_indonesia: 'Nama Chiu Tao', ... }
     */
    Columns?: Record<string, string> | ColumnDefinition[] | null
    columns?: Record<string, string> | ColumnDefinition[] | null
    /**
     * Multiple selection mode. Default is false (single selection).
     */
    multiple?: boolean
    /**
     * Pre-selected rows for multiple selection.
     */
    selected?: any[]
    /**
     * Unique key attribute or getter function for row identity.
     */
    rowKey?: string | ((row: any) => string | number)
    pageSize?: number
    width?: string
    appendToBody?: boolean
  }>(),
  {
    modelValue: false,
    visible: undefined,
    title: 'Popup Umat',
    fetchFotangApi: null,
    fotangOptions: () => [],
    filterName: undefined,
    filterFotang: undefined,
    Columns: null,
    columns: null,
    multiple: false,
    selected: () => [],
    rowKey: undefined,
    pageSize: 10,
    width: '950px',
    appendToBody: true
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'update:visible', value: boolean): void
  (e: 'select', selected: any): void
  (e: 'onselect', selected: any): void
  (e: 'close'): void
}>()

// Modal Visibility Handling
const internalVisible = computed({
  get: () => {
    if (props.visible !== undefined) return props.visible
    return props.modelValue
  },
  set: (val: boolean) => {
    emit('update:modelValue', val)
    emit('update:visible', val)
    if (!val) {
      emit('close')
    }
  }
})

const dialogWidth = computed(() => {
  return props.width || '950px'
})

// Filter Setup
const defaultFilterNames: Record<string, string> = {
  namaindonesia: 'Nama Chiu Tao',
  namamandarin: 'Nama Lain',
  alias: 'Alias / Pin Yin'
}

const defaultFilterFotang: Record<string, string> = {
  fotang_chiutao: 'Fotang Ciu Tao'
}

const resolvedFilterNames = computed<{ key: string; label: string }[] | null>(() => {
  if (props.filterName === false || props.filterName === null) return null
  const source = props.filterName !== undefined ? props.filterName : defaultFilterNames
  if (!source || typeof source !== 'object') return null
  return Object.entries(source).map(([key, label]) => ({ key, label }))
})

const resolvedFilterFotang = computed<{ key: string; label: string }[] | null>(() => {
  if (props.filterFotang === false || props.filterFotang === null) return null
  const source = props.filterFotang !== undefined ? props.filterFotang : defaultFilterFotang
  if (!source || typeof source !== 'object') return null
  return Object.entries(source).map(([key, label]) => ({ key, label }))
})

const hasFilters = computed(() => {
  return (
    (resolvedFilterNames.value && resolvedFilterNames.value.length > 0) ||
    (resolvedFilterFotang.value && resolvedFilterFotang.value.length > 0)
  )
})

const filterValues = ref<Record<string, any>>({})

// Columns Setup
const defaultColumnsList: ColumnDefinition[] = [
  { prop: 'kode', label: 'Kode', minWidth: 120 },
  { prop: 'nama_indonesia', label: 'Nama Chiu Tao', minWidth: 150 },
  { prop: 'alias', label: 'Alias / Pin Yin', minWidth: 120 },
  { prop: 'nama_mandarin', label: 'Nama Lain', minWidth: 120 },
  { prop: 'fotang_aktif_desc', label: 'Fotang Aktif', minWidth: 140 },
  { prop: 'pengajak', label: 'Pengajak', minWidth: 120 },
  { prop: 'penanggung', label: 'Penanggung', minWidth: 120 },
  { prop: 'alamat', label: 'Alamat', minWidth: 200 }
]

const resolvedColumnsList = computed<ColumnDefinition[]>(() => {
  const custom = props.Columns || props.columns
  if (!custom) return defaultColumnsList

  if (Array.isArray(custom)) {
    return custom.map((c) => {
      if (typeof c === 'string') return { prop: c, label: c, minWidth: 120 }
      return c as ColumnDefinition
    })
  }

  return Object.entries(custom).map(([key, label]) => ({
    prop: key,
    label: String(label),
    minWidth: 130
  }))
})

// Data & Table State
const tableRef = ref<TableInstance>()
const tableData = ref<any[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSizeInternal = ref(props.pageSize || 10)
const total = ref(0)
let abortController: AbortController | null = null

// Fotang Options Handling
const fotangLoading = ref(false)
const internalFotangOptions = ref<FotangOption[]>([])

async function loadFotangOptions() {
  if (props.fotangOptions && props.fotangOptions.length > 0) {
    internalFotangOptions.value = props.fotangOptions
    return
  }
  if (!props.fetchFotangApi) return

  try {
    fotangLoading.value = true
    const res = await props.fetchFotangApi({ limit: 0 })
    const list = Array.isArray(res) ? res : res?.data || []
    internalFotangOptions.value = list.map((item: any) => ({
      value: item.lookup_value ?? item.value ?? item.id ?? item.code,
      label:
        item.lookup_description ??
        item.label ??
        item.name ??
        item.fotang_name ??
        String(item.lookup_value ?? item.value)
    }))
  } catch (err) {
    console.error('Error loading fotang options in UmatPopupSelector:', err)
  } finally {
    fotangLoading.value = false
  }
}

// Selection State
const selectedSingleId = ref<string | number | null>(null)
const selectedSingleRow = ref<any | null>(null)
const selectedMultipleRows = ref<any[]>([])

const isConfirmDisabled = computed(() => {
  if (props.multiple) {
    return selectedMultipleRows.value.length === 0
  }
  return !selectedSingleRow.value
})

function selectSingleRow(row: any) {
  selectedSingleId.value = row.id
  selectedSingleRow.value = row
}

function getRowKey(row: any): string {
  if (!row) return ''
  if (typeof props.rowKey === 'function') {
    return String(props.rowKey(row))
  }
  if (typeof props.rowKey === 'string' && props.rowKey) {
    return String(row[props.rowKey] ?? '')
  }
  const val = row.id_peserta ?? row.id ?? row.kode ?? ''
  return val ? String(val) : JSON.stringify(row)
}

function handleRowClick(row: any) {
  if (!props.multiple) {
    selectSingleRow(row)
  }
}

function handleRowDblClick(row: any) {
  if (!props.multiple) {
    selectSingleRow(row)
    handleConfirm()
  }
}

function handleSelectionChange(rows: any[]) {
  if (props.multiple) {
    const currentPageKeys = new Set(tableData.value.map(getRowKey))
    // Keep selections that belong to other pages
    const offPageSelected = selectedMultipleRows.value.filter((r) => !currentPageKeys.has(getRowKey(r)))
    const combined = [...offPageSelected]
    const combinedKeys = new Set(combined.map(getRowKey))
    rows.forEach((r) => {
      const key = getRowKey(r)
      if (!combinedKeys.has(key)) {
        combined.push(r)
        combinedKeys.add(key)
      }
    })
    selectedMultipleRows.value = combined
  }
}

function formatCellValue(row: any, prop: string): string {
  if (!row) return '-'
  if (prop === 'fotang_aktif_desc' && !row.fotang_aktif_desc && row.fotang_aktif) {
    return row.fotang_aktif
  }
  if (prop === 'fotang_chiutao_desc' && !row.fotang_chiutao_desc && row.fotang_chiutao) {
    return row.fotang_chiutao
  }
  const val = row[prop]
  if (val === undefined || val === null || val === '') return '-'
  return String(val)
}

// Data Fetching
async function fetchData() {
  if (!props.fetchApi) return

  if (abortController) abortController.abort()
  abortController = new AbortController()

  loading.value = true
  try {
    const params: Record<string, any> = {
      page: currentPage.value,
      limit: pageSizeInternal.value
    }

    // Append non-empty filter parameters
    Object.entries(filterValues.value).forEach(([k, v]) => {
      if (v !== undefined && v !== null && String(v).trim() !== '') {
        params[k] = typeof v === 'string' ? v.trim() : v
      }
    })

    const res = await props.fetchApi(params, abortController.signal)
    if (res) {
      tableData.value = Array.isArray(res) ? res : res.data || []
      total.value = res.meta?.total ?? res.total ?? tableData.value.length
    } else {
      tableData.value = []
      total.value = 0
    }

    if (props.multiple && selectedMultipleRows.value.length > 0) {
      await nextTick()
      if (tableRef.value) {
        const selectedKeys = new Set(selectedMultipleRows.value.map(getRowKey))
        tableData.value.forEach((row) => {
          if (selectedKeys.has(getRowKey(row))) {
            tableRef.value?.toggleRowSelection(row, true)
          }
        })
      }
    }
  } catch (err: any) {
    if (err.name === 'AbortError' || err.name === 'CanceledError') return
    console.error('Error fetching data in UmatPopupSelector:', err)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  fetchData()
}

function handleSizeChange(val: number) {
  pageSizeInternal.value = val
  currentPage.value = 1
  fetchData()
}

function handleCurrentChange(val: number) {
  currentPage.value = val
  fetchData()
}

function handleConfirm() {
  if (props.multiple) {
    if (selectedMultipleRows.value.length === 0) {
      ElMessage.warning('Pilih setidaknya satu data Umat')
      return
    }
    emit('select', selectedMultipleRows.value)
    emit('onselect', selectedMultipleRows.value)
  } else {
    if (!selectedSingleRow.value) {
      ElMessage.warning('Pilih salah satu data Umat terlebih dahulu')
      return
    }
    emit('select', selectedSingleRow.value)
    emit('onselect', selectedSingleRow.value)
  }
  closeDialog()
}

function closeDialog() {
  internalVisible.value = false
}

function onDialogOpen() {
  loadFotangOptions()
  currentPage.value = 1
  if (props.multiple) {
    if (props.selected && Array.isArray(props.selected) && props.selected.length > 0) {
      selectedMultipleRows.value = [...props.selected]
    } else {
      selectedMultipleRows.value = []
    }
    nextTick(() => {
      if (tableRef.value) {
        tableRef.value.clearSelection()
      }
    })
  } else {
    selectedSingleId.value = null
    selectedSingleRow.value = null
  }
  fetchData()
}

function onDialogClose() {
  if (abortController) abortController.abort()
}

watch(
  () => props.fotangOptions,
  (newOpts) => {
    if (newOpts && newOpts.length > 0) {
      internalFotangOptions.value = newOpts
    }
  },
  { immediate: true }
)

onUnmounted(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.filter-panel {
  background-color: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
  padding: 12px 14px 4px 14px;
  margin-bottom: 12px;
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  align-items: center;
}

.filter-item {
  margin-right: 0 !important;
  margin-bottom: 8px !important;
}

.filter-actions {
  margin-right: 0 !important;
  margin-bottom: 8px !important;
  display: flex;
  gap: 8px;
}

.action-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.left-action {
  display: flex;
  align-items: center;
  gap: 10px;
}

.selection-indicator {
  font-size: 13px;
  color: var(--el-color-primary);
  font-weight: 500;
}

.right-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.grid-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.single-radio :deep(.el-radio__label) {
  display: none;
}

.pagination-footer {
  margin-top: 12px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-table .el-table__row) {
  cursor: pointer;
}
</style>
