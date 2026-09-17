<template>
  <div class="view-detail-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <el-button :icon="Back" text @click="handleBack">Kembali ke Daftar</el-button>
        <h2 class="page-title">Detail Kelas #{{ kelasId }}</h2>
        <p class="page-subtitle">Informasi detail kegiatan kelas dan daftar peserta terdaftar</p>
      </div>

      <div class="header-actions">
        <el-button type="primary" :icon="Edit" @click="handleEdit">
          Edit Kelas
        </el-button>
      </div>
    </div>

    <!-- Loading Skeleton -->
    <el-card v-if="fetching" shadow="never">
      <el-skeleton :rows="6" animated />
    </el-card>

    <template v-else>
      <!-- Detail Card -->
      <el-card shadow="never" class="detail-card">
        <el-descriptions title="Informasi Kelas" :column="isMobile ? 1 : 2" border>
          <el-descriptions-item label="Nama Kelas">
            <span class="font-semibold">{{ kelasData.kelas_name?.lookup_description || kelasData.kode_kelas || '-'
            }}</span>
          </el-descriptions-item>

          <el-descriptions-item label="Fotang">
            <span>{{ kelasData.fotang_name?.lookup_description || kelasData.kode_fotang || '-' }}</span>
          </el-descriptions-item>

          <el-descriptions-item label="Tanggal Mulai">
            <span>{{ kelasData.start_date || '-' }}</span>
          </el-descriptions-item>

          <el-descriptions-item label="Tanggal Selesai">
            <span>{{ kelasData.end_date || '-' }}</span>
          </el-descriptions-item>

          <el-descriptions-item label="Lokasi Pelaksanaan">
            <span>{{ kelasData.lokasi || '-' }}</span>
          </el-descriptions-item>

          <el-descriptions-item label="PIC / Penanggung Jawab">
            <span>{{ kelasData.pic || '-' }}</span>
          </el-descriptions-item>

          <el-descriptions-item label="Tingkat / Level">
            <span>{{ kelasData.level || '-' }}</span>
          </el-descriptions-item>

          <el-descriptions-item label="Batas Pendaftaran (Deadline)">
            <span>{{ kelasData.deadline || '-' }}</span>
          </el-descriptions-item>

          <el-descriptions-item label="Keterangan" :span="isMobile ? 1 : 2">
            <span>{{ kelasData.keterangan || '-' }}</span>
          </el-descriptions-item>

          <!-- MC List if available -->
          <el-descriptions-item v-if="hasMcInfo" label="Daftar MC" :span="isMobile ? 1 : 2">
            <div class="mc-list-tags">
              <el-tag v-if="kelasData.mc1" type="info" size="small">MC 1: {{ kelasData.mc1 }}</el-tag>
              <el-tag v-if="kelasData.mc2" type="info" size="small">MC 2: {{ kelasData.mc2 }}</el-tag>
              <el-tag v-if="kelasData.mc3" type="info" size="small">MC 3: {{ kelasData.mc3 }}</el-tag>
              <el-tag v-if="kelasData.mc4" type="info" size="small">MC 4: {{ kelasData.mc4 }}</el-tag>
              <el-tag v-if="kelasData.mc5" type="info" size="small">MC 5: {{ kelasData.mc5 }}</el-tag>
            </div>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Sub-model Lists -->
      <KelasPesertaTable :kelas-id="kelasId" :kode-kelas="kelasData?.kode_kelas" :start-date="kelasData?.start_date"
        :end-date="kelasData?.end_date" readonly />
      <KelasPengabdiTable :kelas-id="kelasId" :start-date="kelasData?.start_date" :end-date="kelasData?.end_date"
        readonly />
      <KelasTopikTable :kelas-id="kelasId" :start-date="kelasData?.start_date" :end-date="kelasData?.end_date"
        readonly />
      <KelasKendaraanTable :kelas-id="kelasId" :start-date="kelasData?.start_date" :end-date="kelasData?.end_date"
        readonly />
      <KelasDonasiTable :kelas-id="kelasId" readonly />
      <KelasDonasiBarangTable :kelas-id="kelasId" readonly />
      <!-- <KelasPengeluaranTable :kelas-id="kelasId" readonly /> -->
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import { ElMessage } from 'element-plus'
import { Back, Edit } from '@element-plus/icons-vue'
import KelasPesertaTable from '../../components/kelas/KelasPesertaTable.vue'
import KelasPengabdiTable from '../../components/kelas/KelasPengabdiTable.vue'
import KelasTopikTable from '../../components/kelas/KelasTopikTable.vue'
import KelasDonasiTable from '../../components/kelas/KelasDonasiTable.vue'
import KelasDonasiBarangTable from '../../components/kelas/KelasDonasiBarangTable.vue'
import KelasKendaraanTable from '../../components/kelas/KelasKendaraanTable.vue'
import KelasPengeluaranTable from '../../components/kelas/KelasPengeluaranTable.vue'
import { kelasApi } from '../../api/kelas'
import type { Kelas } from '../../types/kelas'

const route = useRoute()
const router = useRouter()
const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')

const kelasId = route.params.id as string
const kelasData = ref<Partial<Kelas>>({})
const fetching = ref(true)

let abortController: AbortController | null = null

const hasMcInfo = computed(() => {
  return !!(kelasData.value.mc1 || kelasData.value.mc2 || kelasData.value.mc3 || kelasData.value.mc4 || kelasData.value.mc5)
})

function handleBack() {
  router.push('/transaction/kelas')
}

function handleEdit() {
  router.push(`/transaction/kelas/edit/${kelasId}`)
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

onMounted(() => {
  loadKelas()
})

onUnmounted(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.view-detail-container {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  max-width: 1000px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1rem;
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

.detail-card {
  border-radius: 10px;
  background-color: var(--el-bg-color-overlay);
}

.font-semibold {
  font-weight: 600;
}

.mc-list-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
</style>
