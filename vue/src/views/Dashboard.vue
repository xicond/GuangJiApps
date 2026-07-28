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
        <h1>Welcome, {{ usernameDisplay }}!</h1>
        <p class="welcome-subtitle">
          You have successfully logged in to GuangJi Apps Dashboard.
        </p>

        <div class="action-bar" style="margin-top: 1rem;">
          <el-button type="primary" :icon="Key" @click="isChangePasswordVisible = true">
            Change Password
          </el-button>
        </div>

        <el-divider />

        <div class="info-section">
          <el-descriptions title="Session Details" :column="1" border>
            <el-descriptions-item label="Username">
              {{ usernameDisplay }}
            </el-descriptions-item>
            <el-descriptions-item label="Status">
              <el-tag type="success" effect="dark">Active</el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </el-card>

    <ChangePasswordModal v-model="isChangePasswordVisible" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { CircleCheckFilled, Key } from '@element-plus/icons-vue'
import ChangePasswordModal from '../components/ChangePasswordModal.vue'

const authStore = useAuthStore()
const isChangePasswordVisible = ref(false)

const usernameDisplay = computed(() => {
  return authStore.user?.username || authStore.user?.Username || 'User'
})
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
