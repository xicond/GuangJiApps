<template>
  <el-form
    ref="formRef"
    :model="formData"
    :rules="rules"
    label-position="top"
    size="default"
    class="umat-form"
    v-loading="submitting"
    @submit.prevent="handleSubmit"
  >
    <el-alert
      v-if="props.fieldErrors && Object.keys(props.fieldErrors).length > 0"
      type="error"
      show-icon
      title="Invalid Inputs"
      description="Terdapat kesalahan pengisian form pada beberapa kolom di bawah ini. Silahkan periksa pesan kesalahan berwarna merah."
      class="validation-alert mb-4"
    />

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
          <el-form-item label="Kode Umat" prop="kode" required :error="hasFieldError('kode') ? ' ' : undefined">
            <el-input v-model="formData.kode" placeholder="Masukkan kode umat (e.g. UMT-001)" />
            <FieldErrors :errors="getFieldErrors('kode')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Indonesia" prop="nama_indonesia" required :error="hasFieldError('nama_indonesia') ? ' ' : undefined">
            <el-input v-model="formData.nama_indonesia" placeholder="Masukkan nama indonesia" />
            <FieldErrors :errors="getFieldErrors('nama_indonesia')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Alias" prop="alias" :error="hasFieldError('alias') ? ' ' : undefined">
            <el-input v-model="formData.alias" placeholder="Masukkan alias / panggilan" />
            <FieldErrors :errors="getFieldErrors('alias')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Marga" prop="marga" :error="hasFieldError('marga') ? ' ' : undefined">
            <el-input v-model="formData.marga" placeholder="Masukkan marga" />
            <FieldErrors :errors="getFieldErrors('marga')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Mandarin" prop="nama_mandarin" :error="hasFieldError('nama_mandarin') ? ' ' : undefined">
            <el-input v-model="formData.nama_mandarin" placeholder="Masukkan nama mandarin" />
            <FieldErrors :errors="getFieldErrors('nama_mandarin')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Jenis Kelamin" prop="jenis_kelamin" :error="hasFieldError('jenis_kelamin') ? ' ' : undefined">
            <el-select v-model="formData.jenis_kelamin" placeholder="Pilih jenis kelamin" class="w-full" clearable>
              <el-option label="乾 PRIA" value="001" />
              <el-option label="坤 WANITA" value="002" />
              <el-option label="童 ANAK PRIA" value="003" />
              <el-option label="女 ANAK WANITA" value="004" />
            </el-select>
            <FieldErrors :errors="getFieldErrors('jenis_kelamin')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tempat Lahir" prop="tempat_lahir" :error="hasFieldError('tempat_lahir') ? ' ' : undefined">
            <el-input v-model="formData.tempat_lahir" placeholder="Tempat lahir" />
            <FieldErrors :errors="getFieldErrors('tempat_lahir')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tanggal Lahir" prop="tanggal_lahir" :error="hasFieldError('tanggal_lahir') ? ' ' : undefined">
            <el-date-picker
              v-model="formData.tanggal_lahir"
              type="date"
              placeholder="Pilih tanggal lahir"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              class="w-full"
            />
            <FieldErrors :errors="getFieldErrors('tanggal_lahir')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Pendidikan" prop="pendidikan" :error="hasFieldError('pendidikan') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.pendidikan"
              placeholder="Pilih pendidikan"
              :fetch-api="lookupApi.getLookupPendidikan"
            />
            <FieldErrors :errors="getFieldErrors('pendidikan')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Pekerjaan" prop="pekerjaan" :error="hasFieldError('pekerjaan') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.pekerjaan"
              placeholder="Pilih pekerjaan"
              :fetch-api="lookupApi.getLookupPekerjaan"
            />
            <FieldErrors :errors="getFieldErrors('pekerjaan')" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-card>

    <!-- Section 2: Kontak, Alamat & Fotang -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon><Phone /></el-icon>
          <span>Kontak & Fotang Aktif</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="No. Telepon" prop="telepon" :error="hasFieldError('telepon') ? ' ' : undefined">
            <el-input v-model="formData.telepon" placeholder="Nomor telepon rumah/kantor" />
            <FieldErrors :errors="getFieldErrors('telepon')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="No. HP / Mobile" prop="mobile" :error="hasFieldError('mobile') ? ' ' : undefined">
            <el-input v-model="formData.mobile" placeholder="Nomor HP" />
            <FieldErrors :errors="getFieldErrors('mobile')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Email" prop="email" :error="hasFieldError('email') ? ' ' : undefined">
            <el-input v-model="formData.email" placeholder="Alamat email" type="email" />
            <FieldErrors :errors="getFieldErrors('email')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-form-item label="Alamat Utama" prop="alamat" :error="hasFieldError('alamat') ? ' ' : undefined">
            <el-input
              v-model="formData.alamat"
              type="textarea"
              :rows="2"
              placeholder="Alamat domisili utama"
            />
            <FieldErrors :errors="getFieldErrors('alamat')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-form-item label="Alamat Tambahan" prop="alamat2" :error="hasFieldError('alamat2') ? ' ' : undefined">
            <el-input
              v-model="formData.alamat2"
              type="textarea"
              :rows="2"
              placeholder="Alamat kedua (opsional)"
            />
            <FieldErrors :errors="getFieldErrors('alamat2')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Fotang Aktif" prop="fotang_aktif" :error="hasFieldError('fotang_aktif') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.fotang_aktif"
              placeholder="Pilih fotang aktif"
              :fetch-api="fotangApi.getFotangLookup"
            />
            <FieldErrors :errors="getFieldErrors('fotang_aktif')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Fotang Lain" prop="nama_fotang_lain" :error="hasFieldError('nama_fotang_lain') ? ' ' : undefined">
            <el-input v-model="formData.nama_fotang_lain" placeholder="Nama fotang lain (opsional)" />
            <FieldErrors :errors="getFieldErrors('nama_fotang_lain')" />
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
          <el-form-item label="Tanggal Ciu Tao (Int)" prop="tanggal_chiutao_int" :error="hasFieldError('tanggal_chiutao_int') ? ' ' : undefined">
            <el-date-picker
              v-model="formData.tanggal_chiutao_int"
              type="date"
              placeholder="Tanggal Ciu Tao International"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              class="w-full"
            />
            <FieldErrors :errors="getFieldErrors('tanggal_chiutao_int')" />
          </el-form-item>
        </el-col>

        <!-- <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tanggal Ciu Tao (Man)" prop="tanggal_chiutao_man" :error="hasFieldError('tanggal_chiutao_man') ? ' ' : undefined">
            <el-input v-model="formData.tanggal_chiutao_man" placeholder="Tanggal Ciu Tao Mandarin" />
            <FieldErrors :errors="getFieldErrors('tanggal_chiutao_man')" />
          </el-form-item>
        </el-col> -->

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Waktu Ciu Tao" prop="waktu_chiutao_mandarin" :error="hasFieldError('waktu_chiutao_mandarin') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.waktu_chiutao_mandarin"
              placeholder="Pilih waktu ciu tao"
              :fetch-api="lookupApi.getLookupWaktuCiuTao"
            />
            <FieldErrors :errors="getFieldErrors('waktu_chiutao_mandarin')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Fotang Ciu Tao" prop="fotang_chiutao" :error="hasFieldError('fotang_chiutao') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.fotang_chiutao"
              placeholder="Pilih fotang ciu tao"
              :fetch-api="fotangApi.getFotangLookup"
            />
            <FieldErrors :errors="getFieldErrors('fotang_chiutao')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tien Chuan Se (TCS)" prop="tcs" :error="hasFieldError('tcs') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.tcs"
              placeholder="Pilih TCS"
              :fetch-api="lookupApi.getLookupTcs"
            />
            <FieldErrors :errors="getFieldErrors('tcs')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama TCS Lain" prop="nama_tcs_lain" :error="hasFieldError('nama_tcs_lain') ? ' ' : undefined">
            <el-input v-model="formData.nama_tcs_lain" placeholder="Nama TCS lain (opsional)" />
            <FieldErrors :errors="getFieldErrors('nama_tcs_lain')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Uang Pahala" prop="uang_pahala" :error="hasFieldError('uang_pahala') ? ' ' : undefined">
            <el-input-number
              v-model="formData.uang_pahala"
              :min="0"
              :step="10000"
              placeholder="0"
              class="w-full"
            />
            <FieldErrors :errors="getFieldErrors('uang_pahala')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Pengajak" prop="pengajak" :error="hasFieldError('pengajak') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.pengajak"
              placeholder="Cari & pilih pengajak"
              :fetch-api="umatApi.getUmats"
              value-key="kode"
              :label-formatter="formatUmatLabel"
            />
            <FieldErrors :errors="getFieldErrors('pengajak')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Penanggung" prop="penanggung" :error="hasFieldError('penanggung') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.penanggung"
              placeholder="Cari & pilih penanggung"
              :fetch-api="umatApi.getUmats"
              value-key="kode"
              :label-formatter="formatUmatLabel"
            />
            <FieldErrors :errors="getFieldErrors('penanggung')" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-card>

    <!-- Section 4: Kelas & Sidang Dharma (Sd2 / Sd3) -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon><Notebook /></el-icon>
          <span>Kelas & Tempat Sidang Dharma</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Kelas Umum" prop="kelas_umum" :error="hasFieldError('kelas_umum') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.kelas_umum"
              placeholder="Pilih kelas umum"
              :fetch-api="lookupApi.getLookupKelasUmum"
            />
            <FieldErrors :errors="getFieldErrors('kelas_umum')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Kelas Khusus" prop="kelas_khusus" :error="hasFieldError('kelas_khusus') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.kelas_khusus"
              placeholder="Pilih kelas khusus"
              :fetch-api="lookupApi.getLookupKelas"
            />
            <FieldErrors :errors="getFieldErrors('kelas_khusus')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tempat Sidang Dharma 2 (Sd2)" prop="tempat_sd2" :error="hasFieldError('tempat_sd2') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.tempat_sd2"
              placeholder="Pilih tempat Sd2"
              :fetch-api="fotangApi.getFotangLookup"
            />
            <FieldErrors :errors="getFieldErrors('tempat_sd2')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tempat Sidang Dharma 3 (Sd3)" prop="tempat_sd3" :error="hasFieldError('tempat_sd3') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.tempat_sd3"
              placeholder="Pilih tempat Sd3"
              :fetch-api="fotangApi.getFotangLookup"
            />
            <FieldErrors :errors="getFieldErrors('tempat_sd3')" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-card>

    <!-- Section 5: Status Umat, Ikrar & Tingkatan -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon><Notebook /></el-icon>
          <span>Status Umat, Ikrar & Tingkatan</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tim Kerja" prop="tim_kerja" :error="hasFieldError('tim_kerja') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.tim_kerja"
              placeholder="Pilih tim kerja"
              :fetch-api="lookupApi.getLookupTimKerja"
            />
            <FieldErrors :errors="getFieldErrors('tim_kerja')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Status Umat" prop="status_umat" :error="hasFieldError('status_umat') ? ' ' : undefined">
            <LookupSelect
              v-model="formData.status_umat"
              placeholder="Pilih status umat"
              :fetch-api="lookupApi.getLookupStatus"
            />
            <FieldErrors :errors="getFieldErrors('status_umat')" />
          </el-form-item>
        </el-col>

        <!-- Ikrar Checkboxes -->
        <el-col :xs="24">
          <el-form-item label="Ikrar" prop="ikrar">
            <div class="ikrar-checkbox-group">
              <el-checkbox v-model="formData.ikrar_1" label="Ikrar 1" />
              <el-checkbox v-model="formData.ikrar_2" label="Ikrar 2" />
              <el-checkbox v-model="formData.ikrar_3" label="Ikrar 3" />
              <el-checkbox v-model="formData.ikrar_4" label="Ikrar 4" />
              <el-checkbox v-model="formData.ikrar_5" label="Ikrar 5" />
              <el-checkbox v-model="formData.ikrar_6" label="Ikrar 6" />
            </div>
          </el-form-item>
        </el-col>

        <!-- Ren Chai Pan -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Ren Chai Pan" prop="ren_chai_pan" :error="hasFieldError('ren_chai_pan') || hasFieldError('tanggal_ren_chai_pan') ? ' ' : undefined">
            <div class="status-date-container">
              <el-checkbox v-model="formData.ren_chai_pan" />
              <el-date-picker
                v-model="formData.tanggal_ren_chai_pan"
                type="date"
                placeholder="Tanggal Ren Chai Pan"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                class="w-full"
                :disabled="!formData.ren_chai_pan"
              />
            </div>
            <FieldErrors :errors="getFieldErrors('ren_chai_pan').concat(getFieldErrors('tanggal_ren_chai_pan'))" />
          </el-form-item>
        </el-col>

        <!-- Lien Ciang Pan -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Lien Ciang Pan" prop="lien_ciang_pan" :error="hasFieldError('lien_ciang_pan') || hasFieldError('tanggal_lien_ciang_pan') ? ' ' : undefined">
            <div class="status-date-container">
              <el-checkbox v-model="formData.lien_ciang_pan" />
              <el-date-picker
                v-model="formData.tanggal_lien_ciang_pan"
                type="date"
                placeholder="Tanggal Lien Ciang Pan"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                class="w-full"
                :disabled="!formData.lien_ciang_pan"
              />
            </div>
            <FieldErrors :errors="getFieldErrors('lien_ciang_pan').concat(getFieldErrors('tanggal_lien_ciang_pan'))" />
          </el-form-item>
        </el-col>

        <!-- Ciang Yen Pan -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Ciang Yen Pan" prop="ciang_yen_pan" :error="hasFieldError('ciang_yen_pan') || hasFieldError('tanggal_ciang_yen_pan') ? ' ' : undefined">
            <div class="status-date-container">
              <el-checkbox v-model="formData.ciang_yen_pan" />
              <el-date-picker
                v-model="formData.tanggal_ciang_yen_pan"
                type="date"
                placeholder="Tanggal Ciang Yen Pan"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                class="w-full"
                :disabled="!formData.ciang_yen_pan"
              />
            </div>
            <FieldErrors :errors="getFieldErrors('ciang_yen_pan').concat(getFieldErrors('tanggal_ciang_yen_pan'))" />
          </el-form-item>
        </el-col>

        <!-- Ching Khou -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Ching Khou" prop="ching_khou" :error="hasFieldError('ching_khou') || hasFieldError('tanggal_ching_khou') ? ' ' : undefined">
            <div class="status-date-container">
              <el-checkbox v-model="formData.ching_khou" />
              <el-date-picker
                v-model="formData.tanggal_ching_khou"
                type="date"
                placeholder="Tanggal Ching Khou"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                class="w-full"
                :disabled="!formData.ching_khou"
              />
            </div>
            <FieldErrors :errors="getFieldErrors('ching_khou').concat(getFieldErrors('tanggal_ching_khou'))" />
          </el-form-item>
        </el-col>

        <!-- Meninggal -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Sudah Meninggal" prop="meninggal" :error="hasFieldError('meninggal') || hasFieldError('tanggal_meninggal') ? ' ' : undefined">
            <div class="status-date-container">
              <el-checkbox v-model="formData.meninggal" />
              <el-date-picker
                v-model="formData.tanggal_meninggal"
                type="date"
                placeholder="Tanggal Meninggal"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                class="w-full"
                :disabled="!formData.meninggal"
              />
            </div>
            <FieldErrors :errors="getFieldErrors('meninggal').concat(getFieldErrors('tanggal_meninggal'))" />
          </el-form-item>
        </el-col>

        <el-col :xs="24">
          <el-form-item label="Keterangan" prop="keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
            <el-input
              v-model="formData.keterangan"
              type="textarea"
              :rows="3"
              placeholder="Keterangan tambahan..."
            />
            <FieldErrors :errors="getFieldErrors('keterangan')" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-card>

    <!-- Form Action Buttons -->
    <div class="form-actions">
      <el-button size="large" :disabled="submitting" @click="handleCancel">Batal</el-button>
      <el-button
        type="primary"
        size="large"
        :loading="submitting"
        :disabled="submitting"
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
import LookupSelect from '../common/LookupSelect.vue'
import FieldErrors from '../common/FieldErrors.vue'
import lookupApi from '../../api/lookup'
import fotangApi from '../../api/fotang'
import umatApi from '../../api/umat'

