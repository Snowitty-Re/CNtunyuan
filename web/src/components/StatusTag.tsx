import { Tag } from 'antd';
import {
  USER_STATUS_MAP,
  MISSING_PERSON_STATUS_MAP,
  TASK_STATUS_MAP,
  DIALECT_STATUS_MAP,
  AUDIT_STATUS_MAP,
} from '@/utils/constants';

const STATUS_COLORS: Record<string, string> = {
  active: 'green',
  inactive: 'default',
  banned: 'red',
  missing: 'red',
  searching: 'orange',
  found: 'blue',
  reunited: 'green',
  closed: 'default',
  draft: 'default',
  pending: 'gold',
  assigned: 'blue',
  processing: 'cyan',
  completed: 'green',
  cancelled: 'default',
  rejected: 'red',
  success: 'green',
  failure: 'red',
};

const ALL_STATUS_MAP: Record<string, string> = {
  ...USER_STATUS_MAP,
  ...MISSING_PERSON_STATUS_MAP,
  ...TASK_STATUS_MAP,
  ...DIALECT_STATUS_MAP,
  ...AUDIT_STATUS_MAP,
};

interface Props {
  status: string;
  map?: Record<string, string>;
}

export default function StatusTag({ status, map }: Props) {
  const label = map?.[status] || ALL_STATUS_MAP[status] || status;
  return <Tag color={STATUS_COLORS[status] || 'default'}>{label}</Tag>;
}
