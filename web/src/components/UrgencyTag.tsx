import { Tag } from 'antd';
import { URGENCY_MAP } from '@/utils/constants';

const URGENCY_COLORS: Record<string, string> = {
  urgent: 'red',
  high: 'orange',
  normal: 'blue',
  low: 'green',
};

export default function UrgencyTag({ level }: { level: string }) {
  return (
    <Tag color={URGENCY_COLORS[level] || 'default'}>
      {URGENCY_MAP[level] || level}
    </Tag>
  );
}
