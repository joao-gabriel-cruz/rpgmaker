import { xpAvailable, sumValues } from './derive'
import { APTITUDE_LABELS, INSTINCT_LABELS } from './labels'
import { BACKPACK_LIMIT, BODY_LIMIT, type Sheet, type SheetIssue } from './types'

export const INSTINCT_MAX = 3
export const INSTINCT_TOTAL = 9
export const APTITUDE_MAX = 2
export const APTITUDE_TOTAL = 7

export interface ValidateOptions {
  blockCreationLimits?: boolean
}

type Add = (code: string, field: string, message: string, creation?: boolean) => void

function checkAttributes(sheet: Sheet, add: Add) {
  for (const [key, v] of Object.entries(sheet.instincts)) {
    const label = INSTINCT_LABELS[key as keyof typeof INSTINCT_LABELS]
    if (v > INSTINCT_MAX) add('instinct_max', `instincts.${key}`, `${label} não pode passar de ${INSTINCT_MAX}.`, true)
    if (v < 1) add('instinct_min', `instincts.${key}`, `${label} não pode ser menor que 1.`, true)
  }
  if (sumValues(sheet.instincts) > INSTINCT_TOTAL) {
    add('instinct_total', 'instincts', `Instintos somam mais que ${INSTINCT_TOTAL} pontos.`, true)
  }
  const apt: Record<string, number> = { ...sheet.knowledge, ...sheet.practices }
  for (const [key, v] of Object.entries(apt)) {
    const field = `${key in sheet.knowledge ? 'knowledge' : 'practices'}.${key}`
    const label = APTITUDE_LABELS[key as keyof typeof APTITUDE_LABELS]
    if (v > APTITUDE_MAX) add('aptitude_max', field, `${label} não pode passar de ${APTITUDE_MAX}.`, true)
    if (v < 0) add('aptitude_min', field, `${label} não pode ser negativa.`, true)
  }
  if (sumValues(apt) > APTITUDE_TOTAL) {
    add('aptitude_total', 'aptitudes', `Aptidões somam mais que ${APTITUDE_TOTAL} pontos.`, true)
  }
}

function checkTug(sheet: Sheet, add: Add) {
  const t = sheet.tugOfWar
  if (t.determinationLevel + t.assimilationLevel !== 10) {
    add('tug_sum', 'tugOfWar', 'Determinação e Assimilação devem somar 10.')
  }
  if (t.determinationPoints < 0 || t.determinationPoints > t.determinationLevel) {
    add('determination_points', 'tugOfWar.determinationPoints', 'Pontos de Determinação fora do intervalo do nível.')
  }
  if (t.assimilationPoints < 0 || t.assimilationPoints > t.assimilationLevel) {
    add('assimilation_points', 'tugOfWar.assimilationPoints', 'Pontos de Assimilação fora do intervalo do nível.')
  }
}

function checkRest(sheet: Sheet, add: Add) {
  if (xpAvailable(sheet) < 0) add('xp_overspent', 'characteristics', 'XP gasto maior que o total.')
  if (sheet.inventory.body.length > BODY_LIMIT) add('body_full', 'inventory.body', `O corpo comporta ${BODY_LIMIT} itens.`)
  if (sheet.inventory.backpack.length > BACKPACK_LIMIT) {
    add('backpack_full', 'inventory.backpack', `A mochila comporta ${BACKPACK_LIMIT} itens.`)
  }
  if (sheet.identity.personalPurposes.length !== 2) {
    add('personal_purposes', 'identity.personalPurposes', 'São necessários 2 propósitos pessoais.')
  }
  if (sheet.identity.collectivePurposes.length !== 2) {
    add('collective_purposes', 'identity.collectivePurposes', 'São necessários 2 propósitos coletivos.')
  }
}

export function validateSheet(sheet: Sheet, opts: ValidateOptions = {}): SheetIssue[] {
  const blockCreation = opts.blockCreationLimits ?? true
  const issues: SheetIssue[] = []
  const add: Add = (code, field, message, creation = false) => {
    issues.push({ code, field, message, blocking: creation ? blockCreation : true })
  }
  checkAttributes(sheet, add)
  checkTug(sheet, add)
  checkRest(sheet, add)
  return issues
}
