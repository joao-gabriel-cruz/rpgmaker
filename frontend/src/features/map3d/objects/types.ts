import type { ComponentType } from 'react'
import type { ObjectProps } from './catalogData'

export interface PrefabProps {
  props: ObjectProps
}

export type PrefabComponent = ComponentType<PrefabProps>
