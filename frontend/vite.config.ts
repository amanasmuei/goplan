import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    // Target modern browsers for smaller bundle size and modern JS features
    target: 'es2020',
    // Increase chunk size warning limit (default 500kB)
    chunkSizeWarningLimit: 600,
    // Enable source maps for production debugging (can be disabled if not needed)
    sourcemap: false,
    // Minification settings
    minify: 'esbuild',
    // CSS code splitting
    cssCodeSplit: true,
    rollupOptions: {
      output: {
        manualChunks: {
          // Core React runtime - changes infrequently, cached long-term
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          // Data fetching layer
          'vendor-query': ['@tanstack/react-query'],
          // State management
          'vendor-state': ['zustand'],
          // UI icons library
          'vendor-ui': ['lucide-react'],
          // HTTP client
          'vendor-http': ['axios'],
        },
      },
    },
  },
})
