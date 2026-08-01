import { useState } from 'react'

import { Button } from '@/components/primitives/Button'
import { useLogout } from '@/features/auth/hooks/useLogout'
import { useUser } from '@/stores/auth-store'
import styles from './Navigation.module.css'

const MENU_ID = 'primary-navigation-menu'

/**
 * Persistent authenticated-shell navigation.
 *
 * Log Out delegates to `useLogout` — store teardown, cache clear, and the
 * redirect all live there; this component only triggers the mutation.
 */
export function Navigation() {
  const user = useUser()
  const logout = useLogout()
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  const menuClassName = [styles.menu, isMenuOpen ? styles.menuOpen : '']
    .filter(Boolean)
    .join(' ')

  return (
    <nav aria-label="Main navigation" className={styles.nav}>
      <div className={styles.bar}>
        <span className={styles.brand}>TrAIveler</span>
        <button
          type="button"
          className={styles.menuToggle}
          aria-expanded={isMenuOpen}
          aria-controls={MENU_ID}
          onClick={() => {
            setIsMenuOpen((open) => !open)
          }}
        >
          <span className={styles.srOnly}>Menu</span>
          <span aria-hidden="true">☰</span>
        </button>
      </div>
      <div id={MENU_ID} className={menuClassName}>
        {user && (
          <div className={styles.userInfo}>
            <span className={styles.userName}>{user.full_name}</span>
            {user.has_subscription && (
              <span className={styles.badge}>Subscriber</span>
            )}
          </div>
        )}
        <Button
          variant="secondary"
          onClick={() => {
            logout.mutate()
          }}
        >
          Log Out
        </Button>
      </div>
    </nav>
  )
}
