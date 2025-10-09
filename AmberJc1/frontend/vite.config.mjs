import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173, // 前端运行端口
    proxy: {
      // 代理后端接口
      '/api': {
        target: 'http://localhost:8080', // 你的 Go 后端地址
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/api/v1'), // 将前端 /api 替换成 /api/v1
      },
    },
  },
})

