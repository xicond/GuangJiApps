import { ref } from 'vue'

const speedMbps = ref<number | null>(null)
const isTesting = ref<boolean>(false)
const lastTestCompletedAt = ref<number>(0)

let lastTestTime = 0
const THROTTLE_MS = 5000

export function useQuickStreamSpeedTest() {
    const measureSpeedInOneSecond = async (
        timeLimitMs = 300, // Batasi tepat 300ms
        fileUrl = 'https://speed.cloudflare.com/__down?bytes=280000' // Gunakan file besar agar tidak habis duluan
    ) => {
        const now = Date.now()
        if (isTesting.value || (now - lastTestTime < THROTTLE_MS)) {
            return
        }
        lastTestTime = now
        isTesting.value = true

        const controller = new AbortController()
        const startTime = performance.now()
        let receivedBytes = 0

        // Timer untuk membatalkan fetch tepat setelah timeLimitMs
        const timer = setTimeout(() => {
            controller.abort()
        }, timeLimitMs)

        try {
            const response = await fetch(`${fileUrl}&t=${Date.now()}`, {
                cache: 'no-store',
                mode: 'cors',
                signal: controller.signal // Menghubungkan pembatalan ke fetch
            })

            if (!response.body) throw new Error('ReadableStream tidak didukung.')

            const reader = response.body.getReader()

            // Membaca stream byte per byte / chunk per chunk
            while (true) {
                const { done, value } = await reader.read()
                if (done) break
                if (value) {
                    receivedBytes += value.length // Akumulasi ukuran chunk (dalam Bytes)
                }
            }
        } catch (error: unknown) {
            // AbortError adalah hal yang disengaja saat timer 1 detik memutus koneksi
            if (error instanceof Error && error.name !== 'AbortError') {
                console.error('Speed test error:', error)
            }
        } finally {
            clearTimeout(timer)
            const endTime = performance.now()

            // Hitung durasi aktual (mendekati 1000ms)
            const durationSeconds = (endTime - startTime) / 1000

            if (receivedBytes > 0 && durationSeconds > 0) {
                // Rumus: (Total Bytes * 8 bit) / Durasi (detik) / 1,000,000
                const totalBits = receivedBytes * 8
                const mbps = totalBits / durationSeconds / 1000000
                speedMbps.value = parseFloat(mbps.toFixed(2))
                console.warn(`[Network Monitor] Speed test ${speedMbps.value}`)

            } else {
                speedMbps.value = 0
            }
            isTesting.value = false
            lastTestCompletedAt.value = Date.now()
        }
    }

    const canRunTest = () => {
        return !isTesting.value && (Date.now() - lastTestTime >= THROTTLE_MS)
    }

    return {
        speedMbps,
        isTesting,
        lastTestCompletedAt,
        canRunTest,
        measureSpeedInOneSecond
    }
}