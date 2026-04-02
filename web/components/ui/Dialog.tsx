'use client'

import { PropsWithChildren } from 'react'

export function Dialog({
  open,
  title,
  onClose,
  children,
}: PropsWithChildren<{
  open: boolean
  title: string
  onClose: () => void
}>) {
  if (!open) return null
  return (
    <div className="dialog-mask" onClick={onClose}>
      <div className="dialog" onClick={(e) => e.stopPropagation()}>
        <div className="dialog-header">
          <b>{title}</b>
          <button className="btn ghost" type="button" onClick={onClose}>
            关闭
          </button>
        </div>
        <div className="dialog-body">{children}</div>
      </div>
    </div>
  )
}
