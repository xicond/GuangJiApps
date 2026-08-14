<template>
  <el-dialog :model-value="modelValue" title="Change Password" :width="isMobile ? '90%' : '600px'"
    :before-close="handleClose" destroy-on-close append-to-body>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @keyup.enter="handleSubmit">
      <el-form-item label="Old Password" prop="oldPassword">
        <el-input v-model="form.oldPassword" type="password" show-password placeholder="Enter current password"
          :prefix-icon="Lock" />
      </el-form-item>

      <el-form-item label="New Password" prop="newPassword">
        <el-input v-model="form.newPassword" type="password" show-password placeholder="Enter new password"
          :prefix-icon="Key" />
      </el-form-item>

      <el-form-item label="Confirm New Password" prop="confirmPassword">
        <el-input v-model="form.confirmPassword" type="password" show-password placeholder="Confirm new password"
          :prefix-icon="Key" />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">Cancel</el-button>
        <el-button type="primary" :loading="loading" @click="handleSubmit">
          Change Password
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { Lock, Key } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'

const props = defineProps<{
  modelValue: boolean
}>()

const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')   // True if width < 768px

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}>()

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

watch(() => props.modelValue, (val) => {
  if (!val) {
    resetForm()
  }
})

const resetForm = () => {
  form.oldPassword = ''
  form.newPassword = ''
  form.confirmPassword = ''
  formRef.value?.resetFields()
}

const handleClose = () => {
  resetForm()
  emit('update:modelValue', false)
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await authStore.changePassword(form.oldPassword, form.newPassword)
        ElMessage.success('Password changed successfully')
        emit('success')
        handleClose()
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
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
