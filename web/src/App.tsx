import { RouterProvider } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import ErrorBoundary from './components/ErrorBoundary'
import router from './router'

dayjs.locale('zh-cn')

const themeConfig = {
  token: {
    colorPrimary: '#e67e22',
    borderRadius: 6,
    fontSize: 14,
  },
}

export default function App() {
  return (
    <ConfigProvider locale={zhCN} theme={themeConfig}>
      <AntdApp>
        <ErrorBoundary>
          <RouterProvider router={router} />
        </ErrorBoundary>
      </AntdApp>
    </ConfigProvider>
  )
}
