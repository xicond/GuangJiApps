<template>
  <el-select
    :model-value="modelValue"
    :placeholder="placeholder"
    filterable
    remote
    :remote-method="handleRemoteSearch"
    :loading="loading"
    :clearable="props.clearable"
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

    <template v-if="maxPage>1" #footer>
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
import { ref, computed, watch, onMounted } from 'vue'
import type { AppLookup, LookupQueryParams } from '../../types/lookup'
import { cachedFetchLookup } from '../../utils/lookupCache'

const props = withDefaults(
  defineProps<{
    modelValue?: string | number
    placeholder?: string
    fetchApi: (params: LookupQueryParams | any, signal?: AbortSignal) => Promise<any>
    valueKey?: string
    labelKey?: string
    labelFormatter?: (item: any) => string
    clearable?: boolean
    pageSize?: number
    initialOption?: any
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

const options = ref<any[]>([])
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

function getOptionValue(item: any): string | number {
  if (!item) return ''
  const val = item[props.valueKey]
  return val !== undefined && val !== null ? val : item.lookup_value || item.id || ''
}

function getOptionLabel(item: any): string {
  if (!item) return ''
  if (props.labelFormatter) {
    return props.labelFormatter(item)
  }
  const label = item[props.labelKey]
  if (label !== undefined && label !== null && label !== '') return String(label)
  return item.lookup_description || item.lookup_value || item.lookup_id || item.nama_indonesia || ''
}

function ensureInitialOption(opt?: AppLookup) {
  if (!opt) return
  const val = getOptionValue(opt)
  if (val === undefined || val === null || val === '') return
  const exists = options.value.some((item) => getOptionValue(item) === val)
  if (!exists) {
    options.value = [opt, ...options.value]
  }
}

watch(
  () => props.initialOption,
  (newOpt) => {
    if (newOpt) {
      ensureInitialOption(newOpt)
    }
  },
  { immediate: true, deep: true }
)

async function loadData(targetPage = 1, query = '', forceRefresh = false) {
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()

  loading.value = true
  page.value = targetPage
  searchQuery.value = query

  try {
    const params = {
      page: targetPage,
      limit: props.pageSize,
      lookup_description: query.trim() || undefined
    }
    const res = await cachedFetchLookup(
      props.fetchApi,
      params,
      currentAbortController.signal,
      forceRefresh
    )

    options.value = res.data || []
    if (props.initialOption) {
      ensureInitialOption(props.initialOption)
    }
    await checkAndFetchMissingSelectedValue()
    total.value = res.meta?.total || options.value.length
  } catch (err: any) {
    if (err.name === 'AbortError' || err.name === 'CanceledError') return
    console.error('Error fetching lookup options:', err)
  } finally {
    loading.value = false
  }
}

async function checkAndFetchMissingSelectedValue() {
  const currentVal = props.modelValue
  if (currentVal === undefined || currentVal === null || currentVal === '') return
  const exists = options.value.some((item) => String(getOptionValue(item)) === String(currentVal))
  if (!exists && props.fetchApi) {
    try {
      const resVal = await props.fetchApi({ lookup_value: String(currentVal), limit: 1 })
      if (resVal?.data && resVal.data.length > 0) {
        ensureInitialOption(resVal.data[0])
        return
      }
      const resId = await props.fetchApi({ lookup_id: String(currentVal), limit: 1 })
      if (resId?.data && resId.data.length > 0) {
        ensureInitialOption(resId.data[0])
        return
      }
    } catch (e) {
      // Ignore lookup error for missing selected item
    }
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
  loadData(1, '', true)
}

function onVisibleChange(visible: boolean) {
  if (visible && options.value.length === 0) {
    loadData(1, '')
  }
}

watch(
  () => props.fetchApi,
  () => {
    loadData(1, '', true)
  }
)

watch(
  () => props.modelValue,
  async (newVal) => {
    if (newVal !== undefined && newVal !== null && newVal !== '') {
      await checkAndFetchMissingSelectedValue()
    }
  },
  { immediate: true }
)

onMounted(() => {
  if (
    (props.modelValue !== undefined && props.modelValue !== null && props.modelValue !== '') ||
    props.initialOption
  ) {
    loadData(1, '')
  }
})

defineExpose({
  loadData: (targetPage = 1, query = '', forceRefresh = true) => loadData(targetPage, query, forceRefresh),
  reload: () => loadData(1, '', true)
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
