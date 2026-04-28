-- Growth Tracker Pro 数据库初始化脚本
-- 版本: v2.0
-- 日期: 2026-04-28
-- 基于 PRD v2.0 补全所有功能模块

-- 创建数据库
CREATE DATABASE IF NOT EXISTS growth_tracker
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE growth_tracker;

-- =====================================================
-- P0 核心功能表
-- =====================================================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  openid VARCHAR(64) NOT NULL UNIQUE COMMENT '微信OpenID',
  nickname VARCHAR(50) COMMENT '昵称',
  avatar VARCHAR(500) COMMENT '头像URL',
  phone VARCHAR(20) COMMENT '手机号',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_openid (openid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 宝宝表
CREATE TABLE IF NOT EXISTS children (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id INT UNSIGNED NOT NULL COMMENT '所属用户ID',
  family_id INT UNSIGNED COMMENT '所属家庭组ID',
  nickname VARCHAR(50) NOT NULL COMMENT '宝宝昵称',
  gender VARCHAR(10) NOT NULL COMMENT '性别 male/female',
  birthday DATE NOT NULL COMMENT '出生日期',
  initial_height DECIMAL(5,1) COMMENT '初始身高cm',
  initial_weight DECIMAL(5,1) COMMENT '初始体重kg',
  father_height DECIMAL(5,1) COMMENT '父亲身高cm',
  mother_height DECIMAL(5,1) COMMENT '母亲身高cm',
  standard_type VARCHAR(10) DEFAULT 'cn' COMMENT 'cn/who',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_family_id (family_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='宝宝表';

-- 生长记录表
CREATE TABLE IF NOT EXISTS growth_records (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  child_id INT UNSIGNED NOT NULL COMMENT '关联宝宝ID',
  measure_date DATE NOT NULL COMMENT '测量日期',
  height DECIMAL(5,1) NOT NULL COMMENT '身高cm',
  weight DECIMAL(5,1) COMMENT '体重kg',
  height_percentile DECIMAL(5,2) COMMENT '身高百分位',
  weight_percentile DECIMAL(5,2) COMMENT '体重百分位',
  height_zscore DECIMAL(5,3) COMMENT '身高Z分数',
  weight_zscore DECIMAL(5,3) COMMENT '体重Z分数',
  height_status VARCHAR(20) DEFAULT 'normal' COMMENT 'normal/low/high/very_low/very_high',
  weight_status VARCHAR(20) DEFAULT 'normal' COMMENT 'normal/low/high/very_low/very_high',
  remarks VARCHAR(500) COMMENT '备注',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_child_id (child_id),
  INDEX idx_measure_date (measure_date),
  INDEX idx_child_date (child_id, measure_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生长记录表';

-- =====================================================
-- P1 MVP功能表
-- =====================================================

-- 家庭组表
CREATE TABLE IF NOT EXISTS families (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(50) NOT NULL COMMENT '家庭名称',
  invite_code VARCHAR(6) UNIQUE COMMENT '6位邀请码',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_invite_code (invite_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家庭组表';

-- 家庭成员表
CREATE TABLE IF NOT EXISTS family_members (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  family_id INT UNSIGNED NOT NULL COMMENT '家庭组ID',
  user_id INT UNSIGNED NOT NULL COMMENT '用户ID',
  role VARCHAR(20) DEFAULT 'member' COMMENT 'creator/member/guest',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_family_user (family_id, user_id),
  INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家庭成员表';

-- 医院表
CREATE TABLE IF NOT EXISTS hospitals (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(200) NOT NULL COMMENT '医院名称',
  level VARCHAR(20) COMMENT '三级甲等/二级甲等',
  address VARCHAR(500) COMMENT '地址',
  latitude DECIMAL(10,6) COMMENT '纬度',
  longitude DECIMAL(10,6) COMMENT '经度',
  phone VARCHAR(20) COMMENT '电话',
  logo VARCHAR(500) COMMENT '医院Logo',
  pediatric_endo BOOLEAN DEFAULT TRUE COMMENT '有儿童内分泌科',
  estimated_fee VARCHAR(50) COMMENT '预估费用',
  city VARCHAR(50) COMMENT '城市',
  district VARCHAR(50) COMMENT '区县',
  INDEX idx_city (city),
  INDEX idx_pediatric_endo (pediatric_endo)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='医院表';

-- 医院科室表
CREATE TABLE IF NOT EXISTS hospital_departments (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  hospital_id INT UNSIGNED NOT NULL COMMENT '医院ID',
  name VARCHAR(50) NOT NULL COMMENT '科室名称',
  description VARCHAR(500) COMMENT '科室描述',
  INDEX idx_hospital_id (hospital_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='医院科室表';

-- 会员表
CREATE TABLE IF NOT EXISTS memberships (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id INT UNSIGNED NOT NULL UNIQUE COMMENT '用户ID',
  plan_type VARCHAR(20) COMMENT 'monthly/quarterly/yearly',
  start_date DATE COMMENT '开始时间',
  end_date DATE COMMENT '到期时间',
  status VARCHAR(20) DEFAULT 'active' COMMENT 'active/expired/cancelled',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员表';

-- 额度使用表
CREATE TABLE IF NOT EXISTS usage_quotas (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id INT UNSIGNED NOT NULL COMMENT '用户ID',
  year INT NOT NULL COMMENT '年份',
  month INT NOT NULL COMMENT '月份',
  used_count INT DEFAULT 0 COMMENT '已使用次数',
  free_quota INT DEFAULT 3 COMMENT '免费额度',
  paid_quota INT DEFAULT 20 COMMENT '付费额度',
  UNIQUE KEY uk_user_year_month (user_id, year, month),
  INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='额度使用表';

-- 化验单报告表
CREATE TABLE IF NOT EXISTS reports (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id INT UNSIGNED NOT NULL COMMENT '用户ID',
  child_id INT UNSIGNED COMMENT '关联宝宝ID',
  report_type VARCHAR(50) COMMENT '报告类型',
  image_url VARCHAR(500) NOT NULL COMMENT '图片URL',
  hospital VARCHAR(100) COMMENT '医院名称',
  report_date DATE COMMENT '报告日期',
  analyze_result TEXT COMMENT '解析结果JSON',
  ai_response TEXT COMMENT 'AI解读文本',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_child_id (child_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='化验单报告表';

-- 订阅提醒表
CREATE TABLE IF NOT EXISTS subscription_reminders (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id INT UNSIGNED NOT NULL COMMENT '用户ID',
  child_id INT UNSIGNED COMMENT '宝宝ID',
  day_of_week INT DEFAULT 0 COMMENT '0=周日, 1=周一...',
  time VARCHAR(10) DEFAULT '09:00' COMMENT '提醒时间',
  enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_child_id (child_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订阅提醒表';

-- =====================================================
-- AI功能表
-- =====================================================

-- AI对话表
CREATE TABLE IF NOT EXISTS ai_conversations (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id INT UNSIGNED COMMENT '用户ID',
  child_id INT UNSIGNED COMMENT '宝宝ID',
  session_id VARCHAR(36) COMMENT '会话ID',
  messages TEXT COMMENT '消息列表JSON',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_session_id (session_id),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI对话表';

-- =====================================================
-- 运营表
-- =====================================================

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id INT UNSIGNED COMMENT '操作用户ID',
  action VARCHAR(50) NOT NULL COMMENT '操作类型',
  target_type VARCHAR(50) COMMENT '目标类型',
  target_id INT UNSIGNED COMMENT '目标ID',
  detail TEXT COMMENT '详情JSON',
  ip_address VARCHAR(50) COMMENT 'IP地址',
  user_agent VARCHAR(512) COMMENT 'UserAgent',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_action (action),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';

-- =====================================================
-- 初始数据
-- =====================================================

-- 插入示例医院数据
INSERT INTO hospitals (name, level, address, latitude, longitude, phone, pediatric_endo, estimated_fee, city, district) VALUES
('北京儿童医院', '三级甲等', '北京市西城区南礼士路1号', 39.9425, 116.3567, '010-59616161', TRUE, '500-1500元', '北京市', '西城区'),
('复旦大学附属儿科医院', '三级甲等', '上海市闵行区万源路399号', 31.0282, 121.4612, '021-64931990', TRUE, '500-1500元', '上海市', '闵行区'),
('广州市妇女儿童医疗中心', '三级甲等', '广州市天河区珠江新城金穗路9号', 23.1189, 113.3275, '020-81886332', TRUE, '400-1200元', '广州市', '天河区'),
('浙江大学医学院附属儿童医院', '三级甲等', '杭州市拱墅区竹竿巷57号', 30.2588, 120.1535, '0571-88873114', TRUE, '500-1500元', '杭州市', '拱墅区'),
('四川大学华西第二医院', '三级甲等', '成都市武侯区人民南路三段20号', 30.6415, 104.0445, '028-88570100', TRUE, '400-1200元', '成都市', '武侯区')
ON DUPLICATE KEY UPDATE name=VALUES(name);
