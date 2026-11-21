import { fireEvent, render, screen } from '@testing-library/react'

import { useState } from 'react'
import { expect, it, vi } from 'vitest'
import { AssimilacaoSheet } from './AssimilacaoSheet'
import { newItem, newSheet, type RollRequest, type Sheet } from './types'

const click = (el: HTMLElement) => fireEvent.click(el)

function Harness({ initial = newSheet(), onRoll, readOnly }: { initial?: Sheet; onRoll?: (r: RollRequest) => void; readOnly?: boolean }) {
  const [sheet, setSheet] = useState(initial)
  return <AssimilacaoSheet sheet={sheet} onChange={setSheet} onRollRequest={onRoll} readOnly={readOnly} />
}

it('marking damage changes level and penalty', () => {
  const s = newSheet()
  s.instincts.potency = 2
  render(<Harness initial={s} />)
  expect(screen.getByTestId('health-summary')).toHaveTextContent('Saudável')
  click(screen.getByLabelText('Caixa de dano 9'))
  expect(screen.getByTestId('health-summary')).toHaveTextContent('Laceração · −1 Sucesso')
  click(screen.getByLabelText('Caixa de dano 9'))
  click(screen.getByLabelText('Caixa de dano 7'))
  expect(screen.getByTestId('health-summary')).toHaveTextContent('Escoriação · sem penalidade')
})

it('stepper respects bounds and warns when creation total is exceeded', () => {
  render(<Harness />)
  const plus = (n: string) => screen.getByLabelText(`Aumentar ${n}`)
  for (const n of ['Potência', 'Resolução', 'Reação', 'Influência']) {
    click(plus(n))
    click(plus(n))
  }
  expect(plus('Potência')).toBeDisabled()
  expect(screen.getByLabelText('Diminuir Percepção')).toBeDisabled()
  expect(screen.getByText(/Instintos somam mais que 9/)).toBeInTheDocument()
})

it('moves items between body and backpack within limits', () => {
  const s = newSheet()
  s.inventory.body = [{ ...newItem(), name: 'Faca' }]
  s.inventory.backpack = [newItem(), newItem(), newItem(), newItem(), newItem(), newItem()]
  render(<Harness initial={s} />)
  expect(screen.getByRole('button', { name: 'Mover para mochila' })).toBeDisabled()
  click(screen.getAllByRole('button', { name: 'Mover para corpo' })[0])
  expect(screen.getByText('Corpo (2/3)')).toBeInTheDocument()
  expect(screen.getByText('Mochila (5/6)')).toBeInTheDocument()
  click(screen.getByRole('button', { name: 'Adicionar item ao corpo' }))
  expect(screen.getByRole('button', { name: 'Adicionar item ao corpo' })).toBeDisabled()
})

it('selecting instinct then aptitude requests a roll', () => {
  const onRoll = vi.fn()
  render(<Harness onRoll={onRoll} />)
  click(screen.getByRole('button', { name: 'Potência' }))
  click(screen.getByRole('button', { name: 'Armas' }))
  click(screen.getByRole('button', { name: 'Rolar' }))
  expect(onRoll).toHaveBeenCalledWith({ instinct: 'potency', aptitude: 'weapons' })
})

it('readOnly disables editing and rolling', () => {
  render(<Harness readOnly onRoll={vi.fn()} />)
  expect(screen.getByLabelText('Aumentar Potência')).toBeDisabled()
  expect(screen.getByLabelText('Caixa de dano 1')).toBeDisabled()
  expect(screen.queryByRole('button', { name: 'Rolar' })).toBeNull()
})