const props = withDefaults(
  defineProps<{
    initialData?: Partial<Umat>
    submitting?: boolean
    submitText?: string
    fieldErrors?: Record<string, string[]>
  }>(),
  {
    initialData: () => ({}),
    submitting: false,
    submitText: 'Simpan Data',
    fieldErrors: () => ({})
  }
)

const emit = defineEmits<{
  (e: 'submit', payload: Partial<Umat>): void
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
  pendidikan: '',
  pekerjaan: '',
  telepon: '',
  mobile: '',
  email: '',
  alamat: '',
  alamat2: '',
  fotang_aktif: '',
  nama_fotang_lain: '',
  tahun_chiutao_mandarin: '',
  waktu_chiutao_mandarin: '',
  tanggal_chiutao_int: null,
  tanggal_chiutao_man: '',
  pengajak: '',
  penanggung: '',
  tcs: '',
  nama_tcs_lain: '',
  uang_pahala: 0,
  fotang_chiutao: '',
  tempat_sd2: '',
  tempat_sd3: '',
  kelas_umum: '',
  kelas_khusus: '',
  status_umat: '',
  tim_kerja: '',
  posisi: '',
  ikrar_1: false,
  ikrar_2: false,
  ikrar_3: false,
  ikrar_4: false,
  ikrar_5: false,
  ikrar_6: false,
  ren_chai_pan: false,
  tanggal_ren_chai_pan: null,
  lien_ciang_pan: false,
  tanggal_lien_ciang_pan: null,
  ciang_yen_pan: false,
  tanggal_ciang_yen_pan: null,
  ching_khou: false,
  tanggal_ching_khou: null,
  meninggal: false,
  tanggal_meninggal: null,
  keterangan: '',
  status: true
})

