export function PageState({ loading, error, empty, onRetry }: { loading?: boolean; error?: string; empty?: boolean; onRetry?: () => void }) {
  if (loading) {
    return (
      <div className="page-state">
        <div className="state-badge">处理中</div>
        <div className="state-title">页面正在加载</div>
        <div className="state-desc">请稍候，正在同步最新数据。</div>
      </div>
    )
  }
  if (error) {
    return (
      <div className="page-state error">
        <div className="state-badge">异常</div>
        <div className="state-title">加载失败</div>
        <div className="state-desc">{error}</div>
        {onRetry ? (
          <button className="btn" type="button" onClick={onRetry}>
            重试
          </button>
        ) : null}
      </div>
    )
  }
  if (empty) {
    return (
      <div className="page-state">
        <div className="state-badge">空数据</div>
        <div className="state-title">暂无可展示内容</div>
        <div className="state-desc">可以调整筛选条件，或先创建第一条记录。</div>
      </div>
    )
  }
  return null
}
