export type Phase = 'preparation' | 'infected' | 'assimilator' | 'end'

export interface RoundState {
  active: boolean
  round: number
  phase: Phase
}

export const PHASES: { id: Phase; label: string }[] = [
  { id: 'preparation', label: 'Preparação' },
  { id: 'infected', label: 'Infectados' },
  { id: 'assimilator', label: 'Assimilador' },
  { id: 'end', label: 'Fim' },
]

export const IDLE: RoundState = { active: false, round: 0, phase: 'preparation' }
