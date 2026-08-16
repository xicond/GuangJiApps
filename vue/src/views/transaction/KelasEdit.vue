<template>
  <div class="edit-container">
    <!-- Header Section -->
    <div class="page-header">
      <el-button :icon="Back" text @click="handleBack">Kembali ke Daftar</el-button>
      <h2 class="page-title">Edit Data Kelas #{{ kelasId }}</h2>
      <p class="page-subtitle">Perbarui jadwal, informasi kegiatan kelas, serta daftar peserta di bawah ini</p>
    </div>

    <!-- Loading Skeleton -->
    <el-card v-if="fetching" shadow="never">
      <el-skeleton :rows="8" animated />
    </el-card>

    <template v-else>
      <!-- Form Component -->
      <KelasForm
        :initial-data="kelasData"
        :submitting="submitting"
        :field-errors="fieldErrors"
        :submit-text="isMobile ? 'Simpan' : 'Simpan Perubahan'"
        @submit="handleUpdate"
        @cancel="handleBack"
      />

      <!-- Sub-model Lists -->
      <KelasPesertaTable :kelas-id="kelasId" :start-date="kelasData?.start_date" :end-date="kelasData?.end_date" />
      <KelasPengabdiTable :kelas-id="kelasId" :start-date="kelasData?.start_date" :end-date="kelasData?.end_date" />
      <KelasTopikTable :kelas-id="kelasId" :start-date="kelasData?.start_date" :end-date="kelasData?.end_date" />
      <KelasDonasiTable :kelas-id="kelasId" />
      <KelasDonasiBarangTable :kelas-id="kelasId" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { Back } from '@element-plus/icons-vue'
import KelasForm from '../../components/kelas/KelasForm.vue'
import KelasPesertaTable from '../../components/kelas/KelasPesertaTable.vue'
import KelasPengabdiTable from '../../components/kelas/KelasPengabdiTable.vue'
import KelasTopikTable from '../../components/kelas/KelasTopikTable.vue'
import KelasDonasiTable from '../../components/kelas/KelasDonasiTable.vue'
import KelasDonasiBarangTable from '../../components/kelas/KelasDonasiBarangTable.vue'
import { kelasApi } from '../../api/kelas'
import type { Kelas } from '../../types/kelas'
import { scrollToFormError } from '../../utils/scroll'

const route = useRoute()
const router = useRouter()

// Initialize breakpoints (Tailwind or custom layout mapping)
const breakpoints = useBreakpoints(breakpointsTailwind)

// Subscribe to reactive states
const isMobile = breakpoints.smaller('md')   // True if width < 768px
const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px

const kelasId = route.params.id as string
const kelasData = ref<Partial<Kelas>>({})
const fetching = ref(true)
const submitting = ref(false)
const fieldErrors = ref<Record<string, string[]>>({})

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
  fieldErrors.value = {}
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
    if (err.response?.data?.details) {
      fieldErrors.value = err.response.data.details
      ElMessage.error(err.response.data.error || 'Invalid Inputs, Silahkan periksa kolom form')
    } else {
      ElMessage.error(err.response?.data?.error || err.message || 'Gagal memperbarui data kelas')
    }
    scrollToFormError()
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
