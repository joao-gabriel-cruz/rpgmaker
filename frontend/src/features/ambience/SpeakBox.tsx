import { useEffect, useRef, useState, type FormEvent } from 'react'
import { MAX_SAY } from './presence'

interface Props {
  name: string
  onSay: (text: string) => Promise<void>
  onClose: () => void
}

/** One-line "fala em personagem" field: Enter sends, Esc closes; at most 140 characters. */
export function SpeakBox({ name, onSay, onClose }: Props) {
  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const input = useRef<HTMLInputElement>(null)
  useEffect(() => input.current?.focus(), [])
  const submit = (e: FormEvent) => {
    e.preventDefault()
    const t = text.trim()
    if (!t) return
    onSay(t).then(onClose, (err: Error) => setError(err.message))
  }
  return (
    <form onSubmit={submit} aria-label="Falar em personagem" className="flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-1 rounded-xl bg-zinc-900/95 p-2 shadow-lg shadow-black/40 ring-1 ring-white/10">
      <label className="text-[11px] font-medium uppercase tracking-wide text-zinc-400" htmlFor="speak-input">Falar como {name}</label>
      <div className="flex items-center gap-2">
        <input id="speak-input" ref={input} value={text} maxLength={MAX_SAY} autoComplete="off" placeholder="O que seu personagem diz?"
          onChange={(e) => setText(e.target.value)} onKeyDown={(e) => e.key === 'Escape' && onClose()}
          className="min-w-0 flex-1 rounded-md bg-zinc-800 px-2 py-1.5 text-sm text-zinc-100 placeholder:text-zinc-500 focus-visible:outline-2 focus-visible:outline-emerald-400" />
        <span aria-hidden className={`w-8 text-right text-xs tabular-nums ${text.length > MAX_SAY - 20 ? 'text-amber-400' : 'text-zinc-500'}`}>{MAX_SAY - text.length}</span>
      </div>
      {error && <p role="alert" className="text-xs text-red-300">{error}</p>}
    </form>
  )
}
