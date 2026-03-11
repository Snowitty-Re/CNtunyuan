import { useEffect, useState } from 'react';
import { TreeSelect } from 'antd';
import { getOrganizationTree } from '@/api/organization';
import type { OrganizationTreeResponse } from '@/types';

interface Props {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  allowClear?: boolean;
  style?: React.CSSProperties;
}

function toTreeData(nodes: OrganizationTreeResponse[]): any[] {
  return nodes.map((n) => ({
    title: n.name,
    value: n.id,
    children: n.children ? toTreeData(n.children) : [],
  }));
}

export default function OrgTreeSelect({ value, onChange, placeholder = '选择组织', allowClear = true, style }: Props) {
  const [treeData, setTreeData] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    getOrganizationTree()
      .then((data) => setTreeData(toTreeData(data || [])))
      .finally(() => setLoading(false));
  }, []);

  return (
    <TreeSelect
      value={value}
      onChange={onChange}
      treeData={treeData}
      placeholder={placeholder}
      allowClear={allowClear}
      loading={loading}
      showSearch
      treeNodeFilterProp="title"
      style={style}
    />
  );
}
