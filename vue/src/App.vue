<template>
  <el-config-provider>
    <router-view />
  </el-config-provider>
</template>

<script setup vapor lang="ts">
import { onMounted } from 'vue'
import { useThemeStore } from './stores/theme'

const themeStore = useThemeStore()

onMounted(() => {
  themeStore.initTheme()
})
</script>
