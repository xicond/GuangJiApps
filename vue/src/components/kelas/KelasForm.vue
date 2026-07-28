<template>
  <el-card shadow="never" class="form-card">
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
          <el-form-item label="Kelas" prop="kode_kelas">
            <LookupSelect
              v-model="formData.kode_kelas"
              placeholder="Pilih Kelas..."
              :fetch-api="kelasApi.getKelasLookup"
            />
          </el-form-item>
        </el-col>

        <!-- Fotang Selection -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Fotang" prop="kode_fotang">
            <LookupSelect
              v-model="formData.kode_fotang"
              placeholder="Pilih Fotang..."
              :fetch-api="fotangApi.getFotangLookup"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <!-- Start Date -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Tanggal Mulai (Start Date)" prop="start_date">
            <el-date-picker
              v-model="formData.start_date"
              type="date"
              placeholder="Pilih Tanggal Mulai"
              value-format="YYYY-MM-DD"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>

        <!-- End Date -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Tanggal Selesai (End Date)" prop="end_date">
            <el-date-picker
              v-model="formData.end_date"
              type="date"
              placeholder="Pilih Tanggal Selesai"
              value-format="YYYY-MM-DD"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <!-- Lokasi -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Lokasi Pelaksanaan" prop="lokasi">
            <el-input
              v-model="formData.lokasi"
              placeholder="Contoh: Gedung Utama Lt. 2"
              maxlength="50"
              show-word-limit
            />
          </el-form-item>
        </el-col>

        <!-- PIC -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="PIC / Penanggung Jawab" prop="pic">
            <el-input
              v-model="formData.pic"
              placeholder="Nama Penanggung Jawab"
              maxlength="100"
              show-word-limit
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <!-- Level -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Tingkat / Level" prop="level">
            <el-input
              v-model="formData.level"
              placeholder="Level kelas (max 3 karakter)"
              maxlength="3"
            />
          </el-form-item>
        </el-col>

        <!-- Deadline -->
        <el-col :xs="24" :sm="12">
          <el-form-item label="Batas Pendaftaran (Deadline)" prop="deadline">
            <el-date-picker
              v-model="formData.deadline"
              type="date"
              placeholder="Pilih Tanggal Deadline"
              value-format="YYYY-MM-DD"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- Keterangan -->
      <el-form-item label="Keterangan" prop="keterangan">
        <el-input
          v-model="formData.keterangan"
          type="textarea"
          :rows="3"
          placeholder="Catatan tambahan mengenai kegiatan kelas..."
          maxlength="200"
          show-word-limit
        />
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
              <el-form-item label="MC 1" prop="mc1">
                <el-input v-model="formData.mc1" placeholder="Nama MC 1" maxlength="100" />
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="MC 2" prop="mc2">
                <el-input v-model="formData.mc2" placeholder="Nama MC 2" maxlength="100" />
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="MC 3" prop="mc3">
                <el-input v-model="formData.mc3" placeholder="Nama MC 3" maxlength="100" />
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="MC 4" prop="mc4">
                <el-input v-model="formData.mc4" placeholder="Nama MC 4" maxlength="100" />
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="8">
              <el-form-item label="MC 5" prop="mc5">
                <el-input v-model="formData.mc5" placeholder="Nama MC 5" maxlength="100" />
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
import type { Kelas } from '../../types/kelas'

const props = withDefaults(
  defineProps<{
    initialData?: Partial<Kelas>
    submitting?: boolean
    submitText?: string
  }>(),
  {
    initialData: () => ({}),
    submitting: false,
    submitText: 'Simpan'
  }
)

const emit = defineEmits<{
  (e: 'submit', payload: Partial<Kelas>): void
  (e: 'cancel'): void
}>()

const formRef = ref<FormInstance>()

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
  mc5: ''
})

const formRules = reactive<FormRules>({
  kode_kelas: [
    { required: true, message: 'Silakan pilih Kode Kelas', trigger: 'change' }
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
        mc5: newData.mc5 || ''
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

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1rem;
}
</style>
