import { useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'

export type CrisisKind = 'overpopulation' | 'breach' | 'shortage' | 'custom' | 'internal_fight' | 'disease' | 'creature' | 'rebellion' | 'invasion'
export type CrisisStatus = 'suggested' | 'active' | 'contained' | 'resolved'

export interface SeverityRoll { rollId: string; faces: number[]; discarded: number | null; symbols: number | null }
export interface CrisisSuggestion { id: string; kind: string; text: string; status: 'pending' | 'applied' | 'dismissed' }

export interface Crisis {
  id: string
  name: string
  kind: CrisisKind
  grade: number
  status: CrisisStatus
  description: string
  week: number
  createdBy: string
  trigger?: string
  roll?: SeverityRoll
  suggestions: CrisisSuggestion[]
}

export type ResidentRole = 'guard' | 'collector' | 'medic' | 'artisan' | 'leader' | 'other'
export interface Resident { id: string; characterId: string; name: string; role: ResidentRole; note: string }

export type ExpeditionStatus = 'proposed' | 'ready' | 'rolling' | 'done'
export interface Expedition {
  id: string
  status: ExpeditionStatus
  characterIds: string[]
  leaderName: string
  resource: string
  regionNote: string
  proposedBy: string
  week: number
  rollId?: string
  factor: number
  successes: number
  adaptations: number
  pressures: number
  found: number
  discarded: number
  extras: number
  complications: number
  log: string[]
}

export const KIND_LABEL: Record<CrisisKind, string> = {
  overpopulation: 'Superlotação', breach: 'Defesa insuficiente', shortage: 'Falta de recursos', custom: 'Personalizada',
  internal_fight: 'Briga interna', disease: 'Doença', creature: 'Criatura assimilada', rebellion: 'Rebelião', invasion: 'Invasores',
}
export const KINDS = Object.keys(KIND_LABEL) as CrisisKind[]
export const STATUS_LABEL: Record<CrisisStatus, string> = { suggested: 'Sugerida', active: 'Ativa', contained: 'Contida', resolved: 'Resolvida' }
export const STATUS_TONE: Record<CrisisStatus, string> = {
  suggested: 'border-amber-500 text-amber-300', active: 'border-red-500 text-red-300',
  contained: 'border-yellow-500 text-yellow-300', resolved: 'border-emerald-600 text-emerald-300',
}
export const ROLE_LABEL: Record<ResidentRole, string> = {
  guard: 'Guarda', collector: 'Coletor', medic: 'Médico', artisan: 'Artesão', leader: 'Líder', other: 'Outro',
}
export const ROLES = Object.keys(ROLE_LABEL) as ResidentRole[]
export const EXPEDITION_STATUS_LABEL: Record<ExpeditionStatus, string> = {
  proposed: 'Aguardando aprovação', ready: 'Pronta para rolar', rolling: 'Rolando', done: 'Concluída',
}

/** Sends a realtime command and keeps the last error message for display. */
export function useRefugeCommand() {
  const { send } = useRealtime()
  const [error, setError] = useState('')
  const run = (type: string, payload: object = {}) => {
    setError('')
    send(type, payload).catch((e: Error) => setError(e.message))
  }
  return { run, error }
}
