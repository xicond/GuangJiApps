<template>
  <el-form
    ref="formRef"
    :model="formData"
    :rules="rules"
    label-position="top"
    size="default"
    class="umat-form"
    @submit.prevent="handleSubmit"
  >
    <!-- Section 1: Informasi Utama -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon><User /></el-icon>
          <span>Informasi Utama</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Kode Umat" prop="kode" required>
            <el-input v-model="formData.kode" placeholder="Masukkan kode umat (e.g. UMT-001)" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Indonesia" prop="nama_indonesia" required>
            <el-input v-model="formData.nama_indonesia" placeholder="Masukkan nama indonesia" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Alias" prop="alias">
            <el-input v-model="formData.alias" placeholder="Masukkan alias / panggilan" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Marga" prop="marga">
            <el-input v-model="formData.marga" placeholder="Masukkan marga" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Mandarin" prop="nama_mandarin">
            <el-input v-model="formData.nama_mandarin" placeholder="Masukkan nama mandarin" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Jenis Kelamin" prop="jenis_kelamin">
            <el-select v-model="formData.jenis_kelamin" placeholder="Pilih jenis kelamin" class="w-full" clearable>
              <el-option label="Pria (L)" value="001" />
              <el-option label="Wanita (P)" value="002" />
            </el-select>
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tempat Lahir" prop="tempat_lahir">
            <el-input v-model="formData.tempat_lahir" placeholder="Tempat lahir" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tanggal Lahir" prop="tanggal_lahir">
            <el-date-picker
              v-model="formData.tanggal_lahir"
              type="date"
              placeholder="Pilih tanggal lahir"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              class="w-full"
            />
          </el-form-item>
        </el-col>

        <!-- <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Status Record" prop="status">
            <el-switch
              v-model="formData.status"
              active-text="Aktif"
              inactive-text="Non-aktif"
            />
          </el-form-item>
        </el-col> -->
      </el-row>
    </el-card>

    <!-- Section 2: Kontak & Alamat -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon><Phone /></el-icon>
          <span>Kontak & Alamat</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="No. Telepon" prop="telepon">
            <el-input v-model="formData.telepon" placeholder="Nomor telepon rumah/kantor" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="No. HP / Mobile" prop="mobile">
            <el-input v-model="formData.mobile" placeholder="Nomor HP" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Email" prop="email">
            <el-input v-model="formData.email" placeholder="Alamat email" type="email" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-form-item label="Alamat Utama" prop="alamat">
            <el-input
              v-model="formData.alamat"
              type="textarea"
              :rows="2"
              placeholder="Alamat domisili utama"
            />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-form-item label="Alamat Tambahan" prop="alamat2">
            <el-input
              v-model="formData.alamat2"
              type="textarea"
              :rows="2"
              placeholder="Alamat kedua (opsional)"
            />
          </el-form-item>
        </el-col>
      </el-row>
    </el-card>

    <!-- Section 3: Informasi Ciu Tao -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon><Calendar /></el-icon>
          <span>Informasi Ciu Tao</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tahun Ciu Tao Mandarin" prop="tahun_chiutao_mandarin">
            <el-input v-model="formData.tahun_chiutao_mandarin" placeholder="Contoh: 2024" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tanggal Ciu Tao (Int)" prop="tanggal_chiutao_int">
            <el-date-picker
              v-model="formData.tanggal_chiutao_int"
              type="date"
              placeholder="Tanggal Ciu Tao International"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              class="w-full"
            />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tanggal Ciu Tao (Man)" prop="tanggal_chiutao_man">
            <el-input v-model="formData.tanggal_chiutao_man" placeholder="Tanggal Ciu Tao Mandarin" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Pengajak" prop="pengajak">
            <el-input v-model="formData.pengajak" placeholder="Nama pengajak" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Penanggung" prop="penanggung">
            <el-input v-model="formData.penanggung" placeholder="Nama penanggung" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Fotang Ciu Tao" prop="fotang_chiutao">
            <el-input v-model="formData.fotang_chiutao" placeholder="Fotang tempat ciu tao" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-card>

    <!-- Section 4: Catatan & Keterangan -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon><Notebook /></el-icon>
          <span>Catatan & Status Umat</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Status Umat" prop="status_umat">
            <el-input v-model="formData.status_umat" placeholder="Status umat" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tim Kerja" prop="tim_kerja">
            <el-input v-model="formData.tim_kerja" placeholder="Tim kerja" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Posisi" prop="posisi">
            <el-input v-model="formData.posisi" placeholder="Posisi" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Ching Khou" prop="ching_khou">
            <el-switch
              v-model="formData.ching_khou"
              active-text="Ya"
              inactive-text="Tidak"
            />
          </el-form-item>
        </el-col>

        <el-col :xs="24">
          <el-form-item label="Keterangan" prop="keterangan">
            <el-input
              v-model="formData.keterangan"
              type="textarea"
              :rows="3"
              placeholder="Keterangan tambahan..."
            />
          </el-form-item>
        </el-col>
      </el-row>
    </el-card>

    <!-- Form Action Buttons -->
    <div class="form-actions">
      <el-button size="large" @click="handleCancel">Batal</el-button>
      <el-button
        type="primary"
        size="large"
        :loading="submitting"
        :icon="Check"
        @click="handleSubmit"
      >
        {{ submitText }}
      </el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { User, Phone, Calendar, Notebook, Check } from '@element-plus/icons-vue'
import type { Umat } from '../../types/umat'

const props = withDefaults(
  defineProps<{
    initialData?: Partial<Umat>
    submitting?: boolean
    submitText?: string
  }>(),
  {
    initialData: () => ({}),
    submitting: false,
    submitText: 'Simpan Data'
  }
)

const emit = defineEmits<{
  (e: 'submit', payload: Partial<Umat>): void
  (e: 'cancel'): void
}>()

const formRef = ref<FormInstance>()

// Local reactive state for form fields
const formData = ref<Partial<Umat>>({
  kode: '',
  nama_indonesia: '',
  alias: '',
  marga: '',
  nama_mandarin: '',
  jenis_kelamin: '',
  tempat_lahir: '',
  tanggal_lahir: null,
  telepon: '',
  mobile: '',
  email: '',
  alamat: '',
  alamat2: '',
  tahun_chiutao_mandarin: '',
  tanggal_chiutao_int: null,
  tanggal_chiutao_man: '',
  pengajak: '',
  penanggung: '',
  fotang_chiutao: '',
  status_umat: '',
  tim_kerja: '',
  posisi: '',
  ching_khou: false,
  keterangan: '',
  status: true
})

// Watch props for initial data updates on edit page
watch(
  () => props.initialData,
  (val) => {
    if (val && Object.keys(val).length > 0) {
      formData.value = { ...formData.value, ...val }
    }
  },
  { immediate: true, deep: true }
)

// Form Validation Rules
const rules: FormRules = {
  kode: [
    { required: true, message: 'Kode umat wajib diisi', trigger: 'blur' },
    { min: 2, message: 'Kode minimal 2 karakter', trigger: 'blur' }
  ],
  nama_indonesia: [
    { required: true, message: 'Nama Indonesia wajib diisi', trigger: 'blur' }
  ]
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      emit('submit', { ...formData.value })
    }
  })
}

function handleCancel() {
  emit('cancel')
}
</script>

<style scoped>
.umat-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-section-card {
  border-radius: 10px;
  background-color: var(--el-bg-color-overlay);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  font-size: 1.05rem;
  color: var(--el-text-color-primary);
}

.w-full {
  width: 100%;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
}
</style>
