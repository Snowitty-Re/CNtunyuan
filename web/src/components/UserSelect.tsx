import { useState, useCallback } from 'react';
import { Select } from 'antd';
import { getUsers } from '@/api/user';
import type { UserResponse } from '@/types';

interface Props {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  allowClear?: boolean;
  style?: React.CSSProperties;
  role?: string;
}

export default function UserSelect({ value, onChange, placeholder = '选择用户', allowClear = true, style, role }: Props) {
  const [options, setOptions] = useState<UserResponse[]>([]);
  const [loading, setLoading] = useState(false);

  const handleSearch = useCallback(
    async (keyword: string) => {
      if (!keyword || keyword.length < 1) return;
      setLoading(true);
      try {
        const res = await getUsers({ keyword, page: 1, page_size: 20, role });
        setOptions(res.list || []);
      } finally {
        setLoading(false);
      }
    },
    [role]
  );

  return (
    <Select
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      allowClear={allowClear}
      showSearch
      filterOption={false}
      onSearch={handleSearch}
      loading={loading}
      style={style}
      options={options.map((u) => ({
        label: `${u.nickname} (${u.phone})`,
        value: u.id,
      }))}
    />
  );
}
