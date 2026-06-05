import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import fs from 'node:fs'
import path from 'node:path'
import type { ViteDevServer, Plugin } from 'vite'
import type { IncomingMessage, ServerResponse } from 'node:http'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'


export default defineConfig({
  plugins: [vue(), vueJsx(),],
  server: {
    host: '0.0.0.0',
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
})
