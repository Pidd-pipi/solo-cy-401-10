import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 28030,
    proxy: {
      '/api': { target: 'http://localhost:29068', changeOrigin: true },
      '/uploads': { target: 'http://localhost:29068', changeOrigin: true }
    }
  },
  build: { outDir: 'dist' }
});
