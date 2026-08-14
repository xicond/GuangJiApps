<template>
  <el-card shadow="never" class="form-card" v-loading="submitting">
    <el-alert v-if="props.fieldErrors && Object.keys(props.fieldErrors).length > 0" type="error" show-icon
      title="Invalid Inputs"
      description="Terdapat kesalahan pengisian form pada beberapa kolom di bawah ini. Silahkan periksa pesan kesalahan berwarna merah."
      class="validation-alert mb-4" />

    <el-form ref="formRef" :model="formData" :rules="formRules" label-position="top" class="kelas-form">
      <el-row :gutter="20">
        <!-- Kelas Selection (Required) -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Kelas" prop="kode_kelas" :error="hasFieldError('kode_kelas') ? ' ' : undefined">
            <LookupSelect v-model="formData.kode_kelas" placeholder="Pilih Kelas..."
              :fetch-api="kelasApi.getKelasLookup" :initial-option="formData.kelas_name" :clearable="false" />
            <FieldErrors :errors="getFieldErrors('kode_kelas')" />
          </el-form-item>
        </el-col>

        <!-- Fotang Selection -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Fotang" prop="kode_fotang" :error="hasFieldError('kode_fotang') ? ' ' : undefined">
            <LookupSelect v-model="formData.kode_fotang" placeholder="Pilih Fotang..."
              :fetch-api="fotangApi.getFotangLookup" :initial-option="formData.fotang_name" :clearable="false" />
            <FieldErrors :errors="getFieldErrors('kode_fotang')" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <!-- Tanggal Pelaksanaan Kelas (Date Range) -->
        <el-col :xs="24" :sm="24">
          <el-form-item label="Tanggal Pelaksanaan Kelas" prop="date_range"
            :error="hasFieldError('start_date') || hasFieldError('end_date') ? ' ' : undefined">
            <el-date-picker v-model="dateRange" type="daterange" range-separator="s/d" start-placeholder="Tanggal Mulai"
              end-placeholder="Tanggal Selesai" value-format="YYYY-MM-DD" style="width: 100%" :clearable="false"
              unlink-panels />
            <FieldErrors :errors="[...getFieldErrors('start_date'), ...getFieldErrors('end_date')]" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <!-- Level -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Tingkat / Level" prop="level" :error="hasFieldError('level') ? ' ' : undefined">
            <LookupSelect v-model="formData.level" placeholder="Pilih Level..."
              :fetch-api="lookupApi.getLookupKelasLevel" :clearable="false" />
            <FieldErrors :errors="getFieldErrors('level')" />
          </el-form-item>
        </el-col>

        <!-- Deadline -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Batas Pendaftaran (Deadline)" prop="deadline"
            :error="hasFieldError('deadline') ? ' ' : undefined">
            <el-date-picker v-model="formData.deadline" type="date" placeholder="Pilih Tanggal Deadline"
              value-format="YYYY-MM-DD" style="width: 100%" />
            <FieldErrors :errors="getFieldErrors('deadline')" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <!-- Lokasi -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Lokasi Pelaksanaan" prop="lokasi" :error="hasFieldError('lokasi') ? ' ' : undefined">
            <el-input v-model="formData.lokasi" placeholder="Contoh: Gedung Utama Lt. 2" maxlength="50"
              show-word-limit />
            <FieldErrors :errors="getFieldErrors('lokasi')" />
          </el-form-item>
        </el-col>

        <!-- PIC -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="PIC / Penanggung Jawab" prop="pic" :error="hasFieldError('pic') ? ' ' : undefined">
            <el-input v-model="formData.pic" placeholder="Nama Penanggung Jawab" maxlength="100" show-word-limit />
            <FieldErrors :errors="getFieldErrors('pic')" />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- Keterangan -->
      <el-form-item label="Keterangan" prop="keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
        <el-input v-model="formData.keterangan" type="textarea" :rows="3"
          placeholder="Catatan tambahan mengenai kegiatan kelas..." maxlength="200" show-word-limit />
        <FieldErrors :errors="getFieldErrors('keterangan')" />
      </el-form-item>

      <!-- Optional MC Details Collapsible Section -->
      <el-collapse class="mc-collapse mb-4">
        <el-collapse-item name="1">
          <template #title>
            <div class="mc-collapse-title">
              <el-icon>
                <User />
              </el-icon>
              <span>{{ isMobile ? 'Detail MC' : 'Daftar Pembawa Acara / MC (Opsional)' }}</span>
            </div>
          </template>

          <el-row :gutter="16">
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="MC 1" prop="mc1" :error="hasFieldError('mc1') ? ' ' : undefined">
                <el-input v-model="formData.mc1" placeholder="Nama MC 1" maxlength="100" />
                <FieldErrors :errors="getFieldErrors('mc1')" />
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="MC 2" prop="mc2" :error="hasFieldError('mc2') ? ' ' : undefined">
                <el-input v-model="formData.mc2" placeholder="Nama MC 2" maxlength="100" />
                <FieldErrors :errors="getFieldErrors('mc2')" />
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="MC 3" prop="mc3" :error="hasFieldError('mc3') ? ' ' : undefined">
                <el-input v-model="formData.mc3" placeholder="Nama MC 3" maxlength="100" />
                <FieldErrors :errors="getFieldErrors('mc3')" />
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="MC 4" prop="mc4" :error="hasFieldError('mc4') ? ' ' : undefined">
                <el-input v-model="formData.mc4" placeholder="Nama MC 4" maxlength="100" />
                <FieldErrors :errors="getFieldErrors('mc4')" />
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="MC 5" prop="mc5" :error="hasFieldError('mc5') ? ' ' : undefined">
                <el-input v-model="formData.mc5" placeholder="Nama MC 5" maxlength="100" />
                <FieldErrors :errors="getFieldErrors('mc5')" />
              </el-form-item>
            </el-col>
          </el-row>
        </el-collapse-item>
      </el-collapse>

      <!-- Form Actions -->
      <div class="form-actions">
        <el-button :disabled="submitting" @click="handleCancel">Batal</el-button>
        <el-button type="primary" :loading="submitting" :disabled="submitting" @click="handleSubmit">
          {{ submitText }}
        </el-button>
      </div>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import type { FormInstance, FormRules } from 'element-plus'
