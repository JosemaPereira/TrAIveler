import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from './App'

describe('App', () => {
  it('mounts without throwing', () => {
    expect(() => render(<App />)).not.toThrow()
  })

  it('renders the TrAIveler heading reached through the full provider stack', async () => {
    render(<App />)

    expect(
      await screen.findByRole('heading', { name: 'TrAIveler' })
    ).toBeInTheDocument()
  })

  it('no longer renders the static placeholder tagline, replaced by routed content', async () => {
    render(<App />)

    await screen.findByRole('heading', { name: 'TrAIveler' })

    expect(
      screen.queryByText('AI-Powered Travel Itinerary Planner')
    ).not.toBeInTheDocument()
  })
})
