import { defineConfig, loadEnv, Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { VitePWA } from 'vite-plugin-pwa'
import { visualizer } from 'rollup-plugin-visualizer'
import { fileURLToPath, URL } from 'node:url'
import fs from 'node:fs'
import path from 'node:path'
import zlib from 'node:zlib'

function compressInPlacePlugin(): Plugin {
  return {
    name: 'vite-plugin-compress-in-place',
    enforce: 'post',
    apply: 'build',
    closeBundle() {
      const distDir = path.resolve(process.cwd(), 'dist')
      if (!fs.existsSync(distDir)) return

      const targetExts = [
        '.js', '.mjs', '.css', '.html', '.svg', '.png',
        '.jpg', '.jpeg', '.webp', '.ico', '.json', '.webmanifest'
      ]

      let count = 0
      const compressRecursive = (dir: string) => {
        const entries = fs.readdirSync(dir, { withFileTypes: true })
        for (const entry of entries) {
          const fullPath = path.join(dir, entry.name)
          if (entry.isDirectory()) {
            compressRecursive(fullPath)
          } else if (entry.isFile()) {
            const ext = path.extname(entry.name).toLowerCase()
            if (targetExts.includes(ext)) {
              const content = fs.readFileSync(fullPath)
              if (content.length > 0) {
                const compressed = zlib.gzipSync(content, { level: 9 })
                // Seamlessly overwrite original file in-place with gzip data
                fs.writeFileSync(fullPath, compressed)
                count++
              }
            }
          }
        }
      }

      compressRecursive(distDir)
      // console.log(`\x1b[36m[vite-plugin-compress-in-place]\x1b[0m Seamlessly compressed ${count} static assets in-place in dist/`)
    }
  }
}

export default defineConfig(({ mode }) => {
  // Muat file .env berdasarkan mode aktif
  const env = loadEnv(mode, process.cwd(), '')

  return {
    base: env.VITE_BASE_URL || '/',
    plugins: [
      vue(),
      AutoImport({
        resolvers: [ElementPlusResolver()],
      }),
      Components({
        resolvers: [ElementPlusResolver({
          importStyle: 'css', // Mengimpor file CSS spesifik HANYA untuk komponen yang dipanggil
        })],
      }),
      VitePWA({
        registerType: 'autoUpdate',
        injectRegister: 'auto',
        strategies: 'injectManifest',
        srcDir: 'src',
        filename: 'sw.ts',
        devOptions: {
          enabled: true,
          type: 'module',
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
      }),
      /* visualizer({
        open: true, // Otomatis membuka browser setelah build selesai
        filename: 'stats.html', // Nama file laporan output
        gzipSize: true, // Menampilkan ukuran setelah kompresi gzip
        brotliSize: true, // Menampilkan ukuran setelah kompresi brotli
        template: 'treemap', // Model visualisasi ('sunburst', 'network', atau 'treemap')
      }), */
      compressInPlacePlugin()
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
