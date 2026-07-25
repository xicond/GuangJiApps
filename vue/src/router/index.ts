import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import LoginView from '../views/LoginView.vue'
import MainLayout from '../layouts/MainLayout.vue'
import DashboardView from '../views/DashboardView.vue'
import PlaceholderView from '../views/PlaceholderView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
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
          component: DashboardView,
          meta: { title: 'Dashboard' }
        },
        // User Management
        {
          path: 'user-management/admin',
          name: 'admin',
          component: () => import('../views/user-management/AdminView.vue'),
          meta: { title: 'Admin' }
        },
        {
          path: 'user-management/admin-group',
          name: 'admin-group',
          component: () => import('../views/user-management/AdminGroupView.vue'),
          meta: { title: 'Admin Group' }
        },
        {
          path: 'user-management/group-menu-mapping',
          name: 'group-menu-mapping',
          component: () => import('../views/user-management/GroupMenuMappingView.vue'),
          meta: { title: 'Group Menu Mapping' }
        },
        {
          path: 'user-management/admin-sub-warehouse',
          name: 'admin-sub-warehouse',
          component: () => import('../views/user-management/AdminSubWarehouseView.vue'),
          meta: { title: 'Admin Sub Warehouse' }
        },
        // Master Data
        {
          path: 'master-data/umat',
          name: 'umat-list',
          component: () => import('../views/master-data/UmatView.vue'),
          meta: { title: 'Daftar Umat' }
        },
        {
          path: 'master-data/umat/create',
          name: 'umat-create',
          component: () => import('../views/master-data/UmatCreateView.vue'),
          meta: { title: 'Tambah Umat' }
        },
        {
          path: 'master-data/umat/edit/:id',
          name: 'umat-edit',
          component: () => import('../views/master-data/UmatEditView.vue'),
          meta: { title: 'Edit Umat' }
        },
        {
          path: 'master-data/topic',
          name: 'topic',
          component: () => import('../views/master-data/TopicView.vue'),
          meta: { title: 'Topic' }
        },
        {
          path: 'master-data/activity',
          name: 'activity',
          component: () => import('../views/master-data/ActivityView.vue'),
          meta: { title: 'Activity' }
        },
        {
          path: 'master-data/tim-kerja',
          name: 'tim-kerja',
          component: () => import('../views/master-data/TimKerjaView.vue'),
          meta: { title: 'Tim Kerja' }
        },
        {
          path: 'master-data/tahun-ciu-tao',
          name: 'tahun-ciu-tao',
          component: () => import('../views/master-data/TahunCiuTaoView.vue'),
          meta: { title: 'Tahun Ciu Tao' }
        },
        {
          path: 'master-data/penggalang-dana',
          name: 'penggalang-dana',
          component: () => import('../views/master-data/PenggalangDanaView.vue'),
          meta: { title: 'Penggalang Dana' }
        },
        {
          path: 'master-data/sxy-donatur',
          name: 'sxy-donatur',
          component: () => import('../views/master-data/SxyDonaturView.vue'),
          meta: { title: 'Sxy Donatur' }
        },
        // Transaction
        {
          path: 'transaction/kelas',
          name: 'kelas',
          component: () => import('../views/transaction/KelasView.vue'),
          meta: { title: 'Kelas' }
        },
        {
          path: 'transaction/donasi-sxy',
          name: 'donasi-sxy',
          component: () => import('../views/transaction/DonasiSxyView.vue'),
          meta: { title: 'Donasi Sxy' }
        },
        // Report
        {
          path: 'report/master',
          name: 'report-master',
          component: () => import('../views/report/MasterReportView.vue'),
          meta: { title: 'Master Report' }
        },
        {
          path: 'report/sxy',
          name: 'report-sxy',
          component: () => import('../views/report/SxyReportView.vue'),
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
