import { useState } from 'react'
import { Modal } from '../table/Modal'
import { BlankSceneForm } from './BlankSceneForm'
import { TemplateGallery } from './TemplateGallery'

interface Props {
  campaignId: string
  onClose: () => void
  onCreateBlank: (name: string, size: number) => Promise<unknown>
  onCreateFromTemplate: (name: string, templateId: string, seed: number) => Promise<unknown>
}

type Tab = 'templates' | 'blank'

const tabClass = (on: boolean) =>
  `rounded-lg px-3 py-1.5 text-sm transition-[background-color,color] ${on ? 'bg-zinc-800 text-white' : 'text-zinc-400 hover:text-white'}`

/** "Nova cena" dialog: template gallery tab and blank scene tab. Closes after a successful creation. */
export function NewSceneDialog({ campaignId, onClose, onCreateBlank, onCreateFromTemplate }: Props) {
  const [tab, setTab] = useState<Tab>('templates')
  const [error, setError] = useState('')
  const run = (p: Promise<unknown>) => p.then(onClose, (e: unknown) => setError(e instanceof Error ? e.message : 'Não foi possível criar a cena.'))
  return (
    <Modal title="Nova cena" onClose={onClose} wide>
      <div role="tablist" aria-label="Tipo de cena" className="mb-3 flex gap-1">
        <button role="tab" type="button" aria-selected={tab === 'templates'} className={tabClass(tab === 'templates')} onClick={() => setTab('templates')}>Modelos</button>
        <button role="tab" type="button" aria-selected={tab === 'blank'} className={tabClass(tab === 'blank')} onClick={() => setTab('blank')}>Em branco</button>
      </div>
      <div role="tabpanel">
        {tab === 'templates'
          ? <TemplateGallery campaignId={campaignId} error={error} onCreate={(n, id, seed) => run(onCreateFromTemplate(n, id, seed))} />
          : <BlankSceneForm error={error} onCreate={(n, s) => run(onCreateBlank(n, s))} />}
      </div>
    </Modal>
  )
}
