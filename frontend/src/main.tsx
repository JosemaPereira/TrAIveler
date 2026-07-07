/**
 * Application entry point for TrAIveler frontend.
 * 
 * Initializes the React application by:
 * 1. Locating the root DOM element
 * 2. Creating a React root with StrictMode enabled for development checks
 * 3. Rendering the main App component
 * 
 * StrictMode activates additional checks and warnings for its descendants,
 * helping identify potential problems in the application during development.
 */
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

// Ensure the root element exists before attempting to mount React
const rootElement = document.getElementById('root')
if (!rootElement) {
  throw new Error('Failed to find the root element')
}

createRoot(rootElement).render(
  <StrictMode>
    <App />
  </StrictMode>
)
