import { ref, onMounted, onUnmounted } from 'vue'

export function useDevice() {
  const isMobile = ref(false)
  const isDesktop = ref(true)

  const checkDevice = () => {
    const width = window.innerWidth
    isMobile.value = width < 768
    isDesktop.value = width >= 1024
  }

  onMounted(() => {
    checkDevice()
    window.addEventListener('resize', checkDevice)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', checkDevice)
  })

  return { isMobile, isDesktop }
}