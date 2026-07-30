<template>
  <div v-if="errors && errors.length > 0" class="field-errors-list">
    <div v-for="(errMsg, idx) in errors" :key="idx" class="field-error-item">
      <span class="error-bullet">&bull;</span>
      <span>{{ errMsg }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  errors?: string[]
}>()
</script>

<style scoped>
.field-errors-list {
  margin-top: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 100%;
}

.field-error-item {
  color: var(--el-color-danger, #f56c6c);
  font-size: 0.75rem;
  line-height: 1.25;
  display: flex;
  align-items: flex-start;
  gap: 4px;
}

.error-bullet {
  font-weight: bold;
}
</style>
