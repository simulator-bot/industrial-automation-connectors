import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => ({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': loadEnv(mode, '.', 'VITE_').VITE_PROCESS_URL || 'http://127.0.0.1:8080',
    },
  },
}))
