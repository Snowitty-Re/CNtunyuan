import { useState } from 'react';
import { Upload, message } from 'antd';
import { InboxOutlined } from '@ant-design/icons';
import { uploadFile } from '@/api/upload';
import type { FileResponse } from '@/types';

interface Props {
  accept?: string;
  maxSize?: number;
  onSuccess?: (file: FileResponse) => void;
}

export default function FileUpload({ accept, maxSize = 10, onSuccess }: Props) {
  const [uploading, setUploading] = useState(false);

  const customUpload = async (options: any) => {
    const { file, onProgress, onSuccess: onUploadSuccess, onError } = options;

    if (maxSize && file.size > maxSize * 1024 * 1024) {
      message.error(`文件大小不能超过 ${maxSize}MB`);
      onError(new Error('文件过大'));
      return;
    }

    setUploading(true);
    try {
      const res = await uploadFile(file, (percent) => onProgress({ percent }));
      onUploadSuccess(res);
      onSuccess?.(res);
      message.success('上传成功');
    } catch (err) {
      onError(err);
    } finally {
      setUploading(false);
    }
  };

  return (
    <Upload.Dragger
      customRequest={customUpload}
      accept={accept}
      showUploadList={true}
      disabled={uploading}
    >
      <p className="ant-upload-drag-icon">
        <InboxOutlined />
      </p>
      <p className="ant-upload-text">点击或拖拽文件上传</p>
      <p className="ant-upload-hint">最大 {maxSize}MB</p>
    </Upload.Dragger>
  );
}
