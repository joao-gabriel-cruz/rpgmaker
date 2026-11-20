import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { HealthStatus } from './HealthStatus'

function mockFetch(impl: () => Promise<Response>) {
  vi.stubGlobal('fetch', vi.fn(impl))
}

const json = (status: number, body: unknown) =>
  Promise.resolve(new Response(JSON.stringify(body), { status }))

describe('HealthStatus', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('shows API and database ok', async () => {
    mockFetch(() => json(200, { api: 'ok', database: 'ok' }))
    render(<HealthStatus />)
    expect(await screen.findByRole('status')).toHaveTextContent('API ok · banco ok')
  })

  it('shows database down when the API responds 503', async () => {
    mockFetch(() => json(503, { api: 'ok', database: 'error' }))
    render(<HealthStatus />)
    expect(await screen.findByRole('status')).toHaveTextContent('API ok · banco fora do ar')
  })

  it('shows API down when the request fails', async () => {
    mockFetch(() => Promise.reject(new TypeError('Failed to fetch')))
    render(<HealthStatus />)
    expect(await screen.findByRole('status')).toHaveTextContent('API fora do ar')
  })
})
