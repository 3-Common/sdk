// @ts-check
import js from '@eslint/js'
import { defineConfig } from 'eslint/config'
import importX from 'eslint-plugin-import-x'
import tsdoc from 'eslint-plugin-tsdoc'
import unicorn from 'eslint-plugin-unicorn'
import tseslint from 'typescript-eslint'

export default defineConfig(
  {
    ignores: ['dist', 'coverage', 'node_modules', 'src/generated'],
  },
  js.configs.recommended,
  tseslint.configs.strictTypeChecked,
  tseslint.configs.stylisticTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.eslint.json'],
        tsconfigRootDir: import.meta.dirname,
      },
    },
    plugins: {
      unicorn,
      'import-x': importX,
      tsdoc,
    },
    settings: {
      'import-x/resolver': {
        typescript: { project: './tsconfig.eslint.json' },
        node: true,
      },
    },
    rules: {
      // Public-surface non-negotiable.
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-non-null-assertion': 'error',
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/switch-exhaustiveness-check': 'error',
      '@typescript-eslint/consistent-type-imports': [
        'error',
        { prefer: 'type-imports', fixStyle: 'separate-type-imports' },
      ],
      '@typescript-eslint/no-import-type-side-effects': 'error',
      '@typescript-eslint/explicit-module-boundary-types': 'error',

      // Imports.
      'import-x/no-cycle': 'error',
      'import-x/order': [
        'error',
        {
          'newlines-between': 'always',
          groups: ['builtin', 'external', 'internal', 'parent', 'sibling', 'index', 'type'],
          alphabetize: { order: 'asc', caseInsensitive: true },
        },
      ],
      'import-x/no-default-export': 'error',
      'import-x/no-duplicates': 'error',

      // Modern Node.
      'unicorn/prefer-node-protocol': 'error',
      'unicorn/no-array-for-each': 'off',
      'unicorn/no-null': 'off',
      'unicorn/prevent-abbreviations': 'off',
      'unicorn/filename-case': ['error', { cases: { pascalCase: true, kebabCase: true } }],

      // Documentation.
      'tsdoc/syntax': 'warn',
    },
  },
  {
    // Test files: relax a few checks.
    files: ['test/**/*.ts', '**/*.test.ts', '**/*.test-d.ts'],
    rules: {
      '@typescript-eslint/no-non-null-assertion': 'off',
      '@typescript-eslint/no-unsafe-assignment': 'off',
      '@typescript-eslint/no-unsafe-member-access': 'off',
      'import-x/no-default-export': 'off',
    },
  },
  {
    // Config files at root.
    files: ['*.config.ts', '*.config.js', 'eslint.config.js', 'tsup.config.ts'],
    rules: {
      'import-x/no-default-export': 'off',
    },
  },
)
