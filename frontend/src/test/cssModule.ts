/**
 * Resolves a CSS Module class name for test assertions.
 *
 * `tsconfig.json` enables `noUncheckedIndexedAccess`, which makes any access
 * into a CSS Module's index-signature type (e.g. `styles.primary`) resolve to
 * `string | undefined`. Component code tolerates that via
 * `.filter(Boolean).join(' ')`, but test assertions like `toHaveClass(...)`
 * require a definite `string`. This helper narrows the type and fails fast
 * with a clear message if the expected class is missing, rather than letting
 * `undefined` silently reach an assertion.
 */
export function cssClass(
  classes: Record<string, string | undefined>,
  key: string
): string {
  const value = classes[key]
  if (!value) {
    throw new Error(`Expected CSS Module class "${key}" to be defined`)
  }
  return value
}
