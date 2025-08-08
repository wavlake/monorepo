import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import App from './App'

// Mock console.log for testing
const mockConsoleLog = vi.spyOn(console, 'log').mockImplementation(() => {})

describe('App', () => {
  it('renders Wavlake title', () => {
    render(<App />)
    expect(screen.getByText('🎵 Wavlake')).toBeInTheDocument()
  })

  it('renders the subtitle', () => {
    render(<App />)
    expect(screen.getByText('Decentralized Music Platform')).toBeInTheDocument()
  })

  it('displays initial count of 0', () => {
    render(<App />)
    expect(screen.getByText('Count is 0')).toBeInTheDocument()
  })

  it('increments count when button is clicked', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    const countButton = screen.getByRole('button', { name: /count is/i })
    await user.click(countButton)
    
    expect(screen.getByText('Count is 1')).toBeInTheDocument()
  })

  it('logs Nostr event kinds when Test Shared Types button is clicked', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    const testButton = screen.getByRole('button', { name: /test shared types/i })
    await user.click(testButton)
    
    expect(mockConsoleLog).toHaveBeenCalledWith('Nostr Text Note Kind:', 1)
    expect(mockConsoleLog).toHaveBeenCalledWith('Nostr Track Metadata Kind:', 31337)
  })

  it('displays feature list', () => {
    render(<App />)
    expect(screen.getByText('🚀 Features')).toBeInTheDocument()
    expect(screen.getByText('React 18 + TypeScript')).toBeInTheDocument()
    expect(screen.getByText('Tailwind CSS for styling')).toBeInTheDocument()
    expect(screen.getByText('Shared types from Go backend')).toBeInTheDocument()
  })

  it('displays development commands', () => {
    render(<App />)
    expect(screen.getByText('🧪 Development Commands')).toBeInTheDocument()
    expect(screen.getByText(/task tdd:frontend/)).toBeInTheDocument()
  })

  it('displays tech stack section', () => {
    render(<App />)
    expect(screen.getByText('🛠️ Tech Stack')).toBeInTheDocument()
    expect(screen.getByText('React 18')).toBeInTheDocument()
    expect(screen.getByText('Tailwind CSS')).toBeInTheDocument()
    expect(screen.getByText('Go Backend')).toBeInTheDocument()
    expect(screen.getByText('Nostr Protocol')).toBeInTheDocument()
  })

  it('displays footer', () => {
    render(<App />)
    expect(screen.getByText('Built with ❤️ for the music community')).toBeInTheDocument()
  })
})