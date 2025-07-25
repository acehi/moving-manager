import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    // 调整代理目标端口为3000以匹配后端实际运行端口
    proxy: {
      '/api': {
        target: 'http://localhost:8080', // 后端服务实际运行在本地8080端口
        changeOrigin: true,
        // 后端路由已包含/api前缀，无需rewrite
        // rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
})
