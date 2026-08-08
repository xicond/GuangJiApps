import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import Login from '../views/Login.vue'
import MainLayout from '../layouts/MainLayout.vue'
// import PlaceholderView from '../views/PlaceholderList.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: Login,
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      component: MainLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/dashboard'
        },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('../views/Dashboard.vue'),
          meta: { title: 'Dashboard' }
        },
        {
          path: 'change-password',
          name: 'change-password',
          component: () => import('../views/ChangePasswordView.vue'),
          meta: { title: 'Change Password' }
        },
        // User Management
        {
          path: 'user-management/admin',
          name: 'admin',
          component: () => import('../views/user-management/AdminList.vue'),
          meta: { title: 'Admin' }
        },
        {
          path: 'user-management/admin-group',
          name: 'admin-group',
          component: () => import('../views/user-management/AdminGroupList.vue'),
          meta: { title: 'Admin Group' }
        },
        {
          path: 'user-management/group-menu-mapping',
          name: 'group-menu-mapping',
          component: () => import('../views/user-management/GroupMenuMappingList.vue'),
          meta: { title: 'Group Menu Mapping' }
        },
        {
          path: 'user-management/admin-sub-warehouse',
          name: 'admin-sub-warehouse',
          component: () => import('../views/user-management/AdminSubWarehouseList.vue'),
          meta: { title: 'Admin Sub Warehouse' }
        },
        // Master Data
        {
          path: 'master-data/umat',
          name: 'umat-list',
          component: () => import('../views/master-data/UmatList.vue'),
          meta: { title: 'Daftar Umat' }
        },
        {
          path: 'master-data/umat/create',
          name: 'umat-create',
          component: () => import('../views/master-data/UmatCreateList.vue'),
          meta: { title: 'Tambah Umat', activeMenu: '/master-data/umat' }
        },
        {
          path: 'master-data/umat/edit/:id',
          name: 'umat-edit',
          component: () => import('../views/master-data/UmatEditList.vue'),
          meta: { title: 'Edit Umat', activeMenu: '/master-data/umat' }
        },
        {
          path: 'master-data/topic',
          name: 'topic',
          component: () => import('../views/master-data/TopicList.vue'),
          meta: { title: 'Topic' }
        },
        {
          path: 'master-data/activity',
          name: 'activity',
          component: () => import('../views/master-data/ActivityList.vue'),
          meta: { title: 'Activity' }
        },
        {
          path: 'master-data/tim-kerja',
          name: 'tim-kerja',
          component: () => import('../views/master-data/TimKerjaList.vue'),
          meta: { title: 'Tim Kerja' }
        },
        {
          path: 'master-data/tahun-ciu-tao',
          name: 'tahun-ciu-tao',
          component: () => import('../views/master-data/TahunCiuTaoList.vue'),
          meta: { title: 'Tahun Ciu Tao' }
        },
        {
          path: 'master-data/penggalang-dana',
          name: 'penggalang-dana',
          component: () => import('../views/master-data/PenggalangDanaList.vue'),
          meta: { title: 'Penggalang Dana' }
        },
        {
          path: 'master-data/sxy-donatur',
          name: 'sxy-donatur',
          component: () => import('../views/master-data/SxyDonaturList.vue'),
          meta: { title: 'Sxy Donatur' }
        },
        // Transaction
        {
          path: 'transaction/kelas',
          name: 'kelas',
          component: () => import('../views/transaction/KelasList.vue'),
          meta: { title: 'Kelas' }
        },
        {
          path: 'transaction/kelas/create',
          name: 'kelas-create',
          component: () => import('../views/transaction/KelasCreate.vue'),
          meta: { title: 'Tambah Kelas', activeMenu: '/transaction/kelas' }
        },
        {
          path: 'transaction/kelas/edit/:id',
          name: 'kelas-edit',
          component: () => import('../views/transaction/KelasEdit.vue'),
          meta: { title: 'Edit Kelas', activeMenu: '/transaction/kelas' }
        },
        {
          path: 'transaction/kelas/view/:id',
          name: 'kelas-view',
          component: () => import('../views/transaction/KelasView.vue'),
          meta: { title: 'Detail Kelas', activeMenu: '/transaction/kelas' }
        },
        {
          path: 'transaction/donasi-sxy',
          name: 'donasi-sxy',
          component: () => import('../views/transaction/DonasiSxyList.vue'),
          meta: { title: 'Donasi Sxy' }
        },
        // Report
        {
          path: 'report/master/umat',
          name: 'report-umat',
          component: () => import('../views/report/UmatReportList.vue'),
          meta: { title: 'Laporan Master Umat' }
        },
        {
          path: 'report/master/tcs',
          name: 'report-tcs',
          component: () => import('../views/report/MasterReportList.vue'),
          meta: { title: 'Report Tcs' }
        },
        {
          path: 'report/master/umat-cheng-chien',
          name: 'report-umat-cheng-chien',
          component: () => import('../views/report/MasterReportList.vue'),
          meta: { title: 'Laporan Umat Cheng Chien' }
        },
        {
          path: 'report/master/lagu',
          name: 'report-lagu',
          component: () => import('../views/report/MasterReportList.vue'),
          meta: { title: 'Lagu Suci' }
        },
        {
          path: 'report/sxy',
          name: 'report-sxy',
          component: () => import('../views/report/SxyReportList.vue'),
          meta: { title: 'Sxy Report' }
        }
      ]
    }
  ]
})

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  const isAuth = authStore.isAuthenticated

  if (to.meta.requiresAuth !== false && !isAuth) {
    next({ name: 'login' })
  } else if (to.name === 'login' && isAuth) {
    next({ name: 'dashboard' })
  } else {
    next()
  }
})

export default router
