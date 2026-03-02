-- ============================================
-- 团圆寻亲系统 - 创建数据库脚本
-- 编码: UTF-8
-- 数据库: PostgreSQL 14+
-- ============================================
-- 
-- 使用说明：
-- 此脚本需要在 postgres 数据库中执行
-- 
-- 执行方式：
-- psql -U postgres -f create_database.sql
-- 
-- 或者手动执行：
-- CREATE DATABASE cntunyuan WITH ENCODING = 'UTF8';
--
-- ============================================

-- 创建数据库（使用 UTF-8 编码）
-- 注意：如果数据库已存在，会报错，可以忽略
CREATE DATABASE cntunyuan 
    WITH 
    ENCODING = 'UTF8'
    LC_COLLATE = 'C'
    LC_CTYPE = 'C'
    TEMPLATE = template0;

-- 完成提示
SELECT 'Database cntunyuan created successfully with UTF-8 encoding!' AS status;
