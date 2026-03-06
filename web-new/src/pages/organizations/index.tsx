import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Tree, Button, Space, Tag, Modal, Descriptions, message, Empty } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, TeamOutlined, SearchOutlined, FileTextOutlined, ApartmentOutlined } from '@ant-design/icons';
import type { Organization } from '@/types';
import { http } from '@/utils/request';
import { usePermission } from '@/utils/permission';

interface TreeNode {
  key: string;
  title: React.ReactNode;
  children?: TreeNode[];
  data: Organization;
}

export default function OrganizationsPage() {
  const navigate = useNavigate();
  const { isAdmin } = usePermission();
  const [treeData, setTreeData] = useState<TreeNode[]>([]);
  const [, setLoading] = useState(false);
  const [selectedOrg, setSelectedOrg] = useState<Organization | null>(null);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res: any = await http.get('/organizations');
      const tree = buildTree(res.list || []);
      setTreeData(tree);
      if (tree.length > 0 && !selectedOrg) {
        setSelectedOrg(tree[0].data);
      }
    } catch (error) {
      console.error('获取组织列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const buildTree = (orgs: Organization[]): TreeNode[] => {
    const nodeMap = new Map<string, TreeNode>();
    
    orgs.forEach((org) => {
      nodeMap.set(org.id, {
        key: org.id,
        title: (
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '4px 0' }}>
            <ApartmentOutlined style={{ color: '#e67e22' }} />
            <span style={{ fontWeight: 500 }}>{org.name}</span>
            <Tag style={{ fontSize: 12 }}>{getTypeLabel(org.type)}</Tag>
          </div>
        ),
        data: org,
        children: [],
      });
    });

    const roots: TreeNode[] = [];
    orgs.forEach((org) => {
      const node = nodeMap.get(org.id)!;
      if (org.parent_id && nodeMap.has(org.parent_id)) {
        const parent = nodeMap.get(org.parent_id)!;
        if (!parent.children) parent.children = [];
        parent.children.push(node);
      } else {
        roots.push(node);
      }
    });

    return roots;
  };

  const handleDelete = (org: Organization) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除组织 "${org.name}" 吗？此操作不可恢复。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await http.delete(`/organizations/${org.id}`);
          message.success('删除成功');
          fetchData();
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  };

  return (
    <Card
      title="组织架构管理"
      extra={
        isAdmin && (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/organizations/create')}
            style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}
          >
            添加组织
          </Button>
        )
      }
      bordered={false}
      bodyStyle={{ padding: 0, height: 'calc(100vh - 180px)' }}
    >
      <div style={{ display: 'flex', height: '100%' }}>
        {/* 组织树 */}
        <div 
          style={{ 
            width: 300, 
            borderRight: '1px solid #f0f0f0', 
            padding: '16px 0',
            overflow: 'auto',
            background: '#fafafa',
          }}
        >
          <Tree
            treeData={treeData}
            onSelect={(_, { node }) => setSelectedOrg((node as any).data)}
            defaultExpandAll
            blockNode
            style={{ background: 'transparent' }}
          />
        </div>

        {/* 组织详情 */}
        <div style={{ flex: 1, padding: 24, overflow: 'auto' }}>
          {selectedOrg ? (
            <>
              <div 
                style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  justifyContent: 'space-between', 
                  marginBottom: 24,
                  paddingBottom: 16,
                  borderBottom: '1px solid #f0f0f0',
                }}
              >
                <div>
                  <h2 style={{ margin: '0 0 8px 0', fontSize: 20, fontWeight: 600, color: '#1f2329' }}>
                    {selectedOrg.name}
                  </h2>
                  <Tag color="orange" style={{ fontSize: 13 }}>
                    {getTypeLabel(selectedOrg.type)}
                  </Tag>
                </div>
                {isAdmin && (
                  <Space>
                    <Button
                      icon={<EditOutlined />}
                      onClick={() => navigate(`/organizations/${selectedOrg.id}/edit`)}
                    >
                      编辑
                    </Button>
                    <Button
                      danger
                      icon={<DeleteOutlined />}
                      onClick={() => handleDelete(selectedOrg)}
                    >
                      删除
                    </Button>
                  </Space>
                )}
              </div>

              <Descriptions 
                bordered 
                column={2}
                labelStyle={{ backgroundColor: '#fafafa', fontWeight: 500, width: 120 }}
                contentStyle={{ backgroundColor: '#fff' }}
              >
                <Descriptions.Item label="组织代码">
                  {selectedOrg.code}
                </Descriptions.Item>
                <Descriptions.Item label="组织类型">
                  {getTypeLabel(selectedOrg.type)}
                </Descriptions.Item>
                <Descriptions.Item label="所在地区" span={2}>
                  {[selectedOrg.province, selectedOrg.city, selectedOrg.district]
                    .filter(Boolean)
                    .join(' / ') || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="层级">
                  第 {selectedOrg.level} 级
                </Descriptions.Item>
                <Descriptions.Item label="状态">
                  <Tag color={selectedOrg.status === 'active' ? 'success' : 'error'}>
                    {selectedOrg.status === 'active' ? '正常' : '禁用'}
                  </Tag>
                </Descriptions.Item>
              </Descriptions>

              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16, marginTop: 24 }}>
                <Card 
                  bordered={true}
                  bodyStyle={{ padding: 20 }}
                  style={{ borderRadius: 8 }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                    <div 
                      style={{ 
                        width: 48, 
                        height: 48, 
                        borderRadius: 8, 
                        background: '#fdf2e9',
                        display: 'flex', 
                        alignItems: 'center', 
                        justifyContent: 'center' 
                      }}
                    >
                      <TeamOutlined style={{ fontSize: 24, color: '#e67e22' }} />
                    </div>
                    <div>
                      <div style={{ color: '#646a73', fontSize: 13, marginBottom: 4 }}>志愿者数</div>
                      <div style={{ fontSize: 28, fontWeight: 700, color: '#1f2329' }}>
                        {selectedOrg.volunteer_count || 0}
                      </div>
                    </div>
                  </div>
                </Card>
                <Card 
                  bordered={true}
                  bodyStyle={{ padding: 20 }}
                  style={{ borderRadius: 8 }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                    <div 
                      style={{ 
                        width: 48, 
                        height: 48, 
                        borderRadius: 8, 
                        background: '#ebf5fb',
                        display: 'flex', 
                        alignItems: 'center', 
                        justifyContent: 'center' 
                      }}
                    >
                      <SearchOutlined style={{ fontSize: 24, color: '#3498db' }} />
                    </div>
                    <div>
                      <div style={{ color: '#646a73', fontSize: 13, marginBottom: 4 }}>案件数</div>
                      <div style={{ fontSize: 28, fontWeight: 700, color: '#1f2329' }}>
                        {selectedOrg.case_count || 0}
                      </div>
                    </div>
                  </div>
                </Card>
                <Card 
                  bordered={true}
                  bodyStyle={{ padding: 20 }}
                  style={{ borderRadius: 8 }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                    <div 
                      style={{ 
                        width: 48, 
                        height: 48, 
                        borderRadius: 8, 
                        background: '#eafaf1',
                        display: 'flex', 
                        alignItems: 'center', 
                        justifyContent: 'center' 
                      }}
                    >
                      <FileTextOutlined style={{ fontSize: 24, color: '#27ae60' }} />
                    </div>
                    <div>
                      <div style={{ color: '#646a73', fontSize: 13, marginBottom: 4 }}>任务数</div>
                      <div style={{ fontSize: 28, fontWeight: 700, color: '#1f2329' }}>
                        0
                      </div>
                    </div>
                  </div>
                </Card>
              </div>
            </>
          ) : (
            <div style={{ 
              display: 'flex', 
              alignItems: 'center', 
              justifyContent: 'center', 
              height: '100%', 
              color: '#8f959e' 
            }}>
              <Empty description="请选择组织查看详情" />
            </div>
          )}
        </div>
      </div>
    </Card>
  );
}

function getTypeLabel(type: string) {
  const labels: Record<string, string> = {
    root: '总部',
    province: '省级',
    city: '市级',
    district: '区级',
    street: '街道',
  };
  return labels[type] || type;
}
