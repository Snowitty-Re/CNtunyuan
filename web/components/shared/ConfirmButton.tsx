'use client'

import { MouseEventHandler } from 'react'

export function ConfirmButton({ text, message, onConfirm, className }: { text: string; message: string; onConfirm: MouseEventHandler<HTMLButtonElement>; className?: string }) {
  return (
    <button
      type="button"
      className={className || 'btn danger'}
      onClick={(e) => {
        if (window.confirm(message)) onConfirm(e)
      }}
    >
      {text}
    </button>
  )
}
