<template>
  <div class="umat-photo-upload">
    <div class="photo-preview-container">
      <div v-if="previewUrl" class="preview-box">
        <img :src="previewUrl" alt="Foto Umat" class="preview-img" />
        <el-button type="danger" circle size="small" class="remove-btn" @click="removePhoto">
          <el-icon>
            <Delete />
          </el-icon>
        </el-button>
      </div>
      <div v-else class="placeholder-box" @click="triggerFileInput">
        <el-icon class="upload-icon">
          <Picture />
        </el-icon>
        <span class="upload-text">{{ deviceType === 'desktop' ? 'Upload Foto (3:4)' : 'Upload / Ambil Foto (3:4)'
        }}</span>
      </div>
    </div>

    <!-- Upload & Camera Controls -->
    <div class="photo-controls">
      <!--  capture="user" -->
      <input ref="fileInputRef" type="file" accept="image/jpeg,image/png" style="display: none"
        @change="onFileSelected" />
      <el-button type="primary" plain size="small" :icon="Upload" @click="triggerFileInput">
        Pilih File Gambar
      </el-button>
      <!-- <el-button type="success" plain size="small" :icon="Camera" @click="openCameraModal">
        Kamera Stream
      </el-button> -->
    </div>

    <!-- Camera Dialog (MediaDevices API) -->
    <el-dialog v-model="showCameraDialog" title="Ambil Foto dari Kamera" width="90%" max-width="500px" destroy-on-close
      @closed="stopCameraStream">
      <div class="camera-container">
        <video ref="videoRef" autoplay playsinline class="camera-video"></video>
        <canvas ref="canvasRef" style="display: none"></canvas>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showCameraDialog = false">Batal</el-button>
          <el-button type="primary" :icon="Camera" @click="capturePhoto">Tangkap Foto</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Cropper Dialog (vue-advanced-cropper) -->
    <el-dialog v-model="showCropperDialog" title="Potong Foto (Rasio 3:4)" width="90%" max-width="600px"
      destroy-on-close>
      <div class="cropper-container">
        <Cropper ref="cropperRef" :src="rawImageSrc" :stencil-props="{ aspectRatio: 3 / 4 }" class="cropper-box" />
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showCropperDialog = false">Batal</el-button>
          <el-button type="primary" :icon="Check" @click="applyCrop">Gunakan Foto Ini</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Cropper } from 'vue-advanced-cropper'
import 'vue-advanced-cropper/dist/style.css'
import { Picture, Upload, Camera, Delete, Check } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  initialUrl?: string
}>()

const emit = defineEmits<{
  (e: 'change', fileBlob: Blob | null): void
}>()


let deviceType: string = "";
const getDeviceType = (): string => {
  const ua = navigator.userAgent
  const maxTouchPoints = navigator.maxTouchPoints || 0

  // 1. Deteksi iPadOS modern (UA mengandung Macintosh/MacIntel, tapi mendukung multi-touch)
  const isIPadOS = /Macintosh/i.test(ua) && maxTouchPoints > 1

  // 2. Deteksi Tablet umum (Android tablet, Playbook, Silk, atau iPadOS)
  if (/tablet|playbook|silk/i.test(ua) || isIPadOS) {
    return 'tablet'
  }

  // 3. Deteksi Mobile (Android Phone, iPhone, iPod, dll)
  if (/mobi|android|iphone|ipod/i.test(ua)) {
    return 'mobile'
  }

  // 4. Selebihnya dianggap Desktop (termasuk Mac asli yang maxTouchPoints-nya 0)
  return 'desktop'
}

onMounted(() => {
  getDeviceType()
})

const fileInputRef = ref<HTMLInputElement | null>(null)
const previewUrl = ref<string>(props.initialUrl || '')
const rawImageSrc = ref<string>('')
const showCropperDialog = ref<boolean>(false)
const showCameraDialog = ref<boolean>(false)

const cropperRef = ref<any>(null)
const videoRef = ref<HTMLVideoElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
let mediaStream: MediaStream | null = null

function triggerFileInput() {
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
    fileInputRef.value.click()
  }
}

