import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from './App'

describe('<App />', () => {
  describe('when rendered through the full provider stack', () => {
    it('should mount without throwing', () => {
      expect(() => render(<App />)).not.toThrow()
    })

    it('should render the TrAIveler heading', async () => {
      render(<App />)

      expect(
        await screen.findByRole('heading', { name: 'TrAIveler' })
      ).toBeInTheDocument()
    })

    it('should not render the static placeholder tagline replaced by routed content', async () => {
      render(<App />)

      await screen.findByRole('heading', { name: 'TrAIveler' })

      expect(
        screen.queryByText('AI-Powered Travel Itinerary Planner')
      ).not.toBeInTheDocument()
    })
  })
})
