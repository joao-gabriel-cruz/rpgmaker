import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ClaudeChangesList } from './ClaudeChangesList'
import { createLink, getLink, relativeTime, revokeLink } from './mcp'

function GeneratedLink({ url }: { url: string }) {
  const [copy, setCopy] = useState<'idle' | 'done' | 'failed'>('idle')
  const copyUrl = () =>
    navigator.clipboard
      .writeText(url)
      .then(() => {
        setCopy('done')
        setTimeout(() => setCopy('idle'), 2000)
      })
      .catch(() => setCopy('failed'))
  return (
    <div className="flex flex-col gap-2">
      <div className="flex max-w-xl items-center gap-2">
        <input readOnly aria-label="Link do Claude" value={url} onFocus={(e) => e.target.select()}
          className="flex-1 rounded bg-zinc-800 px-3 py-2 text-sm" />
        <button onClick={copyUrl}
          className="rounded bg-zinc-800 px-3 py-2 text-sm hover:bg-zinc-700">
          {copy === 'done' ? 'Copiado!' : 'Copiar'}
        </button>
      </div>
      {copy === 'failed' && (
        <p role="alert" className="text-xs text-red-400">
          Não foi possível copiar; selecione o link e copie manualmente.
        </p>
      )}
      <p className="text-xs text-amber-400">
        Guarde este link como uma senha: quem tiver ele age como Mestre nesta campanha. Ele só aparece agora.
      </p>
    </div>
  )
}

function useClaudeLink(campaignId: string) {
  const client = useQueryClient()
  const key = ['mcp-link', campaignId]
  const status = useQuery({ queryKey: key, queryFn: () => getLink(campaignId) })
  const refresh = () => client.invalidateQueries({ queryKey: key })
  const generate = useMutation({ mutationFn: () => createLink(campaignId), onSuccess: refresh })
  const revoke = useMutation({ mutationFn: () => revokeLink(campaignId), onSuccess: () => { generate.reset(); return refresh() } })
  return { status, generate, revoke, exists: status.data?.exists ?? false }
}

function LinkUsage({ lastUsedAt }: { lastUsedAt?: string | null }) {
  return (
    <p className="text-xs text-zinc-400">
      {lastUsedAt ? `Último uso: ${relativeTime(lastUsedAt)}` : 'Nunca usado'}
    </p>
  )
}

function ConnectorSteps() {
  return (
    <ul className="list-disc pl-5 text-xs text-zinc-400">
      <li>claude.ai: Configurações → Conectores → Adicionar conector personalizado → cole o link.</li>
      <li>Claude Desktop: Configurações → Conectores → Adicionar conector personalizado → cole o link.</li>
    </ul>
  )
}

function LinkSection({ campaignId }: { campaignId: string }) {
  const { status, generate, revoke, exists } = useClaudeLink(campaignId)
  const onGenerate = () => {
    if (exists && !window.confirm('O link atual vai parar de funcionar. Continuar?')) return
    generate.mutate()
  }
  const error = generate.error ?? revoke.error

  return (
    <section className="flex flex-col gap-3">
      <h3 className="text-sm font-medium text-zinc-200">Conectar Claude</h3>
      <div className="flex items-center gap-2">
        <button onClick={onGenerate} disabled={generate.isPending || status.isPending}
          className="w-fit rounded bg-emerald-600 px-3 py-1 text-sm font-medium hover:bg-emerald-500">
          {exists ? 'Gerar novo link' : 'Gerar link'}
        </button>
        {exists && (
          <button onClick={() => revoke.mutate()} disabled={revoke.isPending}
            className="rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700">
            Revogar link
          </button>
        )}
      </div>
      {error && <p role="alert" className="text-sm text-red-400">{error.message}</p>}
      {generate.data && <GeneratedLink url={generate.data.url} />}
      {exists && <LinkUsage lastUsedAt={status.data?.lastUsedAt} />}
      <ConnectorSteps />
    </section>
  )
}

export function ClaudePanel({ campaignId }: { campaignId: string }) {
  return (
    <div className="flex flex-col gap-5">
      <LinkSection campaignId={campaignId} />
      <ClaudeChangesList campaignId={campaignId} />
    </div>
  )
}
