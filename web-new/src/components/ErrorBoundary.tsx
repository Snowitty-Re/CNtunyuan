import { Component, ErrorInfo, ReactNode } from 'react';
import { Button, Result } from 'antd';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
}

/**
 * 错误边界组件
 * 捕获 React 组件树中的错误，防止整个应用崩溃
 */
export default class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by ErrorBoundary:', error, errorInfo);
    this.setState({ error, errorInfo });

    // 这里可以发送错误到监控服务
    // reportError(error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined });
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      // 自定义 fallback UI
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <Result
          status="500"
          title="出错了"
          subTitle="抱歉，页面发生了错误。请刷新页面重试。"
          extra={
            <div style={{ textAlign: 'center' }}>
              <Button type="primary" onClick={this.handleReset}>
                刷新页面
              </Button>
              {process.env.NODE_ENV === 'development' && this.state.error && (
                <div style={{ marginTop: 24, textAlign: 'left' }}>
                  <details style={{ whiteSpace: 'pre-wrap' }}>
                    <summary>错误详情</summary>
                    <pre>{this.state.error.toString()}</pre>
                    <pre>{this.state.errorInfo?.componentStack}</pre>
                  </details>
                </div>
              )}
            </div>
          }
        />
      );
    }

    return this.props.children;
  }
}
