<template>
  <div class="create-container">
    <!-- Header Section -->
    <div class="page-header">
      <div class="header-top">
        <el-button :icon="Back" text @click="handleBack">Kembali ke Daftar</el-button>
        <div class="header-actions">
          <input ref="ocrFileInput" type="file" accept="image/*" style="display: none"
            @change="handleOcrFileSelected" />
          <el-button type="primary" plain :icon="Picture" :loading="ocrLoading"
            :disabled="submitting || ocrLoading || ocrRateLimitCountdown > 0" @click="triggerOcrInput">
            {{ ocrRateLimitCountdown > 0 ? `OCR Retry-After (${ocrRateLimitCountdown}s)` : 'OCR Recognition' }}
          </el-button>
        </div>
      </div>
      <h2 class="page-title">Tambah Umat Baru</h2>
      <p class="page-subtitle">Isi formulir berikut untuk menambahkan data umat baru ke dalam sistem</p>
    </div>

    <!-- Form -->
    <UmatForm :initial-data="formInitialData" :submitting="submitting || ocrLoading" :field-errors="fieldErrors"
      submit-text="Tambah Umat" @submit="handleCreate" @cancel="handleBack" />

    <!-- Dialog Hasil OCR -->
    <el-dialog v-model="ocrDialogVisible" title="Hasil Recognition & Parsing OCR (OCR.space Engine 3)"
      :width="isMobile ? '90%' : '600px'" destroy-on-close>
      <div class="ocr-result-container">
        <div v-if="ocrPreviewUrl" class="ocr-image-preview">
          <span class="preview-label">Gambar:</span>
          <img :src="ocrPreviewUrl" alt="OCR Preview" class="preview-img" />
        </div>

        <!-- Section Hasil Parsing Struktur Umat -->
        <div v-if="ocrParsedUmat && Object.keys(ocrParsedUmat).length > 0" class="ocr-parsed-card">
          <div class="parsed-header">
            <span class="parsed-title">Hasil:</span>
            <el-tag v-if="ocrProcessingTime" type="info" size="small">
              Waktu: {{ ocrProcessingTime }} ms
            </el-tag>
          </div>
          <el-descriptions border size="small" :column="isMobile ? 1 : 2">
            <el-descriptions-item label="Nama Indonesia / Chiu Tao">
              <strong>{{ ocrParsedUmat.nama_indonesia || '-' }}</strong>
            </el-descriptions-item>
            <el-descriptions-item label="Nama Mandarin">
              <strong>{{ ocrParsedUmat.nama_mandarin || '-' }}</strong>
            </el-descriptions-item>
            <el-descriptions-item label="Alias (Pin Yin)">
              {{ ocrParsedUmat.alias || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="Usia">
              {{ ocrParsedUmat.usia ? `${ocrParsedUmat.usia}` : '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="Jenis Kelamin">
              <el-tag v-if="ocrParsedUmat.jenis_kelamin === 'WANITA'" type="danger" size="small">WANITA</el-tag>
              <el-tag v-else-if="ocrParsedUmat.jenis_kelamin === 'PRIA'" type="primary" size="small">PRIA
                (L)</el-tag>
              <el-tag v-else-if="ocrParsedUmat.jenis_kelamin === 'ANAK PRIA'" type="primary" size="small">Anak
                Laki-laki</el-tag>
              <el-tag v-else-if="ocrParsedUmat.jenis_kelamin === 'ANAK WANITA'" type="primary" size="small">Anak
                Wanita</el-tag>
              <span v-else>-</span>
            </el-descriptions-item>
            <el-descriptions-item label="Pendidikan">
              {{ ocrParsedUmat.pendidikan || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="Mobile / HP">
              {{ ocrParsedUmat.mobile || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="Uang Pahala">
              {{ ocrParsedUmat.uang_pahala ? `Rp ${ocrParsedUmat.uang_pahala.toLocaleString('id-ID')}` : '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="Pengajak (Perantara)">
              {{ ocrParsedUmat.pengajak_manual || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="Penanggung">
              {{ ocrParsedUmat.penanggung_manual || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="TCS">
              {{ ocrParsedUmat.tcs || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="Waktu Chiu Tao (Mandarin)">
              {{ ocrParsedUmat.waktu_chiutao_mandarin || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="Alamat" :span="2">
              {{ ocrParsedUmat.alamat || '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <div class="ocr-text-result">
          <span class="result-label">Teks Raw Hasil OCR:</span>
          <el-input v-model="ocrParsedText" type="textarea" :rows="6" readonly
            placeholder="Tidak ada teks yang dapat dikenali" />
        </div>

        <div v-if="ocrErrorText" class="ocr-error-alert">
          <el-alert title="Peringatan/Error OCR" type="warning" :description="ocrErrorText" show-icon
            :closable="false" />
        </div>

        <!--  <el-collapse class="raw-json-collapse">
          <el-collapse-item title="Lihat Raw JSON Response" name="1">
            <pre class="json-preview">{{ JSON.stringify(ocrResultRaw, null, 2) }}</pre>
          </el-collapse-item>
        </el-collapse> -->
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button v-if="ocrParsedUmat && Object.keys(ocrParsedUmat).length > 0" type="primary" :icon="Check"
            @click="applyOcrToForm">
            Isi ke Form Umat
          </el-button>
          <!-- <el-button :icon="CopyDocument" type="success" plain @click="copyOcrText">
            Salin Teks Raw
          </el-button> -->
          <el-button @click="ocrDialogVisible = false">Tutup</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core'
import { useRouter } from 'vue-router'
import { ElMessage, ElNotification, ElLoading } from 'element-plus'
import { Back, Picture, CopyDocument, Check } from '@element-plus/icons-vue'
import UmatForm from '../../components/umat/UmatForm.vue'
import { umatApi } from '../../api/umat'
import type { Umat } from '../../types/umat'
import { scrollToFormError } from '../../utils/scroll'

interface OcrWord {
  WordText: string
  Left: number
  Top: number
  Width: number
  Height: number
}

interface OcrLine {
  Words: OcrWord[]
  MaxHeight: number
  MinTop: number
}

interface OcrResultData {
  ParsedResults?: Array<{
    ParsedText?: string
    ErrorMessage?: string
    ErrorDetails?: string
    FileParseExitCode?: number
    TextOverlay?: {
      Lines?: OcrLine[]
    }
  }>
  OCRExitCode?: number
  IsErroredOnProcessing?: boolean
  ProcessingTimeInMilliseconds?: string
  ErrorMessage?: string[]
  ErrorDetails?: string[]
}

interface ParsedUmatData {
  nama_indonesia?: string
  nama_mandarin?: string
  alias?: string
  usia?: number
  tanggal_lahir?: string
  tanggal_chiutao_int?: string
  pendidikan?: string
  jenis_kelamin?: string
  alamat?: string
  mobile?: string
  pengajak_manual?: string
  penanggung_manual?: string
  tcs?: string
  uang_pahala?: number
  waktu_chiutao_mandarin?: string
  fotang_aktif?: string
  fotang_chiutao?: string
}

const router = useRouter()
// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px

const submitting = ref(false)
const fieldErrors = ref<Record<string, string[]>>({})
const formInitialData = ref<Partial<Umat>>({})

// OCR state variables
const ocrLoading = ref(false)
const ocrDialogVisible = ref(false)
const ocrFileInput = ref<HTMLInputElement | null>(null)
const ocrParsedText = ref('')
const ocrProcessingTime = ref('')
const ocrErrorText = ref('')
const ocrResultRaw = ref<unknown>(null)
const ocrParsedUmat = ref<ParsedUmatData | null>(null)
const ocrPreviewUrl = ref('')
const ocrRateLimitCountdown = ref(0)
let ocrCountdownTimer: ReturnType<typeof setInterval> | null = null

function startOcrCountdown(seconds: number) {
  ocrRateLimitCountdown.value = seconds
  if (ocrCountdownTimer) clearInterval(ocrCountdownTimer)
  ocrCountdownTimer = setInterval(() => {
    ocrRateLimitCountdown.value -= 1
    if (ocrRateLimitCountdown.value <= 0) {
      if (ocrCountdownTimer) clearInterval(ocrCountdownTimer)
      ocrCountdownTimer = null
      ocrRateLimitCountdown.value = 0
    }
  }, 1000)
}

onUnmounted(() => {
  if (ocrCountdownTimer) clearInterval(ocrCountdownTimer)
})

function handleBack() {
  router.push('/master-data/umat')
}

function triggerOcrInput() {
  ocrFileInput.value?.click()
}

async function handleOcrFileSelected(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  ocrLoading.value = true
  submitting.value = true
  ocrParsedText.value = ''
  ocrProcessingTime.value = ''
  ocrErrorText.value = ''
  ocrResultRaw.value = null
  ocrParsedUmat.value = null

  // Generate temporary preview URL
  ocrPreviewUrl.value = URL.createObjectURL(file)

  const loadingInstance = ElLoading.service({
    lock: true,
    text: 'Mengenali teks gambar (OCR)...',
    background: 'rgba(0, 0, 0, 0.7)'
  })

  try {
    const res = await umatApi.ocrUmat(file)
    const payload = res.data as { parsed_umat?: ParsedUmatData; ocr_raw?: OcrResultData }
    const rawData: OcrResultData = payload.ocr_raw || (payload as OcrResultData)
    ocrResultRaw.value = payload
    ocrParsedUmat.value = payload.parsed_umat || null

    ocrProcessingTime.value = rawData.ProcessingTimeInMilliseconds || ''

    if (rawData.ParsedResults && rawData.ParsedResults.length > 0) {
      const parsed = rawData.ParsedResults[0]
      ocrParsedText.value = parsed.ParsedText || ''
      if (parsed.ErrorMessage) {
        ocrErrorText.value = `${parsed.ErrorMessage} ${parsed.ErrorDetails || ''}`.trim()
      }
    } else if (rawData.ErrorMessage && rawData.ErrorMessage.length > 0) {
      ocrErrorText.value = rawData.ErrorMessage.join(', ')
    }

    ocrDialogVisible.value = true
    ElMessage.success('Proses OCR & Parsing selesai')
  } catch (err: unknown) {
    console.error('Failed to perform OCR:', err)
    const errorObj = err as {
      response?: {
        status?: number
        data?: { error?: string; retry_after?: number }
        headers?: Record<string, string>
      }
      message?: string
    }

    if (errorObj.response?.status === 429) {
      const retryAfter =
        errorObj.response.data?.retry_after ||
        parseInt(errorObj.response.headers?.['retry-after'] || errorObj.response.headers?.['x-retry-after'] || '60', 10)
      startOcrCountdown(retryAfter || 60)
      ElNotification({
        title: 'Batas Rate Limit OCR Hit',
        message: errorObj.response.data?.error || `Terlalu banyak permintaan OCR. Harap tunggu ${retryAfter} detik.`,
        type: 'warning',
        duration: 6000
      })
    } else {
      ElMessage.error(errorObj.response?.data?.error || errorObj.message || 'Gagal memproses OCR gambar')
    }
  } finally {
    loadingInstance.close()
    ocrLoading.value = false
    submitting.value = false
    if (target) {
      target.value = ''
    }
  }
}

function applyOcrToForm() {
  if (!ocrParsedUmat.value) return
  const data = { ...ocrParsedUmat.value }

  // Map gender string to option code for el-select
  if (data.jenis_kelamin) {
    const jkUpper = String(data.jenis_kelamin).toUpperCase()
    if (jkUpper === 'PRIA' || jkUpper === 'LAKI-LAKI' || jkUpper === 'L' || jkUpper === '001') {
      data.jenis_kelamin = '001'
    } else if (jkUpper === 'WANITA' || jkUpper === 'PEREMPUAN' || jkUpper === 'P' || jkUpper === 'W' || jkUpper === '002') {
      data.jenis_kelamin = '002'
    } else if (jkUpper === 'ANAK PRIA' || jkUpper === '003') {
      data.jenis_kelamin = '003'
    } else if (jkUpper === 'ANAK WANITA' || jkUpper === '004') {
      data.jenis_kelamin = '004'
    }
  }

  formInitialData.value = {
    ...formInitialData.value,
    ...data
  }
  ocrDialogVisible.value = false
  ElMessage.success('Data hasil OCR berhasil diterapkan ke Form!')
}

async function copyOcrText() {
  if (!ocrParsedText.value) {
    ElMessage.warning('Tidak ada teks untuk disalin')
    return
  }
  try {
    await navigator.clipboard.writeText(ocrParsedText.value)
    ElMessage.success('Teks berhasil disalin ke clipboard')
  } catch {
    ElMessage.error('Gagal menyalin teks')
  }
}

async function handleCreate(payload: Partial<Umat>, photoFile?: File | Blob | null) {
  submitting.value = true
  fieldErrors.value = {}
  try {
    const res = await umatApi.createUmat(payload, photoFile)
    ElNotification({
      title: 'Berhasil',
      message: `Umat ${res.data?.nama_indonesia || ''} berhasil ditambahkan`,
      type: 'success'
    })
    router.push('/master-data/umat')
  } catch (err: unknown) {
    console.error('Failed to create umat:', err)
    const errorObj = err as { response?: { data?: { details?: Record<string, string[]>; error?: string } }; message?: string }
    if (errorObj.response?.data?.details) {
      fieldErrors.value = errorObj.response.data.details
      ElMessage.error(errorObj.response.data.error || 'Invalid Inputs, Silahkan periksa kolom form')
    } else {
      ElMessage.error(errorObj.response?.data?.error || errorObj.message || 'Gagal menambahkan data umat')
    }
    scrollToFormError()
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.create-container {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  max-width: 1100px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 0.5rem;
}

.header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--el-text-color-primary);
  margin: 0.5rem 0 0 0;
}

.page-subtitle {
  font-size: 0.875rem;
  color: var(--el-text-color-secondary);
  margin-top: 0.25rem;
}

.ocr-result-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.ocr-image-preview {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.preview-label,
.result-label,
.parsed-title {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--el-text-color-regular);
}

.parsed-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.ocr-parsed-card {
  background-color: var(--el-fill-color-blank);
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
  padding: 0.75rem;
}

.preview-img {
  max-height: 200px;
  object-fit: contain;
  border-radius: 6px;
  border: 1px solid var(--el-border-color-light);
  background-color: #f8f9fa;
  padding: 4px;
}

.ocr-text-result {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.ocr-error-alert {
  margin-top: 0.25rem;
}

.raw-json-collapse {
  margin-top: 0.5rem;
}

.json-preview {
  background-color: #1e1e1e;
  color: #d4d4d4;
  padding: 0.75rem;
  border-radius: 4px;
  max-height: 200px;
  overflow: auto;
  font-size: 0.8rem;
  font-family: monospace;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
