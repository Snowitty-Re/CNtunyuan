'use client'

export type Notice = {
  type: 'success' | 'error' | 'info'
  text: string
}

export function NoticeBar({ notice, onClose }: { notice: Notice | null; onClose?: () => void }) {
  if (!notice) return null
  return (
    <div className={`notice ${notice.type}`}>
      <span>{notice.text}</span>
      {onClose ? (
        <button className="btn ghost" type="button" onClick={onClose}>
          关闭
        </button>
      ) : null}
    </div>
  )
}
