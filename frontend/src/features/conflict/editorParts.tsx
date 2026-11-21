import type { ReactNode } from 'react'

export const input = 'rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm'
export const smallBtn = 'rounded bg-zinc-800 px-2 py-1 text-xs hover:bg-zinc-700'

/** Replaces item `i` with a patch. */
export const patchAt = <T,>(list: T[], i: number, patch: Partial<T>): T[] => list.map((x, j) => (j === i ? { ...x, ...patch } : x))

export function Group({ title, onAdd, children }: { title: string; onAdd: () => void; children: ReactNode }) {
  return (
    <fieldset className="flex flex-col gap-2 rounded border border-zinc-800 p-3">
      <legend className="px-1 text-sm font-semibold">{title}</legend>
      {children}
      <button type="button" className={`${smallBtn} self-start`} onClick={onAdd}>Adicionar</button>
    </fieldset>
  )
}

export function NumberField({ label, value, onChange, min = 0 }: { label: string; value: number; onChange: (n: number) => void; min?: number }) {
  return (
    <label className="flex items-center gap-1 text-xs">
      {label}
      <input type="number" min={min} value={value} onChange={(e) => onChange(Number(e.target.value))} className={`${input} w-16`} />
    </label>
  )
}
