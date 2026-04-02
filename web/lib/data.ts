export function listFrom<T>(data: any): { list: T[]; total: number } {
  if (Array.isArray(data)) {
    return { list: data, total: data.length }
  }
  if (data && Array.isArray(data.list)) {
    return { list: data.list, total: Number(data.total || data.list.length || 0) }
  }
  return { list: [], total: 0 }
}

export function fmtTime(v?: string | null): string {
  if (!v) return '-'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  return d.toLocaleString('zh-CN', { hour12: false })
}

export function joinLocation(row: Record<string, any>): string {
  return [row.province, row.city, row.district, row.address].filter(Boolean).join(' ') || row.location || '-'
}
