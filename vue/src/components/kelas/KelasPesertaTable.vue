<template>
  <el-card shadow="never" class="peserta-card">
    <template #header>
      <div class="card-header">
        <div class="header-title">
          <el-icon class="header-icon"><UserFilled /></el-icon>
          <span>Daftar Peserta Kelas</span>
          <el-tag size="small" type="info" class="ml-2">{{ total }} Peserta</el-tag>
        </div>
        <el-button :icon="Refresh" circle size="small" title="Refresh Peserta" @click="fetchPeserta" />
      </div>
    </template>

    <el-table
      v-loading="loading"
      :data="pesertaList"
      stripe
      border
      max-height="450"
      style="width: 100%"
      empty-text="Belum ada peserta yang terdaftar pada kelas ini"
    >
      <el-table-column type="index" label="No." width="60" align="center" fixed="left" />

      <!-- <el-table-column prop="id_peserta" label="ID Peserta" width="120" align="center" fixed="left">
        <template #default="{ row }">
          <el-tag size="small" type="primary" class="font-mono">{{ row.id_peserta || '-' }}</el-tag>
        </template>
      </el-table-column> -->

      <el-table-column prop="nama_indonesia" label="Nama Ciu Tao" min-width="160">
        <template #default="{ row }">
          <span class="font-semibold">{{ row.nama_indonesia || '-' }}</span>
        </template>
      </el-table-column>

      <el-table-column prop="nama_mandarin" label="Nama Lain" min-width="140">
        <template #default="{ row }">
          <span>{{ row.nama_mandarin || '-' }}</span>
        </template>
      </el-table-column>

      <el-table-column prop="fotang_ciutao_desc" label="Fotang Ciu Tao" min-width="150" />

      <el-table-column prop="fotang_aktif_desc" label="Fotang Aktif" min-width="150" />

      <el-table-column prop="tanggal_ciu_tao_int" label="Tgl Ciu Tao" width="120" align="center" />

      <el-table-column prop="pengajak" label="Pengajak" min-width="140" />

      <el-table-column prop="penanggung" label="Penanggung" min-width="140" />

      <el-table-column label="Status Lulus" width="140" align="center">
        <template #default="{ row }">
          <el-tag :type="row.lulus ? 'success' : 'info'" size="small">
            {{ row.lulus ? 'Lulus' : 'Belum' }}
          </el-tag>
          <div v-if="row.keterangan_lulus" class="lulus-note">
            {{ row.keterangan_lulus }}
          </div>
        </template>
      </el-table-column>

      <el-table-column label="Status Ikrar" min-width="180" align="center">
        <template #default="{ row }">
          <div class="ikrar-tags">
            <el-tag v-if="row.ikrar1" size="small" type="success">I-1</el-tag>
            <el-tag v-if="row.ikrar2" size="small" type="success">I-2</el-tag>
            <el-tag v-if="row.ikrar3" size="small" type="success">I-3</el-tag>
            <el-tag v-if="row.ikrar4" size="small" type="success">I-4</el-tag>
            <el-tag v-if="row.ikrar5" size="small" type="success">I-5</el-tag>
            <el-tag v-if="row.ikrar6" size="small" type="success">I-6</el-tag>
            <span v-if="!row.ikrar1 && !row.ikrar2 && !row.ikrar3 && !row.ikrar4 && !row.ikrar5 && !row.ikrar6" class="no-ikrar">-</span>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { UserFilled, Refresh } from '@element-plus/icons-vue'
import { kelasApi } from '../../api/kelas'
import type { KelasPeserta } from '../../types/kelas'

const props = defineProps<{
  kelasId: string | number
}>()

const pesertaList = ref<KelasPeserta[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
let abortController: AbortController | null = null

async function fetchPeserta() {
  if (!props.kelasId) return
  if (abortController) abortController.abort()
  abortController = new AbortController()

  loading.value = true
  try {
    const res = await kelasApi.getKelasPeserta(
      props.kelasId,
      { page: currentPage.value, limit: pageSize.value },
      abortController.signal
    )
    pesertaList.value = res.data || []
    total.value = res.meta?.total ?? pesertaList.value.length
  } catch (err: any) {
    if (err.name === 'AbortError' || err.name === 'CanceledError') return
    console.error('Error fetching kelas peserta:', err)
  } finally {
    loading.value = false
  }
}

function handleSizeChange(newSize: number) {
  pageSize.value = newSize
  currentPage.value = 1
  fetchPeserta()
}

function handleCurrentChange(newPage: number) {
  currentPage.value = newPage
  fetchPeserta()
}

watch(
  () => props.kelasId,
  (newId) => {
    if (newId) {
      currentPage.value = 1
      fetchPeserta()
    }
  },
  { immediate: true }
)

onMounted(() => {
  fetchPeserta()
})

onUnmounted(() => {
  if (abortController) abortController.abort()
})
</script>

<style scoped>
.peserta-card {
  border-radius: 10px;
  background-color: var(--el-bg-color-overlay);
  margin-top: 1.25rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  font-size: 1rem;
  color: var(--el-text-color-primary);
}

.header-icon {
  color: var(--el-color-primary);
  font-size: 1.2rem;
}

.font-mono {
  font-family: monospace;
}

.font-semibold {
  font-weight: 600;
}

.lulus-note {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.ikrar-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  justify-content: center;
}

.no-ikrar {
  color: var(--el-text-color-placeholder);
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}
</style>
