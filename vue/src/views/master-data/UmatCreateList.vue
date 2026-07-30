<template>
  <div class="create-container">
    <!-- Header Section -->
    <div class="page-header">
      <el-button :icon="Back" text @click="handleBack">Kembali ke Daftar</el-button>
      <h2 class="page-title">Tambah Umat Baru</h2>
      <p class="page-subtitle">Isi formulir berikut untuk menambahkan data umat baru ke dalam sistem</p>
    </div>

    <!-- Form -->
    <UmatForm
      :submitting="submitting"
      :field-errors="fieldErrors"
      submit-text="Tambah Umat"
      @submit="handleCreate"
      @cancel="handleBack"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { Back } from '@element-plus/icons-vue'
import UmatForm from '../../components/umat/UmatForm.vue'
import { umatApi } from '../../api/umat'
import type { Umat } from '../../types/umat'

const router = useRouter()
const submitting = ref(false)
const fieldErrors = ref<Record<string, string[]>>({})

function handleBack() {
  router.push('/master-data/umat')
}

async function handleCreate(payload: Partial<Umat>) {
  submitting.value = true
  fieldErrors.value = {}
  try {
    const res = await umatApi.createUmat(payload)
    ElNotification({
      title: 'Berhasil',
      message: `Umat ${res.data?.nama_indonesia || ''} berhasil ditambahkan`,
      type: 'success'
    })
    router.push('/master-data/umat')
  } catch (err: any) {
    console.error('Failed to create umat:', err)
    if (err.response?.data?.details) {
      fieldErrors.value = err.response.data.details
      ElMessage.error(err.response.data.error || 'Validasi gagal, Silahkan periksa kolom form')
    } else {
      ElMessage.error(err.response?.data?.error || err.message || 'Gagal menambahkan data umat')
    }
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
