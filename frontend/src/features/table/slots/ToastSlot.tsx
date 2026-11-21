import { ActionToasts } from '../../actions/ActionToasts'
import { CardsInbox } from '../../actions/CardsInbox'
import { AppliedToast } from '../../effects/AppliedToast'
import { ActionFxDriver } from '../../fx/ActionFxDriver'

/** Slot "toasts": short event notices (effect applied, action results) plus the headless action fx driver. */
export function ToastSlot({ isGm = false, sceneId = null }: { isGm?: boolean; sceneId?: string | null }) {
  return (
    <>
      <AppliedToast />
      <ActionToasts />
      {isGm && <CardsInbox />}
      <ActionFxDriver isGm={isGm} sceneId={sceneId} />
    </>
  )
}
