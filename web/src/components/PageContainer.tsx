import React from 'react';
import { Breadcrumb, Typography } from 'antd';
import { motion } from 'framer-motion';

interface Props {
  title: string;
  breadcrumb?: { title: string; path?: string }[];
  extra?: React.ReactNode;
  children: React.ReactNode;
}

export default function PageContainer({ title, breadcrumb, extra, children }: Props) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.2 }}
    >
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          {breadcrumb && (
            <Breadcrumb
              style={{ marginBottom: 8 }}
              items={breadcrumb.map((b) => ({ title: b.title }))}
            />
          )}
          <Typography.Title level={4} style={{ margin: 0 }}>
            {title}
          </Typography.Title>
        </div>
        {extra && <div>{extra}</div>}
      </div>
      {children}
    </motion.div>
  );
}
