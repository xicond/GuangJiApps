<template>
  <div class="dashboard-content">
    <el-card class="welcome-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="title">Dashboard</span>
        </div>
      </template>
      <div class="welcome-body">
        <div class="welcome-icon">
          <el-icon :size="48" color="#409eff"><CircleCheckFilled /></el-icon>
        </div>
        <h1>Welcome, {{ authStore.user?.username || 'User' }}!</h1>
        <p class="welcome-subtitle">
          You have successfully logged in to GuangJi Apps Dashboard.
        </p>

        <el-divider />

        <div class="info-section">
          <el-descriptions title="Session Details" :column="1" border>
            <el-descriptions-item label="Username">
              {{ authStore.user?.username || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="Status">
              <el-tag type="success" effect="dark">Active</el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '../stores/auth'
import { CircleCheckFilled } from '@element-plus/icons-vue'

const authStore = useAuthStore()
</script>

<style scoped>
.dashboard-content {
  width: 100%;
}

.welcome-card {
  border-radius: 8px;
}

.card-header .title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.welcome-body {
  text-align: center;
  padding: 1rem 0;
}

.welcome-body h1 {
  margin: 1rem 0 0.5rem 0;
  font-size: 1.75rem;
  color: var(--el-text-color-primary);
}

.welcome-subtitle {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.info-section {
  margin-top: 1.5rem;
  text-align: left;
}
</style>
