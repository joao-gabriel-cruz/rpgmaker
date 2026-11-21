import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef, useState } from 'react'
import { TerrainEditor } from '../map3d/TerrainEditor'
import { TerrainView } from '../map3d/TerrainView'
import { useMe } from '../auth/auth'
import { useRealtime } from '../realtime/realtimeContext'
import { GmPanel } from './GmPanel'
import { SceneSelector } from './SceneSelector'
import { activateScene, createScene, deleteScene, listScenes, renameScene } from './scenes'
import { useLiveScene } from './useLiveScene'
import { DrawingLayer } from '../drawings/DrawingLayer'
import { DrawingToolbar } from '../drawings/DrawingToolbar'
import { useDrawings } from '../drawings/useDrawings'
import { TokenLayer } from '../tokens/TokenLayer'
import { TokensPanel } from '../tokens/TokensPanel'
import { useSceneTokens } from '../tokens/useSceneTokens'
import type { MapState } from './scenes'

function SceneCanvas({ campaignId, sceneId, isGm }: { campaignId: string; sceneId: string; isGm: boolean }) {
  const { state, error, paint, editObjects, command, replacedAt } = useLiveScene(campaignId, sceneId)
  const box = useRef<HTMLDivElement>(null)
  useEffect(() => {
    // Short fade/scale-in so a map.replaced does not look like a glitch.
    if (replacedAt) box.current?.animate?.([{ opacity: 0.2, transform: 'scale(0.97)' }, { opacity: 1, transform: 'scale(1)' }], { duration: 400, easing: 'ease-out' })
  }, [replacedAt])
  return (
    <div className="flex flex-col gap-1">
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
      {isGm && state && <GmPanel campaignId={campaignId} sceneId={sceneId} draft={!!state.draft} onCommand={command} />}
      <div ref={box}>
      {state && <SceneWithTokens campaignId={campaignId} sceneId={sceneId} isGm={isGm} state={state} paint={paint} editObjects={editObjects} />}
      {!state && !error && <p className="p-4 text-zinc-400">Carregando mapa…</p>}
      </div>
    </div>
  )
}

type Live = ReturnType<typeof useLiveScene>
interface WithTokens extends Pick<Live, 'paint' | 'editObjects'> {
  campaignId: string
  sceneId: string
  isGm: boolean
  state: MapState
}

function SceneWithTokens({ campaignId, sceneId, isGm, state, paint, editObjects }: WithTokens) {
  const userId = useMe().data?.id ?? ''
  const tokens = useSceneTokens(campaignId, sceneId, state.grid)
  const draw = useDrawings(campaignId, sceneId, userId)
  const layer = (
    <>
      <TokenLayer grid={state.grid} tokens={tokens} userId={userId} isGm={isGm} />
      <DrawingLayer grid={state.grid} draw={draw} />
    </>
  )
  return (
    <>
      <TokensPanel campaignId={campaignId} tokens={tokens} userId={userId} isGm={isGm} />
      <DrawingToolbar draw={draw} isGm={isGm} />
      <div className="h-[70vh] overflow-hidden rounded border border-zinc-800">
        {isGm
          ? <TerrainEditor grid={state.grid} onPatch={paint} objects={state.objects} onObjectPatch={editObjects}>{layer}</TerrainEditor>
          : <TerrainView grid={state.grid} objects={state.objects}>{layer}</TerrainView>}
      </div>
    </>
  )
}

const scenesKey = (campaignId: string) => ['campaigns', campaignId, 'scenes']

function useSceneList(campaignId: string) {
  const queryClient = useQueryClient()
  const { subscribe } = useRealtime()
  const query = useQuery({ queryKey: scenesKey(campaignId), queryFn: () => listScenes(campaignId) })
  useEffect(() => subscribe('scenes.changed', () => queryClient.invalidateQueries({ queryKey: scenesKey(campaignId) })),
    [subscribe, queryClient, campaignId])
  const refresh = () => queryClient.invalidateQueries({ queryKey: scenesKey(campaignId) })
  return { scenes: Array.isArray(query.data) ? query.data : [], loading: query.isPending, refresh }
}

function GmMap({ campaignId }: { campaignId: string }) {
  const { scenes, refresh } = useSceneList(campaignId)
  const [picked, setPicked] = useState<string | null>(null)
  const selectedId = scenes.find((s) => s.id === picked)?.id ?? scenes.find((s) => s.isActive)?.id ?? scenes[0]?.id ?? null
  const act = useMutation({ mutationFn: (fn: () => Promise<unknown>) => fn(), onSuccess: refresh })
  return (
    <div className="flex flex-col gap-3">
      <SceneSelector
        scenes={scenes}
        selectedId={selectedId}
        onSelect={setPicked}
        onCreate={(name, size) => act.mutate(async () => setPicked((await createScene(campaignId, name, size)).id))}
        onRename={(id, name) => act.mutate(() => renameScene(campaignId, id, name))}
        onDelete={(id) => act.mutate(() => deleteScene(campaignId, id))}
        onActivate={(id) => act.mutate(async () => { await activateScene(campaignId, id); setPicked(id) })}
      />
      {act.error && <p role="alert" className="text-sm text-red-400">{act.error.message}</p>}
      {selectedId && <SceneCanvas key={selectedId} campaignId={campaignId} sceneId={selectedId} isGm />}
    </div>
  )
}

function PlayerMap({ campaignId }: { campaignId: string }) {
  const { scenes, loading } = useSceneList(campaignId)
  const active = scenes.find((s) => s.isActive)
  if (loading) return <p className="text-zinc-400">Carregando…</p>
  if (!active) return <p className="text-zinc-400">O Mestre ainda não ativou nenhuma cena.</p>
  return (
    <div className="flex flex-col gap-2">
      <p className="text-sm text-zinc-400">Cena atual: {active.name}</p>
      <SceneCanvas key={active.id} campaignId={campaignId} sceneId={active.id} isGm={false} />
    </div>
  )
}

/** The "Mapa" area of the campaign page; must render inside CampaignSocketProvider. */
export function MapArea({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  return (
    <div className="flex flex-col gap-2">
      <h2 className="text-xl font-semibold">Mapa</h2>
      {isGm ? <GmMap campaignId={campaignId} /> : <PlayerMap campaignId={campaignId} />}
    </div>
  )
}
