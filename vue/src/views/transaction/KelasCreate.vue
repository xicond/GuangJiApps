<template>
  <div class="create-container">
    <!-- Header Section -->
    <div class="page-header">
      <el-button :icon="Back" text @click="handleBack">Kembali ke Daftar</el-button>
      <h2 class="page-title">Tambah Kelas Baru</h2>
      <p class="page-subtitle">Isi formulir di bawah ini untuk menjadwalkan kegiatan kelas baru</p>
    </div>

    <!-- Form Component -->
    <KelasForm
      :submitting="submitting"
      :field-errors="fieldErrors"
      submit-text="Tambah Kelas"
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
import KelasForm from '../../components/kelas/KelasForm.vue'
import { kelasApi } from '../../api/kelas'
import type { Kelas } from '../../types/kelas'

const router = useRouter()
const submitting = ref(false)
const fieldErrors = ref<Record<string, string[]>>({})

function handleBack() {
  router.push('/transaction/kelas')
}

async function handleCreate(payload: Partial<Kelas>) {
  submitting.value = true
  fieldErrors.value = {}
  try {
    await kelasApi.createKelas(payload)
    ElNotification({
      title: 'Berhasil',
      message: 'Data kelas baru berhasil ditambahkan',
      type: 'success'
    })
    router.push('/transaction/kelas')
  } catch (err: any) {
    console.error('Failed to create kelas:', err)
    if (err.response?.data?.details) {
      fieldErrors.value = err.response.data.details
      ElMessage.error(err.response.data.error || 'Invalid Inputs, Silahkan periksa kolom form')
    } else {
      ElMessage.error(err.response?.data?.error || err.message || 'Gagal menambahkan data kelas')
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
  max-width: 1000px;
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
