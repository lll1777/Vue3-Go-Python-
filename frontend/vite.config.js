import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api/prediction': {
        target: 'http://localhost:5000',
        changeOrigin: true
      },
      '/api/plate': {
        target: 'http://localhost:5000',
        changeOrigin: true
      },
      '/api/anomaly': {
        target: 'http://localhost:5000',
        changeOrigin: true
      },
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
