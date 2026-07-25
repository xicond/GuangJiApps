import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig(({ mode }) => {
  // Muat file .env berdasarkan mode aktif
  const env = loadEnv(mode, process.cwd(), '')

  return {
    base: env.VITE_BASE_URL || '/',
    plugins: [
      vue(),
      VitePWA({
        registerType: 'autoUpdate',
        injectRegister: 'auto',
        strategies: 'injectManifest',
        srcDir: 'src',
        filename: 'sw.ts',
        devOptions: {
          enabled: false
        },
        workbox: {
          navigateFallback: 'index.html',
          globPatterns: ['**/*.{js,css,html,ico,png,svg,json}']
        },
        injectManifest: {
          globPatterns: ['**/*.{js,css,html,ico,png,svg,json}']
        },
        manifest: {
          name: 'GuangJi Apps',
          short_name: 'GuangJi',
          description: 'GuangJi PWA Application',
          theme_color: '#409eff',
          background_color: '#141414',
          display: 'standalone',
          icons: [
            {
              src: 'favicon.ico',
              sizes: '64x64 32x32 24x24 16x16',
              type: 'image/x-icon'
            }
          ]
        }
      })
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    server: {
      port: 5173,
      host: true
    }
  }
})
