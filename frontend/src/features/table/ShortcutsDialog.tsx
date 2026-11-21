import { CAMERA_KEYS } from './camera/cameraKeys'
import { Modal } from './Modal'
import { Kbd } from './Tip'
import { toolsFor } from './tools'

const EXTRA: [string, string][] = [
  ['Espaço', 'Abrir o compositor de rolagem do personagem ativo'],
  ['Enter', 'Rolar de novo a última combinação (fora de botões e campos)'],
  ['← ↑ → ↓', 'No compositor: navegar entre Instintos e Aptidões (Enter rola, Esc fecha)'],
  ['Esc', 'Voltar para Mover'],
  ['[', 'Recolher a barra lateral'],
  [']', 'Expandir a barra lateral'],
  ['Alt + clique', 'Ping rápido em qualquer ferramenta'],
  ['?', 'Mostrar esta lista'],
]

const OBJECT_KEYS: [string, string][] = [
  ['Shift + R', 'Girar o objeto (ferramentas Objetos e Muro)'],
  ['− / +', 'Diminuir / aumentar o pincel (Terreno)'],
  ['1 – 6', 'Sub-ferramenta do Terreno (pintar, subir, descer, altura, preencher, apagar)'],
  ['1 – 3', 'Modo dos Objetos (colocar, selecionar, remover)'],
  ['Ctrl/⌘ + Z', 'Desfazer a última edição do mapa'],
  ['Shift + Ctrl/⌘ + Z', 'Refazer'],
]

function Row({ keys, label }: { keys: string; label: string }) {
  return (
    <li className="flex items-center justify-between gap-4 py-1 text-sm">
      <span className="text-zinc-300">{label}</span>
      <Kbd>{keys}</Kbd>
    </li>
  )
}

/** Lists every table shortcut, driven by the same tool definitions the rail uses. */
export function ShortcutsDialog({ isGm, onClose }: { isGm: boolean; onClose: () => void }) {
  const tools = toolsFor(isGm).map((t) => [t.key.toUpperCase(), t.label] as [string, string])
  return (
    <Modal title="Atalhos de teclado" onClose={onClose}>
      <ul className="divide-y divide-zinc-800">
        {[...tools, ...CAMERA_KEYS, ...(isGm ? OBJECT_KEYS : []), ...EXTRA].map(([k, l]) => <Row key={k + l} keys={k} label={l} />)}
      </ul>
      <p className="mt-3 text-xs text-zinc-500">Os atalhos de uma tecla são ignorados enquanto você digita em um campo.</p>
    </Modal>
  )
}
