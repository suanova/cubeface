import type { ReactNode } from 'react'

interface PageHeadProps {
  kicker: string
  title: string
  lead: string
  notes: string[]
  actions?: ReactNode
}

export function PageHead({ kicker, title, lead, notes, actions }: PageHeadProps) {
  return (
    <div className='ml-page-head'>
      <div className='ml-page-head-copy'>
        <p className='ml-kicker'>{kicker}</p>
        <h1 className='ml-title'>{title}</h1>
        <p className='ml-lead'>{lead}</p>
        {notes.length > 0 ? (
          <p className='ml-head-note'>
            {notes.map((note) => (
              <span key={note}>{note}</span>
            ))}
          </p>
        ) : null}
      </div>
      {actions ? <div className='ml-page-actions'>{actions}</div> : null}
    </div>
  )
}