import { User } from '@element-plus/icons-vue'
import { kelasApi } from '../../api/kelas'
import { fotangApi } from '../../api/fotang'
import lookupApi from '../../api/lookup'
import LookupSelect from '../common/LookupSelect.vue'
import FieldErrors from '../common/FieldErrors.vue'
import type { Kelas } from '../../types/kelas'

const props = withDefaults(
  defineProps<{
    initialData?: Partial<Kelas>
    submitting?: boolean
    submitText?: string
    fieldErrors?: Record<string, string[]>
  }>(),
  {
    initialData: () => ({}),
    submitting: false,
    submitText: 'Simpan',
    fieldErrors: () => ({})
  }
)

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
// const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

const emit = defineEmits<{
  (e: 'submit', payload: Partial<Kelas>): void
  (e: 'cancel'): void
}>()

const formRef = ref<FormInstance>()

function getFieldErrors(fieldName: string): string[] {
  if (!props.fieldErrors) return []
  return props.fieldErrors[fieldName] || []
}

function hasFieldError(fieldName: string): boolean {
  return getFieldErrors(fieldName).length > 0
}

const formData = reactive<Partial<Kelas>>({
  kode_kelas: '',
  kode_fotang: '',
  start_date: '',
  end_date: '',
  lokasi: '',
  pic: '',
  level: '',
  deadline: '',
  keterangan: '',
  mc1: '',
  mc2: '',
  mc3: '',
  mc4: '',
  mc5: '',
  kelas_name: undefined,
  fotang_name: undefined
})

