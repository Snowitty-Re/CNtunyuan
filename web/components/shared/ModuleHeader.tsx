import { ReactNode } from 'react'

export function ModuleHeader({
  title,
  desc,
  right,
}: {
  title: string
  desc?: string
  right?: ReactNode
}) {
  return (
    <div className="module-header">
      <div>
        <h2 className="page-title">{title}</h2>
        {desc ? <p className="page-desc">{desc}</p> : null}
      </div>
      {right ? <div className="module-header-right">{right}</div> : null}
    </div>
  )
}

