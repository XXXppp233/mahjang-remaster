import { mkdir, readFile, rename, stat, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const frontendDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const publicDir = path.join(frontendDir, 'public')
const assetsFile = process.env.ASSETS_FILE
  ? path.resolve(process.env.ASSETS_FILE)
  : path.join(publicDir, 'assets.json')
const requestedConcurrency = Number.parseInt(process.env.ASSET_DOWNLOAD_CONCURRENCY ?? '8', 10)
const concurrency = Number.isFinite(requestedConcurrency) ? Math.max(1, requestedConcurrency) : 8

const assetKinds = {
  head: 'webp',
  full: 'webp',
  confirm: 'webm',
  critical: 'webm',
}

function safeAssetName(name) {
  if (!name || name === '.' || name === '..' || name.includes('/') || name.includes('\\')) {
    throw new Error(`Invalid asset name: ${JSON.stringify(name)}`)
  }
  return name
}

function remoteUrl(base, kind, organization, name, extension) {
  const root = base.endsWith('/') ? base : `${base}/`
  return new URL(
    [kind, organization, `${name}.${extension}`].map(encodeURIComponent).join('/'),
    root,
  )
}

async function exists(file) {
  try {
    await stat(file)
    return true
  } catch (error) {
    if (error.code === 'ENOENT') return false
    throw error
  }
}

async function download({ url, destination }) {
  if (process.env.FORCE_ASSET_DOWNLOAD !== '1' && await exists(destination)) return 'skipped'

  const response = await fetch(url)
  if (!response.ok) {
    throw new Error(`${url}: ${response.status} ${response.statusText}`)
  }

  const body = Buffer.from(await response.arrayBuffer())
  await mkdir(path.dirname(destination), { recursive: true })
  const temporary = `${destination}.tmp-${process.pid}`
  await writeFile(temporary, body)
  await rename(temporary, destination)
  return 'downloaded'
}

async function main() {
  const assets = JSON.parse(await readFile(assetsFile, 'utf8'))
  const characterMap = assets?.CharactersMap
  const urls = characterMap?.Urls
  const indexes = characterMap?.Index

  if (!Array.isArray(urls) || !Array.isArray(indexes)) {
    throw new Error(`${assetsFile} does not contain CharactersMap.Urls and CharactersMap.Index arrays`)
  }

  const jobs = []
  indexes.forEach((index, indexNumber) => {
    const base = urls[indexNumber] ?? urls[0]
    if (!base) throw new Error(`Missing asset URL for index ${indexNumber}`)

    for (const [kind, extension] of Object.entries(assetKinds)) {
      for (const [organization, names] of Object.entries(index?.[kind] ?? {})) {
        safeAssetName(organization)
        if (!Array.isArray(names)) throw new Error(`Invalid ${kind} list for ${organization}`)

        for (const rawName of names) {
          const name = safeAssetName(rawName)
          jobs.push({
            url: remoteUrl(base, kind, organization, name, extension),
            destination: path.join(publicDir, kind, organization, `${name}.${extension}`),
          })
        }
      }
    }
  })

  let nextJob = 0
  let downloaded = 0
  let skipped = 0
  async function worker() {
    while (nextJob < jobs.length) {
      const job = jobs[nextJob++]
      const result = await download(job)
      if (result === 'downloaded') downloaded += 1
      else skipped += 1
      console.log(`${result}: ${job.destination.replace(`${publicDir}${path.sep}`, '')}`)
    }
  }

  await Promise.all(Array.from({ length: Math.max(1, concurrency) }, worker))
  console.log(`Assets ready: ${downloaded} downloaded, ${skipped} already present.`)
}

main().catch((error) => {
  console.error(`Asset download failed: ${error.message}`)
  process.exitCode = 1
})
