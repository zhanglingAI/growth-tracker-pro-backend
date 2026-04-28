-- Growth Tracker Pro 数据库初始化脚本
-- 版本: v1.0
-- 日期: 2026-04-28

-- 创建数据库
CREATE DATABASE IF NOT EXISTS growth_tracker
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE growth_tracker;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(36) PRIMARY KEY,
  open_id VARCHAR(128) NOT NULL UNIQUE COMMENT '微信OpenID',
  nick_name VARCHAR(64) COMMENT '昵称',
  avatar_url VARCHAR(512) COMMENT '头像URL',
  phone VARCHAR(20) COMMENT '手机号',
  settings TEXT COMMENT '用户设置JSON',
  family_id VARCHAR(36) COMMENT '家庭ID',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_open_id (open_id),
  INDEX idx_family_id (family_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 宝宝表
CREATE TABLE IF NOT EXISTS children (
  id VARCHAR(36) PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL COMMENT '所属用户ID',
  name VARCHAR(64) NOT NULL COMMENT '宝宝姓名',
  gender VARCHAR(10) NOT NULL COMMENT '性别 male/female',
  birthday DATE NOT NULL COMMENT '出生日期',
  father_height DECIMAL(5,1) COMMENT '父亲身高cm',
  mother_height DECIMAL(5,1) COMMENT '母亲身高cm',
  is_active BOOLEAN DEFAULT TRUE COMMENT '是否当前选中',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='宝宝表';

-- 生长记录表
CREATE TABLE IF NOT EXISTS growth_records (
  id VARCHAR(36) PRIMARY KEY,
  child_id VARCHAR(36) NOT NULL COMMENT '关联宝宝ID',
  height DECIMAL(5,1) NOT NULL COMMENT '身高cm',
  weight DECIMAL(5,1) NOT NULL COMMENT '体重kg',
  date DATE NOT NULL COMMENT '测量日期',
  age_str VARCHAR(20) COMMENT '年龄字符串',
  age_in_days INT COMMENT '年龄天数',
  note TEXT COMMENT '备注',
  photo VARCHAR(512) COMMENT '照片URL',
  creator_id VARCHAR(36) COMMENT '创建者ID',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_child_id (child_id),
  INDEX idx_date (date),
  INDEX idx_child_date (child_id, date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生长记录表';

-- 订阅表
CREATE TABLE IF NOT EXISTS subscriptions (
  id VARCHAR(36) PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL UNIQUE COMMENT '用户ID',
  plan VARCHAR(20) NOT NULL COMMENT '订阅方案 monthly/quarterly/yearly',
  start_date DATE COMMENT '开始时间',
  end_date DATE COMMENT '到期时间',
  ai_quota INT DEFAULT 0 COMMENT 'AI额度',
  ai_used INT DEFAULT 0 COMMENT '已使用次数',
  referred_by VARCHAR(36) COMMENT '推荐人ID',
  referral_code VARCHAR(20) UNIQUE COMMENT '邀请码',
  subscription_cancelled BOOLEAN DEFAULT FALSE COMMENT '是否取消续费',
  status VARCHAR(20) DEFAULT 'active' COMMENT '状态 active/expired/cancelled',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订阅表';

-- 家庭表
CREATE TABLE IF NOT EXISTS families (
  id VARCHAR(36) PRIMARY KEY,
  family_id VARCHAR(36) NOT NULL UNIQUE COMMENT '家庭唯一标识',
  creator_id VARCHAR(36) COMMENT '创建者用户ID',
  name VARCHAR(64) COMMENT '家庭名称',
  invite_code VARCHAR(20) UNIQUE COMMENT '邀请码',
  max_members INT DEFAULT 10 COMMENT '最大成员数',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_family_id (family_id),
  INDEX idx_invite_code (invite_code),
  INDEX idx_creator_id (creator_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家庭表';

-- 家庭成员表
CREATE TABLE IF NOT EXISTS family_members (
  id VARCHAR(36) PRIMARY KEY,
  family_id VARCHAR(36) NOT NULL COMMENT '家庭ID',
  user_id VARCHAR(36) NOT NULL COMMENT '用户ID',
  name VARCHAR(64) COMMENT '姓名',
  phone VARCHAR(20) COMMENT '手机号',
  role VARCHAR(20) DEFAULT 'viewer' COMMENT '角色 owner/editor/viewer',
  joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_family_id (family_id),
  INDEX idx_user_id (user_id),
  UNIQUE KEY uk_family_user (family_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家庭成员表';

-- 家庭宝宝关联表
CREATE TABLE IF NOT EXISTS family_children (
  id VARCHAR(36) PRIMARY KEY,
  family_id VARCHAR(36) NOT NULL COMMENT '家庭ID',
  child_id VARCHAR(36) NOT NULL COMMENT '宝宝ID',
  added_by VARCHAR(36) COMMENT '添加者ID',
  added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_family_id (family_id),
  INDEX idx_child_id (child_id),
  UNIQUE KEY uk_family_child (family_id, child_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家庭宝宝关联表';

-- 化验单报告表
CREATE TABLE IF NOT EXISTS lab_reports (
  id VARCHAR(36) PRIMARY KEY,
  child_id VARCHAR(36) COMMENT '关联宝宝ID',
  user_id VARCHAR(36) COMMENT '上传用户',
  image_url VARCHAR(512) NOT NULL COMMENT '图片URL',
  ocr_text TEXT COMMENT 'OCR识别文本',
  ai_result TEXT COMMENT 'AI解析结果JSON',
  report_type VARCHAR(50) COMMENT '报告类型',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_child_id (child_id),
  INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='化验单报告表';

-- AI对话表
CREATE TABLE IF NOT EXISTS ai_conversations (
  id VARCHAR(36) PRIMARY KEY,
  user_id VARCHAR(36) COMMENT '用户ID',
  child_id VARCHAR(36) COMMENT '宝宝ID',
  session_id VARCHAR(36) COMMENT '会话ID',
  messages TEXT COMMENT '消息列表JSON',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_session_id (session_id),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI对话表';

-- 会员方案定价表
CREATE TABLE IF NOT EXISTS subscription_plans (
  id INT PRIMARY KEY AUTO_INCREMENT,
  plan_id VARCHAR(20) NOT NULL UNIQUE COMMENT '方案ID',
  name VARCHAR(50) NOT NULL COMMENT '方案名称',
  price INT NOT NULL COMMENT '价格(分)',
  ai_quota INT NOT NULL COMMENT 'AI额度',
  max_members INT NOT NULL COMMENT '最大成员数',
  duration_days INT NOT NULL COMMENT '时长(天)',
  sort_order INT DEFAULT 0 COMMENT '排序',
  is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订阅方案表';

-- 插入默认订阅方案
INSERT INTO subscription_plans (plan_id, name, price, ai_quota, max_members, duration_days, sort_order) VALUES
('monthly', '月卡', 2990, 30, 2, 30, 1),
('quarterly', '季卡', 6990, 100, 5, 90, 2),
('yearly', '年卡', 19990, 400, 10, 365, 3)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id VARCHAR(36) COMMENT '操作用户ID',
  action VARCHAR(50) NOT NULL COMMENT '操作类型',
  target_type VARCHAR(50) COMMENT '目标类型',
  target_id VARCHAR(36) COMMENT '目标ID',
  detail TEXT COMMENT '详情JSON',
  ip_address VARCHAR(50) COMMENT 'IP地址',
  user_agent VARCHAR(512) COMMENT 'UserAgent',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_action (action),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
