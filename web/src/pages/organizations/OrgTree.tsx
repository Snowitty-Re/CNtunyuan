import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Tree, Spin, Button, Typography, Empty } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import { getOrganizationTree } from '@/api/organization';
import type { OrganizationTreeResponse } from '@/types';
import { ORG_TYPE_MAP } from '@/utils/constants';

function toTreeData(nodes: OrganizationTreeResponse[]): any[] {
  return nodes.map((n) => ({
    key: n.id,
    title: `${n.name} (${ORG_TYPE_MAP[n.type] || n.type})`,
    children: n.children ? toTreeData(n.children) : [],
  }));
}

export default function OrgTreePage() {
  const navigate = useNavigate();
  const [treeData, setTreeData] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getOrganizationTree()
      .then((data) => setTreeData(toTreeData(data || [])))
      .finally(() => setLoading(false));
  }, []);

  return (
    <PageContainer
      title="组织树视图"
      breadcrumb={[{ title: '首页' }, { title: '组织管理' }, { title: '树形视图' }]}
      extra={<Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>返回列表</Button>}
    >
      <Card>
        {loading ? (
          <div style={{ textAlign: 'center', padding: 60 }}><Spin size="large" /></div>
        ) : treeData.length > 0 ? (
          <Tree treeData={treeData} defaultExpandAll showLine blockNode />
        ) : (
          <Empty description="暂无组织数据" />
        )}
      </Card>
    </PageContainer>
  );
}