function onFileSelected(event: Event) {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    const file = target.files[0]
    if (!file.type.startsWith('image/')) {
      ElMessage.error('File harus berupa gambar (JPEG, PNG)')
      return
    }
    const reader = new FileReader()
    reader.onload = (e) => {
      if (e.target?.result) {
        rawImageSrc.value = e.target.result as string
        showCropperDialog.value = true
      }
    }
    reader.readAsDataURL(file)
  }
}

async function openCameraModal() {
  showCameraDialog.value = true
  try {
    mediaStream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: 'environment', width: { ideal: 1280 }, height: { ideal: 720 } }
    })
    setTimeout(() => {
      if (videoRef.value && mediaStream) {
        videoRef.value.srcObject = mediaStream
      }
    }, 200)
  } catch (err) {
    console.error('Camera access error:', err)
    ElMessage.error('Tidak dapat mengakses kamera perangkat')
    showCameraDialog.value = false
  }
}

function stopCameraStream() {
  if (mediaStream) {
    mediaStream.getTracks().forEach((track) => track.stop())
    mediaStream = null
  }
}

function capturePhoto() {
  if (!videoRef.value || !canvasRef.value) return
  const video = videoRef.value
  const canvas = canvasRef.value
  canvas.width = video.videoWidth || 640
  canvas.height = video.videoHeight || 480
  const ctx = canvas.getContext('2d')
  if (ctx) {
    ctx.drawImage(video, 0, 0, canvas.width, canvas.height)
    rawImageSrc.value = canvas.toDataURL('image/jpeg', 0.9)
    stopCameraStream()
    showCameraDialog.value = false
    showCropperDialog.value = true
  }
}

function applyCrop() {
  if (!cropperRef.value) return
  const { canvas } = cropperRef.value.getResult()
  if (!canvas) {
    ElMessage.error('Gagal memotong foto')
    return
  }

  // Enforce 3:4 aspect ratio with standard resolution (e.g. 750x1000 px)
  const targetWidth = 750
  const targetHeight = 1000
  const outputCanvas = document.createElement('canvas')
  outputCanvas.width = targetWidth
  outputCanvas.height = targetHeight
  const ctx = outputCanvas.getContext('2d')
  if (ctx) {
    ctx.drawImage(canvas, 0, 0, targetWidth, targetHeight)
    outputCanvas.toBlob(
      (blob) => {
        if (blob) {
          if (previewUrl.value && previewUrl.value.startsWith('blob:')) {
            URL.revokeObjectURL(previewUrl.value)
          }
          previewUrl.value = URL.createObjectURL(blob)
          emit('change', blob)
          showCropperDialog.value = false
          ElMessage.success('Foto berhasil dipotong dan diterapkan')
        }
      },
      'image/jpeg',
      0.9
    )
  }
}

function removePhoto() {
  if (previewUrl.value && previewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(previewUrl.value)
  }
  previewUrl.value = ''
  emit('change', null)
}

onBeforeUnmount(() => {
  stopCameraStream()
  if (previewUrl.value && previewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(previewUrl.value)
  }
})
</script>

<style scoped>
.umat-photo-upload {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
}

.photo-preview-container {
  width: 100%;
  max-width: 210px;
  aspect-ratio: 3 / 4;
  /* 3:4 Ratio preview box */
  border: 2px dashed var(--el-border-color);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  background-color: var(--el-bg-color-page);
  cursor: pointer;
  transition: border-color 0.2s;
}

.photo-preview-container:hover {
  border-color: var(--el-color-primary);
}

.preview-box {
  width: 100%;
  height: 100%;
  position: relative;
}

.preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.remove-btn {
  position: absolute;
  top: 6px;
  right: 6px;
}

.placeholder-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  /* width: 100%; */
  height: 100%;
  padding: 0.5rem;
  text-align: center;
  color: var(--el-text-color-secondary);
}

.upload-icon {
  font-size: 2.2rem;
  margin-bottom: 0.5rem;
  color: var(--el-color-primary);
}

.upload-text {
  font-size: 0.75rem;
  line-height: 1.2;
}

.photo-controls {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  justify-content: center;
}

.camera-container {
  width: 100%;
  max-height: 350px;
  background: #000;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 6px;
  overflow: hidden;
}

.camera-video {
  width: 100%;
  max-height: 350px;
  object-fit: contain;
}

.cropper-container {
  width: 100%;
  height: 400px;
  background: #222;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  overflow: hidden;
}

.cropper-box {
  width: 100%;
  height: 100%;
}
</style>
