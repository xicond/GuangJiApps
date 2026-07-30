<template>
  <el-card shadow="never" class="form-card">
    <el-alert
      v-if="props.fieldErrors && Object.keys(props.fieldErrors).length > 0"
      type="error"
      show-icon
      title="Validasi Gagal"
      description="Terdapat kesalahan pengisian form pada beberapa kolom di bawah ini. Silahkan periksa pesan kesalahan berwarna merah."
      class="validation-alert mb-4"
    />

    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-position="top"
      class="kelas-form"
    >
      <el-row :gutter="20">
        <!-- Kelas Selection (Required) -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Kelas" prop="kode_kelas" :error="hasFieldError('kode_kelas') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.kode_kelas"
              placeholder="Pilih Kelas..."
              :fetch-api="kelasApi.getKelasLookup"
              :initial-option="formData.kelas_name"
            />
            <FieldErrors :errors="getFieldErrors('kode_kelas')" />
          </el-form-item>
        </el-col>

        <!-- Fotang Selection -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Fotang" prop="kode_fotang" :error="hasFieldError('kode_fotang') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.kode_fotang"
              placeholder="Pilih Fotang..."
              :fetch-api="fotangApi.getFotangLookup"
              :initial-option="formData.fotang_name"
            />
            <FieldErrors :errors="getFieldErrors('kode_fotang')" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <!-- Start Date -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Tanggal Mulai (Start Date)" prop="start_date" :error="hasFieldError('start_date') ? ' ' : undefined">
            <el-date-picker
              v-model="formData.start_date"
              type="date"
              placeholder="Pilih Tanggal Mulai"
              value-format="YYYY-MM-DD"
              style="width: 100%"
            />
            <FieldErrors :errors="getFieldErrors('start_date')" />
          </el-form-item>
        </el-col>

        <!-- End Date -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Tanggal Selesai (End Date)" prop="end_date" :error="hasFieldError('end_date') ? ' ' : undefined">
            <el-date-picker
              v-model="formData.end_date"
              type="date"
              placeholder="Pilih Tanggal Selesai"
              value-format="YYYY-MM-DD"
              style="width: 100%"
            />
            <FieldErrors :errors="getFieldErrors('end_date')" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <!-- Lokasi -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Lokasi Pelaksanaan" prop="lokasi" :error="hasFieldError('lokasi') ? ' ' : undefined">
            <el-input
              v-model="formData.lokasi"
              placeholder="Contoh: Gedung Utama Lt. 2"
              maxlength="50"
              show-word-limit
            />
            <FieldErrors :errors="getFieldErrors('lokasi')" />
          </el-form-item>
        </el-col>

        <!-- PIC -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="PIC / Penanggung Jawab" prop="pic" :error="hasFieldError('pic') ? ' ' : undefined">
            <el-input
              v-model="formData.pic"
              placeholder="Nama Penanggung Jawab"
              maxlength="100"
              show-word-limit
            />
            <FieldErrors :errors="getFieldErrors('pic')" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <!-- Level -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Tingkat / Level" prop="level" :error="hasFieldError('level') ? ' ' : undefined">
            <el-input
              v-model="formData.level"
              placeholder="Level kelas (max 3 karakter)"
              maxlength="3"
            />
            <FieldErrors :errors="getFieldErrors('level')" />
          </el-form-item>
        </el-col>

        <!-- Deadline -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Batas Pendaftaran (Deadline)" prop="deadline" :error="hasFieldError('deadline') ? ' ' : undefined">
            <el-date-picker
              v-model="formData.deadline"
              type="date"
              placeholder="Pilih Tanggal Deadline"
              value-format="YYYY-MM-DD"
              style="width: 100%"
            />
            <FieldErrors :errors="getFieldErrors('deadline')" />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- Keterangan -->
      <el-form-item label="Keterangan" prop="keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
        <el-input
          v-model="formData.keterangan"
          type="textarea"
          :rows="3"
          placeholder="Catatan tambahan mengenai kegiatan kelas..."
          maxlength="200"
          show-word-limit
        />
        <FieldErrors :errors="getFieldErrors('keterangan')" />
      </el-form-item>

      <!-- Optional MC Details Collapsible Section -->
      <el-collapse class="mc-collapse mb-4">
        <el-collapse-item name="1">
          <template #title>
            <div class="mc-collapse-title">
              <el-icon><User /></el-icon>
              <span>Daftar Pembawa Acara / MC (Opsional)</span>
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
        <el-button @click="handleCancel">Batal</el-button>
        <el-button
          type="primary"
          :loading="submitting"
          @click="handleSubmit"
        >
          {{ submitText }}
        </el-button>
      </div>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { User } from '@element-plus/icons-vue'
import { kelasApi } from '../../api/kelas'
import { fotangApi } from '../../api/fotang'
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

const formRules = reactive<FormRules>({
  kode_kelas: [
    { required: true, message: 'Silahkan pilih Kode Kelas', trigger: 'change' }
  ]
})

watch(
  () => props.initialData,
  (newData) => {
    if (newData && Object.keys(newData).length > 0) {
      Object.assign(formData, {
        kode_kelas: newData.kode_kelas || '',
        kode_fotang: newData.kode_fotang || '',
        start_date: newData.start_date || '',
        end_date: newData.end_date || '',
        lokasi: newData.lokasi || '',
        pic: newData.pic || '',
        level: newData.level || '',
        deadline: newData.deadline || '',
        keterangan: newData.keterangan || '',
        mc1: newData.mc1 || '',
        mc2: newData.mc2 || '',
        mc3: newData.mc3 || '',
        mc4: newData.mc4 || '',
        mc5: newData.mc5 || '',
        kelas_name: newData.kelas_name,
        fotang_name: newData.fotang_name
      })
    }
  },
  { immediate: true, deep: true }
)

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      const payload: Partial<Kelas> = { ...formData }
      // Clean temporary preloaded lookup object fields before submitting
      delete payload.kelas_name
      delete payload.fotang_name
      emit('submit', payload)
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
</style>
