import { usePresenceState } from './PresenceContext'

const clock = (at: number) => new Date(at).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })

/** Sidebar "Falas": what characters said in character this session (newest at the bottom). */
export function SpeechLog() {
  const log = usePresenceState()?.log ?? []
  return (
    <div className="flex flex-col gap-2">
      <h2 className="text-lg font-semibold">Falas</h2>
      {log.length === 0 && <p className="text-sm text-zinc-400">Ninguém falou ainda. Use <kbd className="rounded bg-zinc-800 px-1">'</kbd> para falar com seu personagem.</p>}
      <ul className="flex flex-col gap-1.5" aria-live="polite">
        {log.map((e) => (
          <li key={e.id} className="rounded-lg bg-zinc-800/60 px-3 py-1.5 text-sm">
            <span className="flex items-baseline justify-between gap-2">
              <strong className="truncate font-semibold text-zinc-100">{e.name}</strong>
              <time className="text-[11px] text-zinc-500">{clock(e.at)}</time>
            </span>
            <span className="block break-words text-zinc-300">“{e.text}”</span>
          </li>
        ))}
      </ul>
    </div>
  )
}
