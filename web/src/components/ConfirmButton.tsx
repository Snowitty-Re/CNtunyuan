import { Button, Popconfirm } from 'antd';
import type { ButtonProps } from 'antd';

interface Props extends ButtonProps {
  title?: string;
  description?: string;
  onConfirm: () => void;
}

export default function ConfirmButton({
  title = '确定要执行此操作吗？',
  description,
  onConfirm,
  children,
  ...rest
}: Props) {
  return (
    <Popconfirm title={title} description={description} onConfirm={onConfirm} okText="确定" cancelText="取消">
      <Button {...rest}>{children}</Button>
    </Popconfirm>
  );
}
