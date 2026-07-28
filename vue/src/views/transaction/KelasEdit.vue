<template>
  <div class="edit-container">
    <!-- Header Section -->
    <div class="page-header">
      <el-button :icon="Back" text @click="handleBack">Kembali ke Daftar</el-button>
      <h2 class="page-title">Edit Data Kelas #{{ kelasId }}</h2>
      <p class="page-subtitle">Perbarui jadwal dan informasi kegiatan kelas di bawah ini</p>
    </div>

    <!-- Loading Skeleton -->
    <el-card v-if="fetching" shadow="never">
      <el-skeleton :rows="8" animated />
    </el-card>

    <!-- Form Component -->
    <KelasForm
      v-else
      :initial-data="kelasData"
      :submitting="submitting"
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
import KelasForm from '../../components/kelas/KelasForm.vue'
import { kelasApi } from '../../api/kelas'
import type { Kelas } from '../../types/kelas'

const route = useRoute()
const router = useRouter()

const kelasId = route.params.id as string
const kelasData = ref<Partial<Kelas>>({})
const fetching = ref(true)
const submitting = ref(false)

let abortController: AbortController | null = null

function handleBack() {
  router.push('/transaction/kelas')
}

async function loadKelas() {
  if (abortController) abortController.abort()
  abortController = new AbortController()

  fetching.value = true
  try {
    const res = await kelasApi.getKelasById(kelasId, abortController.signal)
    kelasData.value = res.data || {}
  } catch (err: any) {
    if (err.name === 'CanceledError' || err.name === 'AbortError') return
    console.error('Failed to fetch kelas detail:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal mengambil detail data kelas')
    router.push('/transaction/kelas')
  } finally {
    fetching.value = false
  }
}

async function handleUpdate(payload: Partial<Kelas>) {
  submitting.value = true
  try {
    await kelasApi.updateKelas(kelasId, payload)
    ElNotification({
      title: 'Berhasil',
      message: 'Data kelas berhasil diperbarui',
      type: 'success'
    })
    router.push('/transaction/kelas')
  } catch (err: any) {
    console.error('Failed to update kelas:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal memperbarui data kelas')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadKelas()
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
