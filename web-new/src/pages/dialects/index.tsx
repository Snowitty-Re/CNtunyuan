import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Button, Input, Tag, Space, Avatar, Dropdown, Modal, message } from 'antd';
import { PlusOutlined, MoreOutlined, EyeOutlined, EditOutlined, DeleteOutlined, PlayCircleOutlined, SoundOutlined } from '@ant-design/icons';
import type { Dialect } from '@/types';
import { http } from '@/utils/request';
import { usePermission } from '@/utils/permission';

export default function DialectsPage() {
  const navigate = useNavigate();
  const { isManager } = usePermission();
  const [data, setData] = useState<Dialect[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  useEffect(() => {
    fetchData();
  }, [pagination.current, pagination.pageSize]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res: any = await http.get('/dialects', {
        params: {
          page: pagination.current,
          page_size: pagination.pageSize,
          keyword: searchText || undefined,
        },
      });
      setData(res.list || []);
      setPagination((prev) => ({ ...prev, total: res.total || 0 }));
    } catch (error) {
      console.error('获取方言列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    setPagination((prev) => ({ ...prev, current: 1 }));
    fetchData();
  };

  const handleDelete = (record: Dialect) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除方言样本 "${record.title}" 吗？`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await http.delete(`/dialects/${record.id}`);
          message.success('删除成功');
          fetchData();
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  };

  const formatDuration = (seconds?: number) => {
    if (!seconds) return '-';
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const columns = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (text: string, record: Dialect) => (
        <Space>
          <Avatar
            size={40}
            icon={<SoundOutlined />}
            style={{ backgroundColor: '#8e44ad' }}
          />
          <div>
            <div style={{ fontWeight: 500, color: '#1f2329' }}>{text}</div>
            {record.description && (
              <div style={{ 
                color: '#8f959e', 
                fontSize: 13, 
                maxWidth: 300, 
                overflow: 'hidden', 
                textOverflow: 'ellipsis', 
                whiteSpace: 'nowrap' 
              }}>
                {record.description}
              </div>
            )}
          </div>
        </Space>
      ),
    },
    {
      title: '地区',
      dataIndex: 'province',
      key: 'location',
      render: (_: any, record: Dialect) => (
        <Tag 
          color="purple" 
          style={{ fontSize: 13, background: '#f5eef8', border: 'none' }}
        >
          {[record.province, record.city, record.district]
            .filter(Boolean)
            .join(' / ') || '未知'}
        </Tag>
      ),
    },
    {
      title: '时长',
      dataIndex: 'duration',
      key: 'duration',
      width: 90,
      align: 'center' as const,
      render: (duration?: number) => (
        <span style={{ color: '#646a73', fontFamily: 'monospace' }}>
          {formatDuration(duration)}
        </span>
      ),
    },
    {
      title: '播放次数',
      dataIndex: 'play_count',
      key: 'play_count',
      width: 100,
      align: 'center' as const,
      render: (count: number) => <span style={{ color: '#646a73' }}>{count || 0}</span>,
    },
    {
      title: '采集人',
      dataIndex: ['collector', 'nickname'],
      key: 'collector',
      width: 120,
      render: (name: string) => <span style={{ color: '#646a73' }}>{name || '未知'}</span>,
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      fixed: 'right' as const,
      render: (_: any, record: Dialect) => (
        <Dropdown
          menu={{
            items: [
              {
                key: 'play',
                icon: <PlayCircleOutlined />,
                label: '播放',
                onClick: () => window.open(record.audio_url, '_blank'),
              },
              {
                key: 'view',
                icon: <EyeOutlined />,
                label: '查看详情',
                onClick: () => navigate(`/dialects/${record.id}`),
              },
              ...(isManager
                ? [
                    {
                      key: 'edit',
                      icon: <EditOutlined />,
                      label: '编辑',
                      onClick: () => navigate(`/dialects/${record.id}/edit`),
                    },
                    {
                      key: 'delete',
                      icon: <DeleteOutlined />,
                      label: <span style={{ color: '#f5222d' }}>删除</span>,
                      onClick: () => handleDelete(record),
                    },
                  ]
                : []),
            ],
          }}
        >
          <Button type="text" icon={<MoreOutlined style={{ fontSize: 18 }} />} style={{ color: '#646a73' }} />
        </Dropdown>
      ),
    },
  ];

  return (
    <Card
      title="方言样本库"
      extra={
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate('/dialects/create')}
          style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}
        >
          录制方言
        </Button>
      }
      bordered={false}
    >
      <div style={{ marginBottom: 16 }}>
        <Input.Search
          placeholder="搜索标题、地区"
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          onSearch={handleSearch}
          style={{ width: 300 }}
          allowClear
        />
      </div>

      <Table
        columns={columns}
        dataSource={data}
        loading={loading}
        rowKey="id"
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        onChange={(p) =>
          setPagination({
            current: p.current || 1,
            pageSize: p.pageSize || 10,
            total: pagination.total,
          })
        }
      />
    </Card>
  );
}
