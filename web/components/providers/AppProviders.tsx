'use client'

import { AntdRegistry } from '@ant-design/nextjs-registry'
import { App, ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { PropsWithChildren } from 'react'
import { antdTheme } from '@/theme/antd'
import { AuthProvider } from '@/components/providers/AuthProvider'

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <AntdRegistry>
      <ConfigProvider locale={zhCN} theme={antdTheme}>
        <App>
          <AuthProvider>{children}</AuthProvider>
        </App>
      </ConfigProvider>
    </AntdRegistry>
  )
}
