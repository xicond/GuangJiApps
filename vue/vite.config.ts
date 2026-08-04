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
          enabled: true,
          type: 'classic',
          suppressWarnings: false, 
          navigateFallbackAllowlist: [/^\//], // Allows caching navigation routes
        },
        workbox: {
          navigateFallback: 'index.html',
          globPatterns: ['**/*.{js,ts,css,html,ico,png,svg,json}']
        },
        injectManifest: {
          globPatterns: ['**/*.{js,ts,css,html,ico,png,svg,json}']
        },
        manifest: {
          name: 'GuangJi Apps',
          short_name: 'GuangJi',
          description: 'GuangJi Application',
          theme_color: '#409eff',
          background_color: '#141414',
          display: 'standalone',
          icons: [
            {
              src: 'favicon.png',
              sizes: '192x192',
              type: 'image/png',
              purpose: 'any'
            },
          /* {
            src: 'favicon_512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable'
          }, */
            {
              src: 'favicon.ico',
              sizes: '192x192',
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
