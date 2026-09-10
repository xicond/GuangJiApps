<template>
  <el-config-provider>
    <router-view />
  </el-config-provider>
</template>

<script setup vapor lang="ts">
import { onMounted } from 'vue'
import { initTheme } from './theme'

onMounted(() => {
  initTheme()
})
</script>
