import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './styles/global.css'
import App from './App.tsx'
import { installSessionExpiryHandler } from './features/auth/session-expiry'

// Composition root: teach the API client what to do when the session expires
// beyond recovery. Done here, before the first render, so the very first
// request of the app session is already covered — and kept out of
// `lib/api-client.ts` itself, which must not import the auth store.
installSessionExpiryHandler()

const rootElement = document.getElementById('root')
if (!rootElement) {
  throw new Error('Failed to find the root element')
}

createRoot(rootElement).render(
  <StrictMode>
    <App />
  </StrictMode>
)
