import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Input, Select, Button, Space, Row, Col, Image, message } from 'antd';
import { SearchOutlined, ReloadOutlined, BarChartOutlined, DownloadOutlined, DeleteOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import ConfirmButton from '@/components/ConfirmButton';
import FileUpload from '@/components/FileUpload';
import { getFiles, deleteFile } from '@/api/upload';
import { usePagination } from '@/hooks/usePagination';
import { formatDateTime, formatFileSize } from '@/utils/format';
import { FILE_TYPE_MAP } from '@/utils/constants';
import type { FileResponse } from '@/types';

export default function FilesPage() {
  const navigate = useNavigate();
  const { page, pageSize, pagination, setTotal, onChange } = usePagination();
  const [data, setData] = useState<FileResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [fileType, setFileType] = useState('');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getFiles({ page, page_size: pageSize, keyword, file_type: fileType || undefined });
      setData(res.list || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, keyword, fileType, setTotal]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleDelete = async (id: string) => {
    await deleteFile(id);
    message.success('删除成功');
    fetchData();
  };

  const columns = [
    {
      title: '预览', key: 'preview', width: 60,
      render: (_: unknown, r: FileResponse) =>
        r.file_type === 'image' ? <Image src={r.url} width={40} height={40} style={{ objectFit: 'cover' }} /> : null,
    },
    { title: '文件名', dataIndex: 'original_name', key: 'original_name', ellipsis: true },
    {
      title: '类型', dataIndex: 'file_type', key: 'file_type', width: 80,
      render: (v: string) => FILE_TYPE_MAP[v] || v,
    },
    {
      title: '大小', dataIndex: 'size', key: 'size', width: 100,
      render: (v: number) => formatFileSize(v),
    },
    {
      title: '上传时间', dataIndex: 'created_at', key: 'created_at', width: 160,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: '操作', key: 'actions', width: 120,
      render: (_: unknown, r: FileResponse) => (
        <Space>
          <Button type="link" size="small" icon={<DownloadOutlined />} href={r.url} target="_blank">下载</Button>
          <ConfirmButton type="link" size="small" danger title="确定删除?" onConfirm={() => handleDelete(r.id)}>删除</ConfirmButton>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer
      title="文件管理"
      breadcrumb={[{ title: '首页' }, { title: '文件管理' }]}
      extra={
        <Button icon={<BarChartOutlined />} onClick={() => navigate('/files/stats')}>存储统计</Button>
      }
    >
      <Card style={{ marginBottom: 16 }}>
        <FileUpload maxSize={20} onSuccess={() => fetchData()} />
      </Card>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8}><Input placeholder="搜索文件名" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => onChange(1, pageSize)} /></Col>
          <Col xs={12} sm={4}>
            <Select placeholder="类型" allowClear style={{ width: '100%' }} value={fileType || undefined} onChange={(v) => setFileType(v || '')}
              options={Object.entries(FILE_TYPE_MAP).map(([k, v]) => ({ label: v, value: k }))} />
          </Col>
          <Col><Space>
            <Button type="primary" icon={<SearchOutlined />} onClick={() => onChange(1, pageSize)}>搜索</Button>
            <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setFileType(''); onChange(1, pageSize); }}>重置</Button>
          </Space></Col>
        </Row>
      </Card>
      <Card><Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={pagination} scroll={{ x: 700 }} /></Card>
    </PageContainer>
  );
}
