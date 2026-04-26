import { describe, it, expect } from 'vitest'

describe('Test Framework Setup', () => {
  it('should run a simple test', () => {
    expect(1 + 1).toBe(2)
  })

  it('should work with jest-dom matchers', () => {
    const div = document.createElement('div')
    document.body.appendChild(div)
    expect(div).toBeInTheDocument()
  })
})