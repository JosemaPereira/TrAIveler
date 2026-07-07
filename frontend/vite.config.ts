/**
 * Vite configuration for TrAIveler frontend.
 * 
 * Key configurations:
 * - React plugin with Fast Refresh for development
 * - Path alias '@' → './src' for cleaner imports
 * - Development server on port 5173 (strictPort prevents fallback)
 * 
 * Path alias usage example:
 *   import { Button } from '@/components/primitives/Button'
 * instead of:
 *   import { Button } from '../../../components/primitives/Button'
 * 
 * @see https://vite.dev/config/
 */
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath } from 'node:url'

export default defineConfig({
  // Enable React Fast Refresh and JSX transformation
  plugins: [react()],
  
  resolve: {
    alias: {
      // Path alias for cleaner imports: '@/components/...' instead of '../../../components/...'
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  
  server: {
    port: 5173,
    // strictPort: true prevents Vite from trying alternative ports if 5173 is busy
    // This ensures consistent port usage across development environments
    strictPort: true,
  },
})
