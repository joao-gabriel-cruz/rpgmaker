import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'
import { stubApi } from '../../test/renderApp'
import { fakeRealtime } from '../../test/fakeRealtime'
import { useEffect } from 'react'
import { TableCharacterProvider, useTableCharacter } from '../hotbar/TableCharacterContext'
import { AmbienceOptions } from './AmbienceOptions'
import type { SceneAmbience } from './useSceneAmbience'

afterEach(() => vi.unstubAllGlobals())

const scene = (over: Partial<SceneAmbience['ambience']> = {}): SceneAmbience => ({
  ambience: { sceneId: 's1', timeOfDay: 'day', weather: 'clear', ...over }, set: vi.fn(), error: '',
})

function Select({ id }: { id?: string }) {
  const { select } = useTableCharacter()
  useEffect(() => select(id ?? null), [id, select])
  return null
}

function mount(sc: SceneAmbience, sheet: Record<string, unknown> = {}, selected?: string) {
  stubApi((_m, path) => {
    if (path === '/auth/me') return { status: 200, body: { id: 'gm', email: '', displayName: 'GM', isAdmin: false, createdAt: '' } }
    return { status: 200, body: [{ id: 'c1', name: 'Rita', kind: 'pc', ownerId: 'u1', campaignId: 'c1', ownerName: '', sheet, updatedAt: '', issues: [] }] }
  })
  const rt = fakeRealtime()
  render(
    <QueryClientProvider client={new QueryClient()}>
      <rt.Provider><TableCharacterProvider><Select id={selected} /><AmbienceOptions campaignId="c1" scene={sc} /></TableCharacterProvider></rt.Provider>
    </QueryClientProvider>,
  )
  return rt
}

it('the Mestre changes the time of day and the weather', async () => {
  const sc = scene({ timeOfDay: 'night' })
  mount(sc)
  expect(screen.getByRole('button', { name: 'Noite' })).toHaveAttribute('aria-pressed', 'true')
  await userEvent.click(screen.getByRole('button', { name: 'Entardecer' }))
  expect(sc.set).toHaveBeenCalledWith({ timeOfDay: 'dusk' })
  await userEvent.click(screen.getByRole('button', { name: 'Chuva' }))
  expect(sc.set).toHaveBeenCalledWith({ weather: 'rain' })
})

it('toggles manual conditions on the selected character and locks automatic ones', async () => {
  const rt = mount(scene(), { conditions: ['scared'], health: { damage: 0 }, tugOfWar: { determinationPoints: 0 }, instincts: { potency: 1, resolution: 1 } }, 'c1')
  const scared = await screen.findByRole('button', { name: /Assustado/ })
  expect(scared).toHaveAttribute('aria-pressed', 'true')
  expect(screen.getByRole('button', { name: /Suscetível/ })).toBeDisabled()
  await userEvent.click(screen.getByRole('button', { name: /Ferido/ }))
  expect(rt.send).toHaveBeenCalledWith('condition.set', { characterId: 'c1', condition: 'wounded', active: true })
  await userEvent.click(scared)
  expect(rt.send).toHaveBeenCalledWith('condition.set', { characterId: 'c1', condition: 'scared', active: false })
})

it('asks for a character when none is selected', async () => {
  mount(scene())
  expect(await screen.findByText(/Selecione um personagem/)).toBeInTheDocument()
})
