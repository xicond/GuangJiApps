<template>
  <div class="placeholder-container">
    <el-card shadow="never" class="page-card">
      <template #header>
        <div class="card-header">
          <span class="title">{{ pageTitle }}</span>
          <el-tag type="info" size="small" effect="plain">Blank Page</el-tag>
        </div>
      </template>
      <div class="card-body">
        <el-empty :description="`Content for ${pageTitle} module`" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const pageTitle = computed(() => (route.meta.title as string) || 'Menu Header')
</script>

<style scoped>
.placeholder-container {
  width: 100%;
}

.page-card {
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.card-body {
  padding: 2rem 0;
}
</style>
