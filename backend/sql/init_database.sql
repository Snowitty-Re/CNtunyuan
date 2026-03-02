-- ============================================
-- 团圆寻亲系统 - 数据库初始化脚本
-- 编码: UTF-8
-- 数据库: PostgreSQL 14+
-- ============================================

-- 创建数据库（如果不存在）
-- 注意：需要在 postgres 数据库中执行
-- CREATE DATABASE cntunyuan WITH ENCODING = 'UTF8' LC_COLLATE = 'zh_CN.UTF-8' LC_CTYPE = 'zh_CN.UTF-8';

-- 连接到数据库
\c cntunyuan;

-- ============================================
-- 扩展
-- ============================================
-- 启用 UUID 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 启用全文搜索扩展（中文）
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ============================================
-- 字符集设置
-- ============================================
-- 确保数据库使用 UTF-8 编码
UPDATE pg_database SET encoding = pg_char_to_encoding('UTF8') WHERE datname = 'cntunyuan';

-- 设置客户端编码
SET client_encoding = 'UTF8';

-- ============================================
-- 表空间（可选）
-- ============================================
-- 如果有大量图片/文件存储，可以考虑单独表空间
-- CREATE TABLESPACE ts_cntunyuan_data LOCATION '/var/lib/postgresql/data/ts';

-- ============================================
-- 完成
-- ============================================
SELECT 'Database initialization completed successfully!' AS status;
