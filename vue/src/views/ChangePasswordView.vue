<template>
  <div class="change-password-page">
    <el-card class="box-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="title">Change Password</span>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        style="max-width: 480px; margin: 0 auto;"
        @keyup.enter="handleSubmit"
      >
        <el-form-item label="Old Password" prop="oldPassword">
          <el-input
            v-model="form.oldPassword"
            type="password"
            show-password
            placeholder="Enter current password"
            :prefix-icon="Lock"
          />
        </el-form-item>

        <el-form-item label="New Password" prop="newPassword">
          <el-input
            v-model="form.newPassword"
            type="password"
            show-password
            placeholder="Enter new password"
            :prefix-icon="Key"
          />
        </el-form-item>

        <el-form-item label="Confirm New Password" prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            show-password
            placeholder="Confirm new password"
            :prefix-icon="Key"
          />
        </el-form-item>

        <el-form-item style="margin-top: 2rem;">
          <el-button type="primary" :loading="loading" style="width: 100%;" @click="handleSubmit">
            Update Password
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { Lock, Key } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const validateConfirm = (_rule: any, value: string, callback: any) => {
  if (value === '') {
    callback(new Error('Please confirm your new password'))
  } else if (value !== form.newPassword) {
    callback(new Error('Passwords do not match'))
  } else {
    callback()
  }
}

const validateNewPassword = (_rule: any, value: string, callback: any) => {
  if (value === '') {
    callback(new Error('Please enter new password'))
  } else if (value === form.oldPassword) {
    callback(new Error('New password must be different from old password'))
  } else {
    if (form.confirmPassword !== '') {
      formRef.value?.validateField('confirmPassword')
    }
    callback()
  }
}

const rules: FormRules = {
  oldPassword: [
    { required: true, message: 'Please enter old password', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, validator: validateNewPassword, trigger: 'blur' },
    { min: 4, message: 'Password length should be at least 4 characters', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, validator: validateConfirm, trigger: 'blur' }
  ]
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await authStore.changePassword(form.oldPassword, form.newPassword)
        ElMessage.success('Password changed successfully')
        router.push('/dashboard')
      } catch (error: unknown) {
        ElMessage.error((error instanceof Error ? error.message : null) || 'Failed to change password')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.change-password-page {
  width: 100%;
}

.box-card {
  border-radius: 8px;
}

.card-header .title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
</style>
