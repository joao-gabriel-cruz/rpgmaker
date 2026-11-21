import { useState } from 'react'
import { Link } from 'react-router'
import { roleLabel, useCampaigns, type Campaign } from './campaigns'
import { NewCampaignForm } from './NewCampaignForm'

function CampaignItem({ c }: { c: Campaign }) {
  return (
    <li>
      <Link to={`/campaigns/${c.id}`}
        className="flex items-center justify-between rounded bg-zinc-900 px-4 py-3 hover:bg-zinc-800">
        <span className="font-medium">{c.name}</span>
        <span className="text-sm text-zinc-400">{c.systemName} · {roleLabel(c.role)}</span>
      </Link>
    </li>
  )
}

export function CampaignsList() {
  const { data: campaigns, isPending, isError } = useCampaigns()
  const [creating, setCreating] = useState(false)

  return (
    <section className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <h2 className="text-xl font-semibold">Minhas campanhas</h2>
        <button onClick={() => setCreating((v) => !v)}
          className="rounded bg-emerald-600 px-3 py-1 text-sm font-medium hover:bg-emerald-500">
          Nova campanha
        </button>
      </div>
      {creating && <NewCampaignForm onDone={() => setCreating(false)} />}
      {isPending && <p className="text-zinc-400">Carregando…</p>}
      {isError && <p role="alert" className="text-red-400">Não foi possível carregar as campanhas</p>}
      {campaigns?.length === 0 && <p className="text-zinc-400">Você ainda não participa de nenhuma campanha.</p>}
      <ul className="flex max-w-2xl flex-col gap-2">
        {campaigns?.map((c) => <CampaignItem key={c.id} c={c} />)}
      </ul>
    </section>
  )
}
