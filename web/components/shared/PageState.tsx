export function PageState({ loading, error, empty, onRetry }: { loading?: boolean; error?: string; empty?: boolean; onRetry?: () => void }) {
  if (loading) return <div className="page-state">加载中...</div>
  if (error) {
    return (
      <div className="page-state error">
        <div>{error}</div>
        {onRetry ? (
          <button className="btn" type="button" onClick={onRetry}>
            重试
          </button>
        ) : null}
      </div>
    )
  }
  if (empty) return <div className="page-state">暂无数据</div>
  return null
}
