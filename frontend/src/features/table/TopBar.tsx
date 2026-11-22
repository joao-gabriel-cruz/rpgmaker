import { ArrowLeft, Clapperboard, Keyboard, Settings } from 'lucide-react'
import { Link } from 'react-router'
import type { Campaign } from '../campaigns/campaigns'
import { roleLabel } from '../campaigns/campaigns'
import { SoundControl } from '../sound/SoundControl'
import { spotlight, useSpotlight } from '../fx/spotlight'
import { DiseaseAlert } from '../refuge/DiseaseAlert'
import { RoundTracker } from '../rounds/RoundTracker'
import type { SceneControls } from '../scenes/useSceneControls'
import { SessionBar } from '../sessions/SessionBar'
import { PresenceAvatars } from './PresenceAvatars'
import { SceneMenu } from './SceneMenu'
import { ICON_BTN } from './surface'

interface Props {
  campaign: Campaign
  scenes: SceneControls
  onOpenSettings: () => void
  onOpenHelp: () => void
}

function SpotlightToggle() {
  const on = useSpotlight()
  return (
    <button type="button" aria-label="Holofote de turno" aria-pressed={on} title="Holofote: a câmera de todos acompanha quem age"
      onClick={() => spotlight.set(!on)} className={`${ICON_BTN} h-9 w-9 ${on ? 'bg-emerald-900/60 text-emerald-300' : ''}`}>
      <Clapperboard size={18} strokeWidth={1.5} aria-hidden />
    </button>
  )
}

const Divider = () => <span aria-hidden className="h-5 w-px bg-zinc-800" />

function TopActions({ campaignId, isGm, onOpenHelp, onOpenSettings }: { campaignId: string; isGm: boolean; onOpenHelp: () => void; onOpenSettings: () => void }) {
  return (
    <div className="ml-auto flex shrink-0 items-center gap-2">
      <PresenceAvatars campaignId={campaignId} />
      <SoundControl />
      <button type="button" aria-label="Atalhos de teclado" title="Atalhos de teclado (?)" onClick={onOpenHelp} className={`${ICON_BTN} h-9 w-9`}>
        <Keyboard size={18} strokeWidth={1.5} aria-hidden />
      </button>
      {isGm && <SpotlightToggle />}
      {isGm && (
        <button type="button" aria-label="Gerenciar campanha" title="Membros, configurações e regras" onClick={onOpenSettings} className={`${ICON_BTN} h-9 w-9`}>
          <Settings size={18} strokeWidth={1.5} aria-hidden />
        </button>
      )}
    </div>
  )
}

/** Compact 48px bar: back, campaign, scene, session, round, presence, shortcuts and (Mestre) the gear menu. */
export function TopBar({ campaign, scenes, onOpenSettings, onOpenHelp }: Props) {
  const isGm = campaign.role === 'gm'
  return (
    <header className="z-30 flex h-12 shrink-0 items-center gap-3 whitespace-nowrap border-b border-zinc-800 bg-zinc-950 px-2">
      <Link to="/" aria-label="Voltar para Minhas campanhas" className={`${ICON_BTN} h-9 w-9`}><ArrowLeft size={18} aria-hidden /></Link>
      <div className="min-w-0 max-w-80 shrink leading-tight compact:max-w-64">
        <h1 title={campaign.name} className="truncate text-sm font-bold tracking-tight">{campaign.name}</h1>
        <p className="truncate text-[11px] text-zinc-400 compact:sr-only">{campaign.systemName} · Você é {roleLabel(campaign.role)}</p>
      </div>
      <Divider />
      <div className="min-w-0 shrink"><SceneMenu scenes={scenes} isGm={isGm} /></div>
      <Divider />
      <SessionBar campaignId={campaign.id} isGm={isGm} />
      <Divider />
      <RoundTracker campaignId={campaign.id} isGm={isGm} />
      <DiseaseAlert campaignId={campaign.id} />
      <TopActions campaignId={campaign.id} isGm={isGm} onOpenHelp={onOpenHelp} onOpenSettings={onOpenSettings} />
    </header>
  )
}
