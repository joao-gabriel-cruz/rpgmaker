import type { AptitudeKey, Generation, InstinctKey, Suit } from './types'

export const INSTINCT_LABELS: Record<InstinctKey, string> = {
  influence: 'Influência',
  perception: 'Percepção',
  potency: 'Potência',
  reaction: 'Reação',
  resolution: 'Resolução',
  sagacity: 'Sagacidade',
}

export const KNOWLEDGE_LABELS = {
  biology: 'Biologia',
  erudition: 'Erudição',
  engineering: 'Engenharia',
  geography: 'Geografia',
  medicine: 'Medicina',
  security: 'Segurança',
} as const

export const PRACTICE_LABELS = {
  weapons: 'Armas',
  athletics: 'Atletismo',
  expression: 'Expressão',
  stealth: 'Furtividade',
  crafting: 'Manufaturas',
  survival: 'Sobrevivência',
} as const

export const APTITUDE_LABELS: Record<AptitudeKey, string> = { ...KNOWLEDGE_LABELS, ...PRACTICE_LABELS }

export const GENERATION_LABELS: Record<Generation, string> = {
  pre_collapse: 'Pré-Colapso',
  during_collapse: 'Durante o Colapso',
  post_collapse: 'Pós-Colapso',
}

export const SUIT_LABELS: Record<Suit, string> = {
  hearts: '♥ Evolutiva',
  diamonds: '♦ Adaptativa',
  spades: '♠ Inoportuna',
  clubs: '♣ Singular',
}

export const HEALTH_LEVEL_NAMES = ['Saudável', 'Escoriação', 'Laceração', 'Ferimentos', 'Debilitação', 'Incapacitação']

export const INSTINCT_ADJECTIVES = ['', 'Comum', 'Notável', 'Excepcional']
export const APTITUDE_ADJECTIVES = ['Leigo', 'Treinado', 'Especialista']
