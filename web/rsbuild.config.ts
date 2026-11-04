import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { defineConfig, loadEnv } from '@rsbuild/core'
import { pluginReact } from '@rsbuild/plugin-react'
import { pluginTailwindcss } from '@rsbuild/plugin-tailwindcss'
import { tanstackRouter } from '@tanstack/router-plugin/rspack'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig(({ envMode }) => {
  const env = loadEnv({ mode: envMode, prefixes: ['VITE_'] })
  const serverUrl =
    process.env.VITE_REACT_APP_SERVER_URL ||
    env.rawPublicVars.VITE_REACT_APP_SERVER_URL ||
    'http://localhost:3000'

  const isProd = envMode === 'production'
  const trimmedBasePath = (process.env.BASE_PATH ?? '')
    .trim()
    .replaceAll(/^\/+|\/+$/g, '')
  const compiledBasePath =
    isProd && trimmedBasePath ? `/${trimmedBasePath}` : ''
  const builtBase = compiledBasePath === '' ? '/' : `${compiledBasePath}/`

  return {
    plugins: [pluginReact(), pluginTailwindcss({ optimize: false })],
    source: {
      entry: {
        index: './src/main.tsx',
      },
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    html: {
      template: './index.html',
      templateParameters: { appBasePath: compiledBasePath },
    },
    server: {
      host: '0.0.0.0',
      port: 5173,
      strictPort: false,
      base: isProd ? builtBase : '/',
      proxy: {
        '/api': { target: serverUrl, changeOrigin: true },
      },
    },
    output: {
      minify: isProd,
      target: 'web',
      assetPrefix: compiledBasePath === '' ? 'auto' : builtBase,
      distPath: { root: 'dist' },
    },
    performance: {
      removeConsole: isProd ? ['log'] : false,
      buildCache: false,
    },
    tools: {
      rspack: {
        plugins: [
          tanstackRouter({
            target: 'react',
            autoCodeSplitting: isProd,
          }),
        ],
      },
    },
  }
})
