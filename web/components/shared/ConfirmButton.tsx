'use client'

import { useState } from 'react'
import { Dialog } from '@/components/ui/Dialog'

export function ConfirmButton({ text, message, onConfirm, className }: { text: string; message: string; onConfirm: () => void; className?: string }) {
  const [open, setOpen] = useState(false)
  return (
    <>
      <button type="button" className={className || 'btn danger'} onClick={() => setOpen(true)}>
        {text}
      </button>
      <Dialog open={open} title="请确认操作" onClose={() => setOpen(false)}>
        <div className="grid">
          <div>{message}</div>
          <div className="row">
            <button className="btn ghost" type="button" onClick={() => setOpen(false)}>
              取消
            </button>
            <button
              className="btn danger"
              type="button"
              onClick={() => {
                onConfirm()
                setOpen(false)
              }}
            >
              确认
            </button>
          </div>
        </div>
      </Dialog>
    </>
  )
}
