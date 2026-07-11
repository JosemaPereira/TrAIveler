/**
 * ESLint configuration for TrAIveler frontend with strict TypeScript rules.
 * 
 * Key features:
 * - Strict TypeScript type checking (no-explicit-any enforced as error)
 * - React Hooks rules to prevent common mistakes
 * - React Fast Refresh compatibility checks
 * - Unused variables allowed with underscore prefix (e.g., _unusedParam)
 * 
 * This configuration enforces:
 * - Zero usage of 'any' type (promotes type safety)
 * - Proper React Hooks dependency arrays
 * - Components suitable for Fast Refresh (avoid inline exports)
 * 
 * @see https://typescript-eslint.io/
 */
import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  // Ignore build output and coverage report directories
  { ignores: ['dist', 'coverage'] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.strictTypeChecked],
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.json'],
        tsconfigRootDir: import.meta.dirname,
      },
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      // Enable all recommended React Hooks rules (exhaustive-deps, rules-of-hooks)
      ...reactHooks.configs.recommended.rules,
      
      // Warn when files export non-components alongside components (breaks Fast Refresh)
      // allowConstantExport: true permits exporting constants like API_URL
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      
      // MANDATORY: Ban 'any' type to enforce type safety
      // Use 'unknown' with type guards instead
      '@typescript-eslint/no-explicit-any': 'error',
      
      // Allow unused variables/parameters that start with underscore
      // Useful for: const { id, ...rest } = props OR function onClick(_event) {}
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_' },
      ],
      
      // Disabled: Return type inference is sufficient for most cases
      // Explicit return types add noise without significant benefit
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',
    },
  },
)
