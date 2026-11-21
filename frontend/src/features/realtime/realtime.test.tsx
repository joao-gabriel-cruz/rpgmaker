import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { FakeWebSocket } from '../../test/fakeWebSocket'
import { CampaignSocket } from './CampaignSocket'
import { resetSockets, socketFactory, useCampaignSocket } from './useCampaignSocket'

const create = (url: string) => new FakeWebSocket(url) as unknown as WebSocket

beforeEach(() => {
  vi.useFakeTimers()
  FakeWebSocket.instances = []
  socketFactory.create = create
})

afterEach(() => {
  resetSockets()
  socketFactory.create = undefined
  vi.useRealTimers()
})

function openSocket() {
  const s = new CampaignSocket({ url: 'ws://x/ws', createSocket: create })
  s.open()
  FakeWebSocket.last.open()
  return s
}

it('resolves send on ack and rejects with the server message on error', async () => {
  const s = openSocket()
  const ok = s.send('token.move', { x: 1 })
  const sent = JSON.parse(FakeWebSocket.last.sent[0])
  expect(sent).toMatchObject({ type: 'token.move', payload: { x: 1 } })
  FakeWebSocket.last.receive({ type: 'ack', payload: { reqId: sent.reqId } })
  await expect(ok).resolves.toBeUndefined()

  const bad = s.send('secret.roll')
  const id = JSON.parse(FakeWebSocket.last.sent[1]).reqId
  FakeWebSocket.last.receive({ type: 'error', payload: { reqId: id, message: 'Sem permissão' } })
  await expect(bad).rejects.toThrow('Sem permissão')
})

it('rejects send while disconnected and pending requests on drop', async () => {
  const s = openSocket()
  const pending = s.send('x')
  FakeWebSocket.last.drop()
  await expect(pending).rejects.toThrow('Conexão perdida')
  await expect(s.send('x')).rejects.toThrow('Sem conexão')
})

it('reconnects with backoff after a drop', () => {
  const s = openSocket()
  FakeWebSocket.last.drop()
  expect(s.status).toBe('reconnecting')
  expect(FakeWebSocket.instances).toHaveLength(1)
  vi.advanceTimersByTime(600)
  expect(FakeWebSocket.instances).toHaveLength(2)
  FakeWebSocket.last.drop()
  vi.advanceTimersByTime(600)
  expect(FakeWebSocket.instances).toHaveLength(2) // second delay is ~1s
  vi.advanceTimersByTime(600)
  expect(FakeWebSocket.instances).toHaveLength(3)
  FakeWebSocket.last.open()
  expect(s.status).toBe('open')
})

it('delivers events to subscribers, replays presence, and unsubscribes', () => {
  const s = openSocket()
  FakeWebSocket.last.receive({ type: 'presence.updated', payload: { users: [] } })
  const seen: unknown[] = []
  const off = s.subscribe('presence.updated', (p) => seen.push(p))
  expect(seen).toHaveLength(1)
  FakeWebSocket.last.receive({ type: 'presence.updated', payload: { users: [1] } })
  expect(seen).toHaveLength(2)
  off()
  FakeWebSocket.last.receive({ type: 'presence.updated', payload: { users: [2] } })
  expect(seen).toHaveLength(2)
})

it('shares one socket between hook callers and closes it on last unmount', () => {
  const a = renderHook(() => useCampaignSocket('c1'))
  const b = renderHook(() => useCampaignSocket('c1'))
  expect(FakeWebSocket.instances).toHaveLength(1)
  expect(FakeWebSocket.last.url).toContain('/api/campaigns/c1/ws')
  expect(a.result.current.status).toBe('connecting')
  act(() => FakeWebSocket.last.open())
  expect(b.result.current.status).toBe('open')
  a.unmount()
  expect(b.result.current.status).toBe('open')
  b.unmount()
  expect(FakeWebSocket.instances).toHaveLength(1)
})
