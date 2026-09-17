<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapsed || isMobile ? '54px' : '240px'" class="aside">
      <div class="sidebar-logo">
        <el-icon class="logo-icon">
          <Grid />
        </el-icon>
        <span v-if="!isCollapsed && !isMobile" class="logo-text">GuangJi Apps</span>
      </div>

      <el-scrollbar class="menu-scrollbar">
        <el-menu ref="menuRef" :default-active="activeMenu" :unique-opened="true" :collapse="isCollapsed || isMobile"
          :menu-trigger="isCollapsed || isMobile || isTablet ? 'click' : 'hover'" :router="true"
          class="el-menu-vertical">
          <!-- Dashboard -->
          <el-menu-item index="/dashboard">
            <el-icon>
              <HomeFilled />
            </el-icon>
            <template #title>Dashboard</template>
          </el-menu-item>

          <!-- User Management -->
          <el-sub-menu v-if="hasMenu('User Management')" index="user-management">
            <template #title>
              <el-icon>
                <User />
              </el-icon>
              <span>User Management</span>
            </template>
            <el-menu-item v-if="hasSubMenu('User Management', 'Create User')"
              index="/user-management/admin">Admin</el-menu-item>
            <el-menu-item v-if="hasSubMenu('User Management', 'Group Control')"
              index="/user-management/admin-group">Admin
              Group</el-menu-item>
            <el-menu-item v-if="hasSubMenu('User Management', 'Menu Control')"
              index="/user-management/group-menu-mapping">Group
              Menu Mapping</el-menu-item>
            <el-menu-item v-if="hasSubMenu('User Management', 'User Sub Warehouse Control')"
              index="/user-management/admin-sub-warehouse">Admin Sub Warehouse</el-menu-item>
          </el-sub-menu>

          <!-- Master Data -->
          <el-sub-menu v-if="hasMenu('Master Data')" index="master-data">
            <template #title>
              <el-icon>
                <Files />
              </el-icon>
              <span>Master Data</span>
            </template>
            <el-menu-item v-if="hasSubMenu('Master Data', 'Umat')" index="/master-data/umat">Umat</el-menu-item>
            <el-menu-item v-if="hasSubMenu('Master Data', 'Topik')" index="/master-data/topic">Topic</el-menu-item>
            <el-menu-item v-if="hasSubMenu('Master Data', 'Kelas')" index="/master-data/kelas">Kelas</el-menu-item>
            <el-menu-item v-if="hasSubMenu('Master Data', 'Kegiatan')"
              index="/master-data/activity">Activity</el-menu-item>
            <!-- <el-menu-item v-if="hasSubMenu('Master Data', 'Tim Kerja')" index="/master-data/tim-kerja">Tim
              Kerja</el-menu-item> -->
            <el-menu-item v-if="hasSubMenu('Master Data', 'Tahun Ciu Tao')" index="/master-data/tahun-ciu-tao">Tahun Ciu
              Tao</el-menu-item>
            <el-menu-item v-if="hasSubMenu('Master Data', 'Penggalang Dana')"
              index="/master-data/penggalang-dana">Penggalang
              Dana</el-menu-item>
            <el-menu-item v-if="hasSubMenu('Master Data', 'SXY Donatur')" index="/master-data/sxy-donatur">Sxy
              Donatur</el-menu-item>
          </el-sub-menu>

          <!-- Transaction -->
          <el-sub-menu v-if="hasMenu('Transaction')" index="transaction">
            <template #title>
              <el-icon>
                <Tickets />
              </el-icon>
              <span>Transaction</span>
            </template>
            <!-- <el-menu-item v-if="hasSubMenu('Transaction', 'Kegiatan')"
              index="/master-data/activity">Activity</el-menu-item> -->
            <el-menu-item v-if="hasSubMenu('Transaction', 'Kelas')" index="/transaction/kelas">Kelas</el-menu-item>
            <el-menu-item v-if="hasSubMenu('Transaction', 'Donasi SXY')" index="/transaction/donasi-sxy">Donasi
              Sxy</el-menu-item>
          </el-sub-menu>

          <!-- Report -->
          <el-sub-menu v-if="hasMenu('Report')" index="report">
            <template #title>
              <el-icon>
                <DataAnalysis />
              </el-icon>
              <span>Report</span>
            </template>
            <el-sub-menu v-if="hasSubMenu('Report', 'Master')" index="report-master" ref="reportMasterRef"
              @click.stop="toggleReportMaster">
              <template #title>
                <span>Master</span>
              </template>

              <!-- Level paling bawah menggunakan el-menu-item -->
              <el-menu-item index="/report/master/umat">Umat</el-menu-item>
              <el-menu-item index="/report/master/tcs">Tcs</el-menu-item>
              <!-- <el-menu-item index="/report/master/umat-cheng-chien">Umat Cheng Cien</el-menu-item>
              <el-menu-item index="/report/master/lagu">Lagu</el-menu-item> -->
            </el-sub-menu>
            <el-menu-item v-if="hasSubMenu('Report', 'SXY')" index="/report/sxy">Sxy</el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-scrollbar>
    </el-aside>

    <el-container class="content-container">
      <el-header class="header">
        <div class="header-left">
          <el-button v-if="!isMobile" circle text size="large" :icon="isCollapsed || isCollapsed ? Expand : Fold"
            @click="isCollapsed = !isCollapsed" />
        </div>

        <div class="header-right">
          <el-dropdown @command="handleThemeChange" trigger="click">
            <el-button circle :icon="currentIcon" />
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="system">
                  <el-icon>
                    <Monitor />
                  </el-icon> System Mode
                </el-dropdown-item>
                <el-dropdown-item command="light">
                  <el-icon>
                    <Sunny />
                  </el-icon> Light Mode
                </el-dropdown-item>
                <el-dropdown-item command="dark">
                  <el-icon>
                    <Moon />
                  </el-icon> Dark Mode
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-dropdown trigger="click" @command="handleUserMenuCommand">
            <span class="user-profile">
              <el-avatar :size="32" :icon="UserFilled" />
              <span class="username">{{ usernameDisplay }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="change-password">
                  <el-icon>
                    <Key />
                  </el-icon> Change Password
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon>
                    <SwitchButton />
                  </el-icon> Logout
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main-body">
        <router-view />
      </el-main>
    </el-container>

    <ChangePasswordModal v-model="isChangePasswordVisible" />
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useBreakpoints, breakpointsTailwind, useOnline } from '@vueuse/core'
import { useRoute, useRouter } from 'vue-router'
import { ElNotification } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { setMode, currentMode, isDark, type ThemeMode } from '../theme'
import ChangePasswordModal from '../components/ChangePasswordModal.vue'
import { useQuickStreamSpeedTest } from '../api/useQuickStreamSpeedTest'
import {
  Grid,
  HomeFilled,
  User,
  Files,
  Tickets,
  DataAnalysis,
  Fold,
  Expand,
  UserFilled,
  Key,
  SwitchButton,
  Sunny,
  Moon,
  Monitor
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 1. State Jaringan & Notifikasi
const isOnline = useOnline()
const isRealInternet = ref(true)
let offlineNotificationHandle: ReturnType<typeof ElNotification> | null = null

// 2. Fungsi Cek Koneksi Kilat ke Cloudflare Jakarta
const cekInternetKilat = async () => {
  if (!isOnline.value) {
    isRealInternet.value = false
    return false
  }

  const nav = navigator as any
  const koneksi = nav.connection || nav.mozConnection || nav.webkitConnection
  let koneksiType = 'wifi'
  console.log("connnection type", koneksi.type)
  if (koneksi && koneksi.type) {
    koneksiType = koneksi.type === 'cellular' ? 'cell' : 'wifi'
  }

  if (koneksiType === 'cell') {

    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), 200) // Timeout 200ms

    try {
      await fetch("https://cp.cloudflare.com/generate_204", {
        method: "HEAD",
        mode: "no-cors",
        cache: "no-store",
        signal: controller.signal
      })
      clearTimeout(timer)
      return true // Internet aktif
    } catch (err) {
      clearTimeout(timer)
      return false // Terblokir iOS atau masalah jaringan
    }
  }

  return isOnline.value

}

