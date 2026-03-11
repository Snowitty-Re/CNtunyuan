import { Button, Result } from 'antd';
import { useNavigate } from 'react-router-dom';

export default function Error500() {
  const navigate = useNavigate();
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh' }}>
      <Result
        status="500"
        title="500"
        subTitle="抱歉，服务器出了点问题"
        extra={<Button type="primary" onClick={() => navigate('/dashboard')}>返回首页</Button>}
      />
    </div>
  );
}
