import { Component } from 'react'
import type { ErrorInfo, ReactNode } from 'react'

import { Button } from './primitives/Button'
import styles from './ErrorBoundary.module.css'

interface ErrorBoundaryProps {
  children: ReactNode
}

interface ErrorBoundaryState {
  error: Error | null
}

const GENERIC_ERROR_MESSAGE = "We're sorry, but something unexpected happened."

/**
 * Performs a full page navigation to the home route. A full navigation (not
 * a router-aware push) is used deliberately: this fallback renders outside
 * the Router's context (ErrorBoundary sits above RouterProvider in App.tsx),
 * and app/component state may be corrupted after an uncaught render error,
 * so a hard reload is the safer recovery path.
 */
function goHome(): void {
  window.location.assign('/')
}

/**
 * Top-level error boundary. Catches rendering errors anywhere in its
 * subtree and replaces the crashed tree with a full-page fallback instead
 * of leaving a blank screen. Error boundaries currently require a class
 * component — React has no hook-based equivalent of
 * getDerivedStateFromError/componentDidCatch.
 */
export class ErrorBoundary extends Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  state: ErrorBoundaryState = { error: null }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  render(): ReactNode {
    const { error } = this.state
    if (!error) {
      return this.props.children
    }

    return (
      <div className={styles.container}>
        <h1 className={styles.title}>Something went wrong</h1>
        <p className={styles.message}>{GENERIC_ERROR_MESSAGE}</p>
        {import.meta.env.DEV && (
          <pre className={styles.details}>{error.message}</pre>
        )}
        <Button onClick={goHome}>Go Home</Button>
      </div>
    )
  }
}
