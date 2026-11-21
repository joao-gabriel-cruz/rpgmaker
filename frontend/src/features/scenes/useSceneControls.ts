import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'
import { activateScene, createScene, deleteScene, listScenes, renameScene, type SceneSummary } from './scenes'

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

export interface SceneControls {
  scenes: SceneSummary[]
  loading: boolean
  /** Scene on the map: the Mestre's pick, else the active scene; players always see the active one. */
  selectedId: string | null
  selected: SceneSummary | null
  error: string
  select: (id: string) => void
  create: (name: string, size: number) => void
  rename: (id: string, name: string) => void
  remove: (id: string) => void
  activate: (id: string) => void
}

/** Scene list + Mestre actions, shared by the top bar selector and the map. */
export function useSceneControls(campaignId: string, isGm: boolean): SceneControls {
  const { scenes, loading, refresh } = useSceneList(campaignId)
  const [picked, setPicked] = useState<string | null>(null)
  const active = scenes.find((s) => s.isActive)
  const selectedId = (isGm ? scenes.find((s) => s.id === picked)?.id ?? active?.id ?? scenes[0]?.id : active?.id) ?? null
  const act = useMutation({ mutationFn: (fn: () => Promise<unknown>) => fn(), onSuccess: refresh })
  return {
    scenes, loading, selectedId, selected: scenes.find((s) => s.id === selectedId) ?? null,
    error: act.error?.message ?? '',
    select: setPicked,
    create: (name, size) => act.mutate(async () => setPicked((await createScene(campaignId, name, size)).id)),
    rename: (id, name) => act.mutate(() => renameScene(campaignId, id, name)),
    remove: (id) => act.mutate(() => deleteScene(campaignId, id)),
    activate: (id) => act.mutate(async () => { await activateScene(campaignId, id); setPicked(id) }),
  }
}