// 3. Logika Utama Pengelolaan Notifikasi
const kelolaNotifikasiBlokir = async () => {
  const terhubung = await cekInternetKilat()
  isRealInternet.value = terhubung

  if (!terhubung || !isOnline.value) {
    // KONDISI: Safari diblokir (Sistem online, internet nyata mati)
    if (!offlineNotificationHandle) {
      offlineNotificationHandle = ElNotification.warning({
        title: 'Koneksi Terputus',
        message: 'Anda sedang offline. Beberapa fitur mungkin terbatas.',
        duration: 0, // Tidak hilang otomatis
        showClose: true,
        onClose: () => { offlineNotificationHandle = null }
      })
    }
  } else if (terhubung) {
    // KONDISI: Internet pulih / Tidak diblokir
    if (offlineNotificationHandle) {
      offlineNotificationHandle.close()
      offlineNotificationHandle = null

      // Opsional: Tampilkan notifikasi sukses singkat bahwa internet pulih
      // ElNotification.success({
      //   title: 'Terhubung Kembali',
      //   message: 'Akses internet Safari telah aktif.',
      //   duration: 3000
      // })
    }
  }
}

// 4. Handler saat Browser Kembali Fokus
const handleBrowserFocus = () => {
  // Hanya jalankan cek jika status dokumen terlihat (active/foreground)
  if (document.visibilityState === 'visible') {
    console.log("User kembali ke Safari, memeriksa status izin internet...")
    kelolaNotifikasiBlokir()
  }
}

