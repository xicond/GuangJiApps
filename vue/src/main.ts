import { createApp /* createVaporApp, vaporInteropPlugin, type VaporComponent */ } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
// import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'element-plus/es/components/notification/style/css'
import 'element-plus/es/components/message/style/css'
import 'element-plus/es/components/message-box/style/css'
import 'element-plus/es/components/loading/style/css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'
import './style.css'
import { initPwaUpdate } from './utils/pwaUpdate'
import * as Sentry from "@sentry/vue";

const app = createApp(App /* as unknown as VaporComponent */)
// app.use(vaporInteropPlugin)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

if (import.meta.env.VITE_SENTRY_DSN) {
  Sentry.init({
    app,
    dsn: import.meta.env.VITE_SENTRY_DSN,
    sendDefaultPii: false,
    // PENTING: Tandai sebagai development agar error lokal terisolasi
    environment: "development",

    // Kontrol apakah Sentry aktif di localhost (true jika ingin dites)
    enabled: true,

    // Aktifkan debug log di console browser untuk melihat status Sentry
    // debug: true,

    beforeSend(event) {
      // 1. Hapus informasi user / PII (username, email, IP) dari event Sentry
      delete event.user;

      // 2. Bersihkan parameter sensitif di URL (token, username, password)
      if (event.request && event.request.url) {
        event.request.url = event.request.url.replace(/([?&])(token|username|user|password|pass|secret)=[^&]+/gi, "$1$2=REDACTED");
      }

      // 3. Redaksi data sensitif di body request (username, password, token, dsb)
      if (event.request && event.request.data) {
        try {
          const data = typeof event.request.data === 'string'
            ? JSON.parse(event.request.data)
            : event.request.data;

          if (data && typeof data === 'object') {
            const sensitivePatterns = ['username', 'user', 'password', 'pass', 'token', 'secret', 'email', 'auth'];
            for (const key of Object.keys(data)) {
              if (sensitivePatterns.some(pattern => key.toLowerCase().includes(pattern))) {
                data[key] = "[REDACTED]";
              }
            }
          }

          event.request.data = JSON.stringify(data);
        } catch (e) {
          // Abaikan jika bukan JSON
        }
      }

      return event;
    },
    // Aktifkan integrasi tracing
    integrations: [
      Sentry.browserTracingIntegration({
        // 1. Integrasi otomatis dengan Vue Router untuk mengukur durasi perpindahan halaman
        router: router,

        // 2. Waktu tunggu (dalam ms) sebelum transaksi ditutup jika masih ada async operations/spans aktif
        idleTimeout: 3000,

        // 3. Menandai span sebagai 'cancelled' jika aplikasi masuk ke background (tab disembunyikan)
        markBackgroundSpan: true,

        // 4. Kustomisasi span sebelum dikirim
        beforeStartSpan: (context: any) => {
          if (context && typeof context.name === 'string') {
            context.name = context.name.replace(/\/\d+/, '/:id');
          }
          return context;
        },
      }),
      Sentry.browserProfilingIntegration()
    ],
    // Persentase sampel data performa yang dikirim ke Sentry (1.0 = 100% untuk testing dev)
    tracesSampleRate: 1.0,
    profilesSampleRate: 1.0,
    profileSessionSampleRate: 1.0, // 1.0 = 100% of sessions profiled
    profileLifecycle: "trace", // attach profiles to spans

    dataCollection: {
      // To disable sending user data and HTTP bodies, uncomment the lines below. For more info visit:
      // https://docs.sentry.io/platforms/javascript/guides/vue/configuration/options/#dataCollection
      // userInfo: false,
      // httpBodies: []
    }
  });

  // Helper testing di Console Browser: ketik testSentry() di devtools console
  if (import.meta.env.DEV) {
    (window as any).testSentry = () => {
      const id = Sentry.captureException(new Error("Test Sentry Error dari GuangJiApps Frontend"));
      console.log("[Sentry Test] Sent test error event ID:", id);
      return id;
    };
  }
  Sentry.metrics.gauge('page_load_time', 150);
  Sentry.metrics.distribution('response_time', 200);
}


app.use(createPinia())
app.use(router)
app.use(ElementPlus)

// Initialize Workbox PWA service worker and route update listener
initPwaUpdate(router)

app.mount('#app')
