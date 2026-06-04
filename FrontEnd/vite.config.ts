import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import fs from 'node:fs'
import path from 'node:path'
import type { ViteDevServer, Plugin } from 'vite'
import type { IncomingMessage, ServerResponse } from 'node:http'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
// Custom plugin to expose a music list endpoint in dev and emit a manifest for prod
function musicListPlugin(): Plugin {
  const musicDir = path.resolve(__dirname, 'public', 'music')
  const manifestName = 'music-list.json'
  const exts = new Set(['.mp3', '.m4a', '.m4s', '.ogg', '.wav'])

  function buildList() {
    let files: string[] = []
    try {
      files = fs.readdirSync(musicDir)
    } catch (e) {
      return []
    }
    return files
      .filter((f) => exts.has(path.extname(f).toLowerCase()))
      .map((f) => {
        const base = f.replace(/\.[^.]+$/, '')
        return {
          file: f,
          name: decodeURIComponent(base),
          url: `/music/${encodeURIComponent(f)}`,
        }
      })
  }

  return {
    name: 'music-list-plugin',
    configureServer(server: ViteDevServer) {
      server.middlewares.use('/api/music-list', (req: IncomingMessage, res: ServerResponse) => {
        const list = buildList()
        res.setHeader('Content-Type', 'application/json; charset=utf-8')
        res.end(JSON.stringify(list))
      })
    },
    buildStart() {
      // Emit /public/music-list.json so production can fetch without a server API
      const list = buildList()
      const outPath = path.resolve(__dirname, 'public', manifestName)
      try {
        fs.writeFileSync(outPath, JSON.stringify(list, null, 2), 'utf-8')
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('[music-list-plugin] Failed to write manifest', e)
      }
    },
  }
}

// New: Characters Head plugin – lists folders and images under public/characters/head
function charactersHeadPlugin(): Plugin {
  const baseDir = path.resolve(__dirname, 'public', 'characters', 'head')
  const manifestName = 'characters-head.json'
  const imageExts = new Set(['.png', '.jpg', '.jpeg', '.webp', '.gif', '.svg'])

  function getFolders(): string[] {
    try {
      return fs
        .readdirSync(baseDir)
        .filter((name) => !name.startsWith('.')) // exclude dot-prefixed folders
        .filter((name) => {
          const full = path.join(baseDir, name)
          try {
            return fs.statSync(full).isDirectory()
          } catch {
            return false
          }
        })
        .sort()
    } catch {
      return []
    }
  }

  function getImagesInFolder(folder: string): string[] {
    const safeFolder = folder.replace(/[\\/]/g, '')
    const dir = path.join(baseDir, safeFolder)
    let files: string[] = []
    try {
      files = fs.readdirSync(dir)
    } catch {
      return []
    }
    return files
      .filter((f) => imageExts.has(path.extname(f).toLowerCase()))
      .map((f) => `/characters/head/${encodeURIComponent(safeFolder)}/${encodeURIComponent(f)}`)
      .sort()
  }

  function buildManifest(): Array<{ folder: string; images: string[] }> {
    const folders = getFolders()
    return folders.map((folder) => ({ folder, images: getImagesInFolder(folder) }))
  }

  return {
    name: 'characters-head-plugin',
    configureServer(server: ViteDevServer) {
      server.middlewares.use(
        '/api/characters/head/folders',
        (_req: IncomingMessage, res: ServerResponse) => {
          const folders = getFolders()
          res.setHeader('Content-Type', 'application/json; charset=utf-8')
          res.end(JSON.stringify(folders))
        },
      )

      server.middlewares.use(
        '/api/characters/head/images',
        (req: IncomingMessage, res: ServerResponse) => {
          try {
            const url = new URL(req.url || '', 'http://localhost')
            const folder = url.searchParams.get('folder') || ''
            const images = getImagesInFolder(folder)
            res.setHeader('Content-Type', 'application/json; charset=utf-8')
            res.end(JSON.stringify(images))
          } catch (e) {
            res.statusCode = 400
            res.end('Bad Request')
          }
        },
      )
    },
    buildStart() {
      // Emit /public/characters-head.json for production fallback
      const manifest = buildManifest()
      const outPath = path.resolve(__dirname, 'public', manifestName)
      try {
        fs.writeFileSync(outPath, JSON.stringify(manifest, null, 2), 'utf-8')
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('[characters-head-plugin] Failed to write manifest', e)
      }
    },
  }
}

export default defineConfig({
  plugins: [vue(), vueJsx(), musicListPlugin(), charactersHeadPlugin()],
  server: {
    host: '0.0.0.0',
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
})
