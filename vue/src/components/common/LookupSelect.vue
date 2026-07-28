<template>
  <el-select
    :model-value="modelValue"
    :placeholder="placeholder"
    filterable
    remote
    :remote-method="handleRemoteSearch"
    :loading="loading"
    clearable
    class="w-full"
    @update:model-value="onValueChange"
    @clear="onClear"
    @visible-change="onVisibleChange"
  >
    <el-option
      v-for="item in options"
      :key="getOptionValue(item)"
      :label="getOptionLabel(item)"
      :value="getOptionValue(item)"
    />

    <template #footer>
      <div v-if="total > 0" class="lookup-pagination-footer">
        <el-button
          size="small"
          :disabled="page <= 1 || loading"
          @click.stop="prevPage"
        >
          &laquo; Prev
        </el-button>
        <span class="page-info">
          Halaman {{ page }} dari {{ maxPage }} (Total {{ total }})
        </span>
        <el-button
          size="small"
          :disabled="page >= maxPage || loading"
          @click.stop="nextPage"
        >
          Next &raquo;
        </el-button>
      </div>
    </template>
  </el-select>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { AppLookup, LookupListResponse, LookupQueryParams } from '../../types/lookup'

const props = withDefaults(
  defineProps<{
    modelValue?: string | number
    placeholder?: string
    fetchApi: (params: LookupQueryParams, signal?: AbortSignal) => Promise<LookupListResponse>
    valueKey?: keyof AppLookup | string
    labelKey?: keyof AppLookup | string
    clearable?: boolean
    pageSize?: number
  }>(),
  {
    modelValue: '',
    placeholder: 'Pilih...',
    valueKey: 'lookup_value',
    labelKey: 'lookup_description',
    clearable: true,
    pageSize: 10
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string | number | undefined): void
  (e: 'change', value: string | number | undefined): void
}>()

const options = ref<AppLookup[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)
const searchQuery = ref('')

const maxPage = computed(() => {
  if (total.value <= 0) return 1
  return Math.ceil(total.value / props.pageSize)
})

let debounceTimer: ReturnType<typeof setTimeout> | null = null
let currentAbortController: AbortController | null = null

function getOptionValue(item: AppLookup): string {
  const val = (item as any)[props.valueKey]
  return val !== undefined && val !== null ? val : item.lookup_value
}

function getOptionLabel(item: AppLookup): string {
  const label = (item as any)[props.labelKey]
  if (label !== undefined && label !== null && label !== '') return String(label)
  return item.lookup_description || item.lookup_value || item.lookup_id
}

async function loadData(targetPage = 1, query = '') {
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()

  loading.value = true
  page.value = targetPage
  searchQuery.value = query

  try {
    const res = await props.fetchApi(
      {
        page: targetPage,
        limit: props.pageSize,
        lookup_description: query.trim() || undefined
      },
      currentAbortController.signal
    )

    options.value = res.data || []
    total.value = res.meta?.total || options.value.length
  } catch (err: any) {
    if (err.name === 'AbortError' || err.name === 'CanceledError') return
    console.error('Error fetching lookup options:', err)
  } finally {
    loading.value = false
  }
}

function handleRemoteSearch(query: string) {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    loadData(1, query)
  }, 300)
}

function prevPage() {
  if (page.value > 1) {
    loadData(page.value - 1, searchQuery.value)
  }
}

function nextPage() {
  if (page.value < maxPage.value) {
    loadData(page.value + 1, searchQuery.value)
  }
}

function onValueChange(val: string | number | undefined) {
  emit('update:modelValue', val)
  emit('change', val)
}

function onClear() {
  onValueChange('')
  loadData(1, '')
}

function onVisibleChange(visible: boolean) {
  if (visible && options.value.length === 0) {
    loadData(1, '')
  }
}

onMounted(() => {
  loadData(1, '')
})
</script>

<style scoped>
.lookup-pagination-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-top: 1px solid var(--el-border-color-lighter, #e4e7ed);
  background-color: var(--el-fill-color-light, #f5f7fa);
  font-size: 12px;
  color: var(--el-text-color-regular, #606266);
}

.page-info {
  white-space: nowrap;
  font-size: 12px;
}
</style>
