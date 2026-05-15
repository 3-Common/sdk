import { fileURLToPath } from 'node:url'

import { defineConfig } from 'vitest/config'

const root = fileURLToPath(new URL('.', import.meta.url))

export default defineConfig({
  resolve: {
    alias: [
      { find: /^@3common\/sdk$/u, replacement: `${root}src/index.ts` },
      { find: /^@\/(.+)$/u, replacement: `${root}src/$1` },
    ],
  },
  test: {
    include: ['test/**/*.test.ts'],
    exclude: ['test/types/**', 'node_modules', 'dist'],
    environment: 'node',
    globals: false,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov', 'html'],
      include: ['src/**/*.ts'],
      exclude: [
        'src/generated/**',
        'src/**/*.d.ts',
        'src/index.ts',
        'src/version.ts',
        'src/types/**',
        'src/resources/index.ts',
        'src/resources/events/index.ts',
        'src/resources/events/types.ts',
        'src/errors/index.ts',
      ],
      thresholds: {
        lines: 100,
        branches: 100,
        functions: 100,
        statements: 100,
      },
    },
  },
})
