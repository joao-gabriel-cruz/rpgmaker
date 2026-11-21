import { useState } from 'react'
import { InvitePanel } from '../campaigns/InvitePanel'
import { MembersSection } from '../campaigns/MembersSection'
import { SettingsPanel } from '../campaigns/SettingsPanel'
import type { Campaign } from '../campaigns/campaigns'
import { RulesEditor } from '../effects/RulesEditor'
import { Modal } from './Modal'

const TABS = [
  { id: 'members', label: 'Membros e convites' },
  { id: 'settings', label: 'Configurações' },
  { id: 'rules', label: 'Regras de efeitos' },
] as const

type TabId = (typeof TABS)[number]['id']

function Body({ tab, campaign }: { tab: TabId; campaign: Campaign }) {
  if (tab === 'members') {
    return (
      <div className="flex flex-col gap-4">
        <MembersSection campaignId={campaign.id} isGm />
        <InvitePanel campaignId={campaign.id} />
      </div>
    )
  }
  return tab === 'settings' ? <SettingsPanel campaign={campaign} /> : <RulesEditor campaignId={campaign.id} />
}

/** Mestre dialog (from the top bar gear): members and invites, campaign settings, effect rules. */
export function SettingsDialog({ campaign, onClose }: { campaign: Campaign; onClose: () => void }) {
  const [tab, setTab] = useState<TabId>('members')
  return (
    <Modal title="Gerenciar campanha" onClose={onClose} wide>
      <div role="tablist" aria-label="Gerenciar campanha" className="mb-4 flex gap-1 border-b border-zinc-800">
        {TABS.map((t) => (
          <button key={t.id} role="tab" aria-selected={tab === t.id} onClick={() => setTab(t.id)}
            className={`-mb-px border-b-2 px-3 py-2 text-sm transition-colors ${tab === t.id ? 'border-emerald-500 text-white' : 'border-transparent text-zinc-400 hover:text-zinc-200'}`}>
            {t.label}
          </button>
        ))}
      </div>
      <Body tab={tab} campaign={campaign} />
    </Modal>
  )
}
