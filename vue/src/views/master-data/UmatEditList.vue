<template>
  <div class="edit-container">
    <!-- Header Section -->
    <div class="page-header">
      <el-button :icon="Back" text @click="handleBack">Kembali ke Daftar</el-button>
      <h2 class="page-title">Edit Data Umat #{{ umatId }}</h2>
      <p class="page-subtitle">Perbarui informasi profil dan data umat di bawah ini</p>
    </div>

    <!-- Loading Skeleton -->
    <el-card v-if="fetching" shadow="never">
      <el-skeleton :rows="8" animated />
    </el-card>

    <!-- Edit Form -->
    <UmatForm
      v-else
      :initial-data="umatData"
      :submitting="submitting"
      :field-errors="fieldErrors"
      submit-text="Simpan Perubahan"
      @submit="handleUpdate"
      @cancel="handleBack"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { Back } from '@element-plus/icons-vue'
import UmatForm from '../../components/umat/UmatForm.vue'
import { umatApi } from '../../api/umat'
import type { Umat } from '../../types/umat'

const route = useRoute()
const router = useRouter()

const umatId = route.params.id as string
const umatData = ref<Partial<Umat>>({})
const fetching = ref(true)
const submitting = ref(false)
const fieldErrors = ref<Record<string, string[]>>({})

let abortController: AbortController | null = null

function handleBack() {
  router.push('/master-data/umat')
}

async function loadUmat() {
  if (abortController) abortController.abort()
  abortController = new AbortController()

  fetching.value = true
  try {
    const res = await umatApi.getUmatById(umatId, abortController.signal)
    umatData.value = res.data || {}
  } catch (err: any) {
    if (err.name === 'CanceledError' || err.name === 'AbortError') return
    console.error('Failed to fetch umat detail:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal mengambil data umat')
    router.push('/master-data/umat')
  } finally {
    fetching.value = false
  }
}

async function handleUpdate(payload: Partial<Umat>) {
  submitting.value = true
  fieldErrors.value = {}
  try {
    await umatApi.updateUmat(umatId, payload)
    ElNotification({
      title: 'Berhasil',
      message: 'Data umat berhasil diperbarui',
      type: 'success'
    })
    router.push('/master-data/umat')
  } catch (err: any) {
    console.error('Failed to update umat:', err)
    if (err.response?.data?.details) {
      fieldErrors.value = err.response.data.details
      ElMessage.error(err.response.data.error || 'Validasi gagal, Silahkan periksa kolom form')
    } else {
      ElMessage.error(err.response?.data?.error || err.message || 'Gagal memperbarui data umat')
    }
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadUmat()
})

onUnmounted(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.edit-container {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  max-width: 1100px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 0.5rem;
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
</style>
