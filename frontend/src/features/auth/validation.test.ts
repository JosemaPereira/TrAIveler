import { describe, expect, it } from 'vitest'

import {
  isValidEmail,
  isValidFullName,
  validatePasswordStrength,
} from './validation'

describe('isValidEmail', () => {
  describe('when the value looks like an email', () => {
    it('should return true', () => {
      expect(isValidEmail('traveler@example.com')).toBe(true)
    })
  })

  describe('when the value has no @ sign', () => {
    it('should return false', () => {
      expect(isValidEmail('traveler.example.com')).toBe(false)
    })
  })

  describe('when the value has no domain', () => {
    it('should return false', () => {
      expect(isValidEmail('traveler@')).toBe(false)
    })
  })

  describe('when the value is empty', () => {
    it('should return false', () => {
      expect(isValidEmail('')).toBe(false)
    })
  })

  describe('when the value has surrounding whitespace but is otherwise valid', () => {
    it('should return true', () => {
      expect(isValidEmail('  traveler@example.com  ')).toBe(true)
    })
  })
})

describe('validatePasswordStrength', () => {
  // Mirrors backend/internal/auth/validator.go: 8-72 characters, at least one
  // uppercase letter, one lowercase letter, and one digit.
  describe('when the password satisfies every rule', () => {
    it('should return no errors', () => {
      expect(validatePasswordStrength('CorrectHorse1!')).toEqual([])
    })
  })

  describe('when the password is shorter than 8 characters', () => {
    it('should report the length rule', () => {
      expect(validatePasswordStrength('Ab1')).toContain(
        'Password must be between 8 and 72 characters'
      )
    })
  })

  describe('when the password is longer than 72 characters', () => {
    it('should report the length rule', () => {
      const tooLong = `Aa1${'a'.repeat(70)}`
      expect(validatePasswordStrength(tooLong)).toContain(
        'Password must be between 8 and 72 characters'
      )
    })
  })

  describe('when the password has no uppercase letter', () => {
    it('should report the uppercase rule', () => {
      expect(validatePasswordStrength('lowercase1')).toContain(
        'Password must contain at least one uppercase letter'
      )
    })
  })

  describe('when the password has no lowercase letter', () => {
    it('should report the lowercase rule', () => {
      expect(validatePasswordStrength('UPPERCASE1')).toContain(
        'Password must contain at least one lowercase letter'
      )
    })
  })

  describe('when the password has no digit', () => {
    it('should report the digit rule', () => {
      expect(validatePasswordStrength('NoDigitsHere')).toContain(
        'Password must contain at least one digit'
      )
    })
  })

  describe('when the password fails multiple rules', () => {
    it('should report every failing rule', () => {
      expect(validatePasswordStrength('short')).toEqual([
        'Password must be between 8 and 72 characters',
        'Password must contain at least one uppercase letter',
        'Password must contain at least one digit',
      ])
    })
  })
})

describe('isValidFullName', () => {
  describe('when the value has visible characters', () => {
    it('should return true', () => {
      expect(isValidFullName('Ada Traveler')).toBe(true)
    })
  })

  describe('when the value is empty', () => {
    it('should return false', () => {
      expect(isValidFullName('')).toBe(false)
    })
  })

  describe('when the value is only whitespace', () => {
    it('should return false', () => {
      expect(isValidFullName('   ')).toBe(false)
    })
  })
})
