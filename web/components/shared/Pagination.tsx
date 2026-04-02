export function Pagination({ page, pageSize, total, onChange }: { page: number; pageSize: number; total: number; onChange: (p: number) => void }) {
  const totalPages = Math.max(1, Math.ceil((total || 0) / (pageSize || 20)))
  return (
    <div className="pagination">
      <button className="btn ghost" type="button" onClick={() => onChange(Math.max(1, page - 1))} disabled={page <= 1}>
        上一页
      </button>
      <span>
        第 {page} / {totalPages} 页，共 {total || 0} 条
      </span>
      <button className="btn ghost" type="button" onClick={() => onChange(Math.min(totalPages, page + 1))} disabled={page >= totalPages}>
        下一页
      </button>
    </div>
  )
}
