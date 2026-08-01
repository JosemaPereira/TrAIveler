import { describe, expect, it } from 'vitest'

import { resolveLoginRedirect } from './login-redirect'

describe('resolveLoginRedirect', () => {
  describe('when there is no redirect param', () => {
    it('should default to the dashboard', () => {
      expect(resolveLoginRedirect(null)).toBe('/dashboard')
    })
  })

  describe('when the redirect param is empty', () => {
    it('should default to the dashboard', () => {
      expect(resolveLoginRedirect('')).toBe('/dashboard')
    })
  })

  describe('when the redirect param is a same-origin relative path', () => {
    it('should return it unchanged', () => {
      expect(resolveLoginRedirect('/trips/trip_1')).toBe('/trips/trip_1')
    })

    it('should preserve a query string on the target path', () => {
      expect(resolveLoginRedirect('/settings?tab=billing')).toBe(
        '/settings?tab=billing'
      )
    })
  })

  describe('when the redirect param is protocol-relative (//)', () => {
    it('should fall back to the dashboard rather than allow an off-site host', () => {
      expect(resolveLoginRedirect('//evil.example.com')).toBe('/dashboard')
    })
  })

  describe('when the redirect param is an absolute URL', () => {
    it('should fall back to the dashboard', () => {
      expect(resolveLoginRedirect('https://evil.example.com')).toBe(
        '/dashboard'
      )
    })

    it('should fall back to the dashboard for a scheme-relative host without protocol', () => {
      expect(resolveLoginRedirect('http://evil.example.com/trips')).toBe(
        '/dashboard'
      )
    })
  })

  describe('when the redirect param uses a backslash to smuggle a host', () => {
    it('should fall back to the dashboard', () => {
      expect(resolveLoginRedirect('/\\evil.example.com')).toBe('/dashboard')
    })
  })

  describe('when the redirect param does not start with a slash', () => {
    it('should fall back to the dashboard', () => {
      expect(resolveLoginRedirect('dashboard')).toBe('/dashboard')
    })
  })

  describe('when the redirect param is exactly the root path', () => {
    it('should return the root path unchanged', () => {
      expect(resolveLoginRedirect('/')).toBe('/')
    })
  })
})
