import { Tag } from 'antd';
import { USER_ROLE_MAP } from '@/utils/constants';

const ROLE_COLORS: Record<string, string> = {
  super_admin: 'red',
  admin: 'orange',
  manager: 'blue',
  volunteer: 'green',
};

export default function RoleTag({ role }: { role: string }) {
  return (
    <Tag color={ROLE_COLORS[role] || 'default'}>
      {USER_ROLE_MAP[role] || role}
    </Tag>
  );
}
