/// <reference types="vite/client" />

// Required, not optional: api-client.ts's getBaseUrl() throws at runtime if this
// is unset, so the type should reflect that callers can rely on a plain string.
interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
