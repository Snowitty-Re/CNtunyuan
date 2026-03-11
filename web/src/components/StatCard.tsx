import { Card, Statistic } from 'antd';
import type { ReactNode } from 'react';

interface Props {
  title: string;
  value: number;
  prefix?: ReactNode;
  suffix?: string;
  color?: string;
  loading?: boolean;
}

export default function StatCard({ title, value, prefix, suffix, color, loading }: Props) {
  return (
    <Card
      hoverable
      style={{ transition: 'all 0.2s' }}
      bodyStyle={{ padding: '20px 24px' }}
    >
      <Statistic
        title={title}
        value={value}
        prefix={prefix}
        suffix={suffix}
        loading={loading}
        valueStyle={{ color: color || '#1677ff' }}
      />
    </Card>
  );
}
