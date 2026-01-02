import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
    plugins: [svelte()],
    base: './',
    build: {
        outDir: 'dist',
        emptyOutDir: true,
        minify: 'terser',
        terserOptions: {
            compress: {
                drop_console: true,
                drop_debugger: true,
            },
        },
        rollupOptions: {
            output: {
                manualChunks: {
                    'monaco-editor-core': ['monaco-editor/esm/vs/editor/editor.api'],
                    'monaco-languages': [
                        'monaco-editor/esm/vs/language/json/json.worker',
                        'monaco-editor/esm/vs/language/css/css.worker',
                        'monaco-editor/esm/vs/language/html/html.worker',
                        'monaco-editor/esm/vs/language/typescript/ts.worker',
                    ],
                },
            },
        },
        chunkSizeWarningLimit: 5000,
    },
    server: {
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            }
        }
    }
})