const dateRange = computed<[string, string] | null>({
  get() {
    if (formData.start_date || formData.end_date) {
      return [formData.start_date || '', formData.end_date || '']
    }
    return null
  },
  set(val) {
    if (val && val.length === 2) {
      formData.start_date = val[0] || ''
      formData.end_date = val[1] || ''
    } else {
      formData.start_date = ''
      formData.end_date = ''
    }
  }
})

const formRules = reactive<FormRules>({
  kode_kelas: [
    { required: true, message: 'Silahkan pilih Kode Kelas', trigger: 'change' }
  ],
  kode_fotang: [
    { required: true, message: 'Silahkan pilih Fotang', trigger: 'change' }
  ],
  date_range: [
    { required: true, message: 'Silahkan pilih Tanggal Mulai', trigger: 'change' }
  ],
  start_date: [
    { required: true, message: 'Silahkan pilih Tanggal Mulai', trigger: 'change' }
  ],
  end_date: [
    { required: true, message: 'Silahkan pilih Tanggal Selesai', trigger: 'change' }
  ],
  level: [
    { required: true, message: 'Silahkan pilih Level', trigger: 'change' }
  ]
})

const TRIM_FIELDS: (keyof Kelas)[] = [
  'kode_kelas',
  'kode_fotang',
  'level',
  'lokasi',
  'pic',
  'keterangan',
  'mc1',
  'mc2',
  'mc3',
  'mc4',
  'mc5'
]

function trimTargetFields(data: Partial<Kelas>): Partial<Kelas> {
  const trimmed: Record<string, unknown> = { ...data }
  for (const field of TRIM_FIELDS) {
    if (typeof trimmed[field] === 'string') {
      trimmed[field] = (trimmed[field] as string).trim()
    }
  }
  return trimmed as Partial<Kelas>
}

watch(
  () => props.initialData,
  (newData) => {
    if (newData && Object.keys(newData).length > 0) {
      const trimmedData = trimTargetFields(newData)
      Object.assign(formData, {
        kode_kelas: trimmedData.kode_kelas || '',
        kode_fotang: trimmedData.kode_fotang || '',
        start_date: trimmedData.start_date || '',
        end_date: trimmedData.end_date || '',
        lokasi: trimmedData.lokasi || '',
        pic: trimmedData.pic || '',
        level: trimmedData.level || '',
        deadline: trimmedData.deadline || '',
        keterangan: trimmedData.keterangan || '',
        mc1: trimmedData.mc1 || '',
        mc2: trimmedData.mc2 || '',
        mc3: trimmedData.mc3 || '',
        mc4: trimmedData.mc4 || '',
        mc5: trimmedData.mc5 || '',
        kelas_name: trimmedData.kelas_name,
        fotang_name: trimmedData.fotang_name
      })
    }
  },
  { immediate: true, deep: true }
)

import { scrollToFormError } from '../../utils/scroll'

watch(
  () => props.fieldErrors,
  (newErrors) => {
    if (newErrors && Object.keys(newErrors).length > 0) {
      scrollToFormError()
    }
  },
  { immediate: true, deep: true }
)

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      const payload: Partial<Kelas> = trimTargetFields({ ...formData })
      // Clean temporary preloaded lookup object fields before submitting
      delete payload.kelas_name
      delete payload.fotang_name
      emit('submit', payload)
    } else {
      scrollToFormError()
    }
  })
}

function handleCancel() {
  emit('cancel')
}
</script>

<style scoped>
.form-card {
  border-radius: 10px;
  background-color: var(--el-bg-color-overlay);
}

.kelas-form {
  padding-top: 0.5rem;
}

.mc-collapse {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 1.5rem;
}

:deep(.mc-collapse .el-collapse-item__header) {
  padding: 0 1rem;
  min-height: 48px;
  background-color: var(--el-bg-color-overlay);
}

:deep(.mc-collapse .el-collapse-item__content) {
  padding: 0 1.25rem 0 1.25rem;
}

.mc-collapse-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  color: var(--el-text-color-regular);
}

.mb-4 {
  margin-bottom: 1rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}
</style>
