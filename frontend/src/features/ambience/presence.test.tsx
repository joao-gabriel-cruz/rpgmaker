import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { fakeRealtime } from '../../test/fakeRealtime'
import { createOverlayRegistry } from './overlayRegistry'
import { PresenceProvider, usePresenceState } from './PresenceContext'
import { BUBBLE_MS, REACTION_MS } from './presence'
import { SpeakBox } from './SpeakBox'
import { SpeechLog } from './SpeechLog'
import { TokenOverlayLayer } from './TokenOverlayLayer'

beforeEach(() => vi.useFakeTimers())
afterEach(() => vi.useRealTimers())

function Probe() {
  const p = usePresenceState()
  return (
    <>
      <output aria-label="bubbles">{p?.bubbles.map((b) => `${b.characterId}:${b.text}`).join('|')}</output>
      <output aria-label="reactions">{p?.reactions.map((r) => r.emoji).join('')}</output>
      <SpeechLog />
    </>
  )
}

function setup() {
  const rt = fakeRealtime()
  render(<rt.Provider><PresenceProvider><Probe /></PresenceProvider></rt.Provider>)
  return rt
}

it('shows a speech bubble for 6 seconds, replaces it by the same speaker and logs every line', () => {
  const rt = setup()
  act(() => rt.emit('chat.said', { characterId: 'c1', name: 'Ana', text: 'Silêncio...', at: 1 }))
  expect(screen.getByLabelText('bubbles')).toHaveTextContent('c1:Silêncio...')
  act(() => { vi.advanceTimersByTime(BUBBLE_MS - 100) })
  act(() => rt.emit('chat.said', { characterId: 'c1', name: 'Ana', text: 'Ouviram isso?', at: 2 }))
  expect(screen.getByLabelText('bubbles')).toHaveTextContent('c1:Ouviram isso?')
  expect(screen.getByLabelText('bubbles')).not.toHaveTextContent('Silêncio')
  act(() => { vi.advanceTimersByTime(200) })
  expect(screen.getByLabelText('bubbles')).toHaveTextContent('Ouviram isso?') // the first timer must not drop the new bubble
  act(() => { vi.advanceTimersByTime(BUBBLE_MS) })
  expect(screen.getByLabelText('bubbles')).toBeEmptyDOMElement()
  expect(screen.getByText('“Silêncio...”')).toBeInTheDocument()
  expect(screen.getByText('“Ouviram isso?”')).toBeInTheDocument()
})

it('shows an emoji reaction briefly and then removes it', () => {
  const rt = setup()
  act(() => rt.emit('chat.reacted', { characterId: 'c1', name: 'Ana', emoji: '😱', at: 1 }))
  act(() => rt.emit('chat.reacted', { characterId: 'c2', name: 'Bob', emoji: '💀', at: 2 }))
  expect(screen.getByLabelText('reactions')).toHaveTextContent('😱💀')
  act(() => { vi.advanceTimersByTime(REACTION_MS + 10) })
  expect(screen.getByLabelText('reactions')).toBeEmptyDOMElement()
})

it('speak box sends trimmed text capped at 140 characters, closes on success and on Esc', async () => {
  vi.useRealTimers()
  const onSay = vi.fn().mockResolvedValue(undefined)
  const onClose = vi.fn()
  render(<SpeakBox name="Ana" onSay={onSay} onClose={onClose} />)
  const input = screen.getByLabelText('Falar como Ana')
  expect(input).toHaveAttribute('maxlength', '140')
  await userEvent.type(input, '  Olá, grupo  {Enter}')
  expect(onSay).toHaveBeenCalledWith('Olá, grupo')
  await vi.waitFor(() => expect(onClose).toHaveBeenCalledTimes(1))
  await userEvent.type(input, '{Escape}')
  expect(onClose).toHaveBeenCalledTimes(2)
})

it('overlay layer puts the bubble, the emoji and the condition icons on the right token only', () => {
  const rt = fakeRealtime()
  const registry = createOverlayRegistry()
  const tokens = [
    { id: 't1', sceneId: 's', characterId: 'c1', x: 1, z: 1, name: 'Ana', portraitUrl: null, ownerId: 'u1', status: { conditions: ['hidden'], health: { damage: 12 }, instincts: { potency: 1, resolution: 1 } } },
    { id: 't2', sceneId: 's', characterId: 'c2', x: 2, z: 2, name: 'Bob', portraitUrl: null, ownerId: 'u2' },
  ]
  render(<rt.Provider><PresenceProvider><TokenOverlayLayer tokens={tokens} registry={registry} /></PresenceProvider></rt.Provider>)
  expect([...registry.keys()].sort()).toEqual(['t1', 't2'])
  const icons = within(registry.get('t1') as HTMLElement).getAllByRole('listitem').map((l) => l.getAttribute('title'))
  expect(icons).toEqual(['Debilitado (automático)', 'Escondido'])
  expect(within(registry.get('t2') as HTMLElement).queryByRole('list')).toBeNull()
  act(() => rt.emit('chat.said', { characterId: 'c2', name: 'Bob', text: 'Vamos embora', at: 1 }))
  expect(within(registry.get('t2') as HTMLElement).getByRole('status')).toHaveTextContent('Vamos embora')
  expect(within(registry.get('t1') as HTMLElement).queryByRole('status')).toBeNull()
  act(() => rt.emit('chat.reacted', { characterId: 'c1', name: 'Ana', emoji: '❓', at: 2 }))
  expect(registry.get('t1')).toHaveTextContent('❓')
  act(() => { vi.advanceTimersByTime(BUBBLE_MS + 10) })
  expect(registry.get('t2')).not.toHaveTextContent('Vamos embora')
  expect(registry.get('t1')).not.toHaveTextContent('❓')
})
