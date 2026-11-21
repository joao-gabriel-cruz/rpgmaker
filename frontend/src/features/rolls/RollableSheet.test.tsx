import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'
import { RealtimeContext } from '../realtime/realtimeContext'
import type { CampaignSocketApi } from '../realtime/useCampaignSocket'
import { RollableSheet } from './RollableSheet'

vi.mock('../sheet-assimilacao/AssimilacaoSheet', () => ({
  AssimilacaoSheet: (p: { onRollRequest: (r: unknown) => void }) => (
    <button onClick={() => p.onRollRequest({ instinct: 'reaction', aptitude: 'athletics' })}>rolar</button>
  ),
}))

function setup(isGm: boolean) {
  const send = vi.fn().mockResolvedValue(undefined)
  const api = { status: 'open', send, subscribe: () => () => {} } as unknown as CampaignSocketApi
  render(
    <RealtimeContext.Provider value={api}>
      <RollableSheet characterId="c1" isGm={isGm} sheet={{} as never} onChange={() => {}} />
    </RealtimeContext.Provider>,
  )
  return send
}

it('sends roll.request for the character', async () => {
  const send = setup(false)
  expect(screen.queryByLabelText('Rolagem secreta')).toBeNull()
  await userEvent.click(screen.getByText('rolar'))
  expect(send).toHaveBeenCalledWith('roll.request', { characterId: 'c1', instinct: 'reaction', aptitude: 'athletics', secret: false })
})

it('lets the Mestre roll in secret', async () => {
  const send = setup(true)
  await userEvent.click(screen.getByLabelText('Rolagem secreta'))
  await userEvent.click(screen.getByText('rolar'))
  expect(send).toHaveBeenCalledWith('roll.request', expect.objectContaining({ secret: true }))
})
