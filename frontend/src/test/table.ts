import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

/** Clicks a rail tool by its accessible name (e.g. "Terreno"). */
export const pickTool = (name: string) => userEvent.click(screen.getByRole('button', { name }))

/** Opens the Mestre's campaign dialog, optionally on one of its tabs. */
export async function openManage(tab?: string) {
  await userEvent.click(await screen.findByRole('button', { name: 'Gerenciar campanha' }))
  if (tab) await userEvent.click(screen.getByRole('tab', { name: tab }))
}

/** Opens a sidebar tab by name (the badge suffix, if any, is ignored). */
export const openTab = (name: string) => userEvent.click(screen.getByRole('tab', { name: new RegExp(`^${name}`) }))