const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('md')   // True if width < 768px
// const isDesktop = breakpoints.greaterOrEqual('lg')  // True if width >= 1024px
const isTablet = breakpoints.between('md', 'lg')

watch(isOnline, () => {
  if (!isMobile.value)
    kelolaNotifikasiBlokir();
}, { immediate: true }) // immediate: true untuk langsung mengecek status saat komponen dimuat

// 5. Lifecycle Hooks untuk Event Listener
onMounted(() => {
  if (isMobile.value) return;

  // Jalankan cek pertama kali saat halaman dimuat
  kelolaNotifikasiBlokir()

  // visibilitychange mendeteksi tab pindah background/foreground (Sangat akurat di iOS)
  document.addEventListener('visibilitychange', handleBrowserFocus)
  // window focus sebagai cadangan penguat interaksi
  window.addEventListener('focus', handleBrowserFocus)
})

onUnmounted(() => {
  document.removeEventListener('visibilitychange', handleBrowserFocus)
  window.removeEventListener('focus', handleBrowserFocus)

  if (offlineNotificationHandle) {
    offlineNotificationHandle.close()
  }
})

// 6. Watcher untuk transisi offline-to-online dasar sistem
// watch(isOnline, (newStatus, oldStatus) => {
//   if (newStatus === true && oldStatus === false) {
//     kelolaNotifikasiBlokir()
//   } else if (newStatus === false) {
//     isRealInternet.value = false
//     if (offlineNotificationHandle) {
//       offlineNotificationHandle.close()
//       offlineNotificationHandle = null
//     }
//   }
// })

const { speedMbps, lastTestCompletedAt } = useQuickStreamSpeedTest()
let lastNotificationTime = 0
const NOTIFY_THROTTLE_MS = 5000



watch(lastTestCompletedAt, () => {


  if (speedMbps.value !== null && speedMbps.value < 1.6) {
    const now = Date.now()
    if (now - lastNotificationTime >= NOTIFY_THROTTLE_MS) {
      lastNotificationTime = now
      ElNotification.warning({
        title: 'Slow Connection',
        message: `Download speed is slow (${speedMbps.value == 0 ? '< 1 Mbps' : `${speedMbps.value} Mbps`}). Some processes will use cache.`,
        duration: 7000,
        position: 'top-right'
      })
    }
  }
})

const isChangePasswordVisible = ref(false)


const menuRef = ref()
const reportMasterRef = ref()

const toggleReportMaster = (e?: Event) => {
  if (isCollapsed.value || isMobile.value || isTablet.value) {
    if (e) {
      e.stopPropagation()
    }
    const isOpen = !!reportMasterRef.value?.opened
    if (isOpen) {
      menuRef.value?.close('report-master')
    } else {
      menuRef.value?.open('report-master')
    }
  }
}

const isCollapsed = ref(isTablet.value)
const activeMenu = computed(() => {
  if (route.meta?.activeMenu) {
    return route.meta.activeMenu as string
  }
  const path = route.path
  const parentPath = path.replace(/\/(create|edit|view)(\/.*)?$/, '')
  return parentPath || path
})
const usernameDisplay = computed(() => authStore.user?.username || authStore.user?.Username || 'User')
const hasMenu = (menuName: string) => authStore.hasMenu(menuName)
const hasSubMenu = (parentMenuName: string, subMenuName: string) => authStore.hasSubMenu(parentMenuName, subMenuName)

const currentIcon = computed(() => {
  if (currentMode.value === 'system') return Monitor
  return isDark.value ? Moon : Sunny
})

const handleThemeChange = (mode: ThemeMode) => {
  setMode(mode)
}

const handleUserMenuCommand = (command: string) => {
  if (command === 'change-password') {
    isChangePasswordVisible.value = true
  } else if (command === 'logout') {
    handleLogout()
  }
}

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style>
:root {
  --el-menu-base-level-padding: 15px;
}
</style>
<style scoped>
.layout-container {
  min-height: 100vh;
  width: 100%;
}

.aside {
  background-color: var(--el-bg-color-overlay);
  border-right: 1px solid var(--el-border-color-light);
  transition: width 0.3s ease;
  display: flex;
  flex-direction: column;
}

.sidebar-logo {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 0.89rem;
  gap: 0.75rem;
  border-bottom: 1px solid var(--el-border-color-light);
  overflow: hidden;
  white-space: nowrap;
}

.logo-icon {
  font-size: 1.5rem;
  color: var(--el-color-primary);
}

.logo-text {
  font-size: 1.15rem;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.menu-scrollbar {
  flex: 1;
}

.el-menu-vertical {
  border-right: none;
  background-color: transparent;
}

.content-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--el-border-color-light);
  background-color: var(--el-bg-color-overlay);
  padding: 0 1.25rem;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.username {
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.main-body {
  flex: 1;
  padding: 1.5rem;
  background-color: var(--el-bg-color-page);
}
</style>
