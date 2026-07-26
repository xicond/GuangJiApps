<template>
  <div class="login-container">
    <div class="theme-switcher">
      <el-dropdown @command="handleThemeChange" trigger="click">
        <el-button circle :icon="currentIcon">
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="system">
              <el-icon><Monitor /></el-icon> System Mode
            </el-dropdown-item>
            <el-dropdown-item command="light">
              <el-icon><Sunny /></el-icon> Light Mode
            </el-dropdown-item>
            <el-dropdown-item command="dark">
              <el-icon><Moon /></el-icon> Dark Mode
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <el-card class="login-card" shadow="always">
      <template #header>
        <div class="card-header">
          <h2>GuangJi Apps</h2>
          <p class="subtitle">Please sign in to continue</p>
        </div>
      </template>

      <el-form
        ref="loginFormRef"
        :model="form"
        :rules="rules"
        label-position="top"
        size="large"
        @keyup.enter="handleLogin"
      >
        <el-form-item label="Username" prop="username">
          <el-input
            v-model="form.username"
            placeholder="Enter username"
            :prefix-icon="User"
            clearable
          />
        </el-form-item>

        <el-form-item label="Password" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="Enter password"
            :prefix-icon="Lock"
            show-password
            clearable
          />
        </el-form-item>

        <el-alert
          v-if="errorMessage"
          :title="errorMessage"
          type="error"
          show-icon
          :closable="false"
          class="error-alert"
        />

        <el-form-item class="submit-item">
          <el-button
            type="primary"
            class="login-button"
            :loading="loading"
            @click="handleLogin"
          >
            Sign In
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { setMode, currentMode, isDark, type ThemeMode } from '../theme'
import { User, Lock, Sunny, Moon, Monitor } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()

const loginFormRef = ref<FormInstance>()
const form = ref({
  username: 'root',
  password: 'password123'
})

const rules: FormRules = {
  username: [{ required: true, message: 'Username is required', trigger: 'blur' }],
  password: [{ required: true, message: 'Password is required', trigger: 'blur' }]
}

const loading = ref(false)
const errorMessage = ref('')

const currentIcon = computed(() => {
  if (currentMode.value === 'system') return Monitor
  return isDark.value ? Moon : Sunny
})

const handleThemeChange = (mode: ThemeMode) => {
  setMode(mode)
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  await loginFormRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    errorMessage.value = ''
    try {
      await authStore.login(form.value.username, form.value.password)
      router.push('/dashboard')
    } catch (err: any) {
      errorMessage.value = err.message || 'Login failed'
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  padding: 1rem;
  background: radial-gradient(circle at center, rgba(64, 158, 255, 0.08) 0%, transparent 70%);
}

.theme-switcher {
  position: absolute;
  top: 1.5rem;
  right: 1.5rem;
}

.login-card {
  width: 100%;
  max-width: 420px;
  border-radius: 12px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
}

.card-header {
  text-align: center;
}

.card-header h2 {
  margin: 0;
  font-size: 1.75rem;
  color: var(--el-text-color-primary);
}

.subtitle {
  margin-top: 0.5rem;
  margin-bottom: 0;
  font-size: 0.9rem;
  color: var(--el-text-color-secondary);
}

.error-alert {
  margin-bottom: 1rem;
}

.submit-item {
  margin-top: 1.5rem;
  margin-bottom: 0;
}

.login-button {
  width: 100%;
}
</style>