function formatUmatLabel(item: any): string {
  if (!item) return ''
  const parts = []
  if (item.nama_indonesia) parts.push(item.nama_indonesia)
  if (item.alias) parts.push(item.alias)
  if (item.nama_mandarin) parts.push(item.nama_mandarin)
  const formatted = parts.join(' / ')
  return formatted ? `${formatted} (${item.kode || item.id || ''})` : item.kode || String(item.id || '')
}

const TRIM_FIELDS: (keyof Umat)[] = [
  'kelas_khusus',
  'kelas_umum',
  'fotang_chiutao',
  'fotang_aktif',
  'tcs',
  'waktu_chiutao_mandarin',
  'pendidikan',
  'pekerjaan',
  'jenis_kelamin',
  'tempat_sd2',
  'tempat_sd3',
  'tim_kerja',
  'posisi',
  'status_umat'
]

function trimTargetFields(data: Partial<Umat>): Partial<Umat> {
  const trimmed: Record<string, unknown> = { ...data }
  for (const field of TRIM_FIELDS) {
    if (typeof trimmed[field] === 'string') {
      trimmed[field] = (trimmed[field] as string).trim()
    }
  }
  return trimmed as Partial<Umat>
}

// Watch props for initial data updates on edit page
watch(
  () => props.initialData,
  (val) => {
    if (val && Object.keys(val).length > 0) {
      const trimmedVal = trimTargetFields(val)
      formData.value = { ...formData.value, ...trimmedVal }
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
      const payload = trimTargetFields(formData.value)
      emit('submit', payload)
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

.mb-4 {
  margin-bottom: 1rem;
}

.ikrar-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 1.25rem;
  padding: 0.25rem 0;
}

.status-date-container {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
}
</style>
