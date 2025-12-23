-- ============================================
-- Astronomer 数据库初始化 SQL 文件
-- 生成时间: 2025-12-22
-- 数据库版本: MySQL 8.0+
-- 字符集: utf8mb4
-- ============================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `astronomer`
DEFAULT CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE `astronomer`;

-- ============================================
-- 1. 用户模块
-- ============================================

-- 用户表
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` VARCHAR(36) PRIMARY KEY COMMENT '主键(UUID)',
  `phone` VARCHAR(20) UNIQUE NOT NULL COMMENT '手机号',
  `username` VARCHAR(255) NOT NULL COMMENT '用户名',
  `password` VARCHAR(255) NOT NULL COMMENT '密码',
  `icon` VARCHAR(500) DEFAULT NULL COMMENT '头像',
  `sex` INT NOT NULL DEFAULT 1 COMMENT '性别(1->男 2->女)',
  `note` VARCHAR(500) DEFAULT NULL COMMENT '备注',
  `intro` VARCHAR(500) DEFAULT NULL COMMENT '个人简介',
  `role` VARCHAR(20) NOT NULL DEFAULT 'user' COMMENT '角色:user-普通用户,admin-管理员,super_admin-超级管理员',
  `following_count` BIGINT NOT NULL DEFAULT 0 COMMENT '关注数量',
  `followed_count` BIGINT NOT NULL DEFAULT 0 COMMENT '被关注数量',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  INDEX `idx_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 用户关注表
DROP TABLE IF EXISTS `user_follow`;
CREATE TABLE `user_follow` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `user_id` VARCHAR(36) NOT NULL COMMENT '关注者ID',
  `follow_user_id` VARCHAR(36) NOT NULL COMMENT '被关注者ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '关注时间',
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_follow_user_id` (`follow_user_id`),
  UNIQUE KEY `uk_user_follow` (`user_id`, `follow_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户关注表';

-- 用户拉黑表
DROP TABLE IF EXISTS `user_block`;
CREATE TABLE `user_block` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
  `block_user_id` VARCHAR(36) NOT NULL COMMENT '被拉黑的用户ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '拉黑时间',
  INDEX `idx_user_id` (`user_id`),
  UNIQUE KEY `uk_user_block` (`user_id`, `block_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户拉黑表';

-- ============================================
-- 2. 文章模块（旧版）
-- ============================================

-- 文章表（旧版）
DROP TABLE IF EXISTS `article`;
CREATE TABLE `article` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  `user_id` VARCHAR(36) NOT NULL COMMENT '用户id',
  `title` VARCHAR(150) NOT NULL COMMENT '标题',
  `preface` VARCHAR(255) DEFAULT NULL COMMENT '简介',
  `photo` VARCHAR(200) DEFAULT NULL COMMENT '图片',
  `tag` VARCHAR(200) DEFAULT NULL COMMENT '标签',
  `status` INT NOT NULL DEFAULT 1 COMMENT '状态：0-草稿 1-已发布 2-已删除',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `visit` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '阅读量',
  `content` TEXT COMMENT '文章内容',
  `good_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文章点赞量',
  `appear` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否出现',
  `comment` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否允许评论',
  `comment_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文章评论量',
  `favorite_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收藏数量',
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章表（旧版）';

-- 文章点赞表
DROP TABLE IF EXISTS `article_star`;
CREATE TABLE `article_star` (
  `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `user_id` VARCHAR(36) NOT NULL,
  `article_id` INT UNSIGNED NOT NULL,
  UNIQUE KEY `uk_user_article` (`user_id`, `article_id`),
  INDEX `idx_article_id` (`article_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章点赞表';

-- ============================================
-- 3. 文章模块（V3版本 - 企业级）
-- ============================================

-- 文章主表 V3
DROP TABLE IF EXISTS `article_v3`;
CREATE TABLE `article_v3` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `user_id` VARCHAR(36) NOT NULL,
  -- 基础信息
  `title` VARCHAR(200) NOT NULL,
  `summary` VARCHAR(500) DEFAULT NULL,
  `cover_image` VARCHAR(500) DEFAULT NULL,
  `content_type` TINYINT NOT NULL DEFAULT 1 COMMENT '1-图文 2-视频 3-音频 4-问答',
  -- 分类与标签
  `category_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `column_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `tags` VARCHAR(500) DEFAULT NULL COMMENT 'JSON数组',
  `topics` VARCHAR(500) DEFAULT NULL COMMENT 'JSON数组',
  -- 状态管理
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1-已发布 2-审核中 3-审核失败 4-已下线 5-已删除',
  `visibility` TINYINT NOT NULL DEFAULT 1 COMMENT '1-公开 2-仅粉丝 3-仅好友 4-私密 5-付费',
  `allow_comment` TINYINT(1) NOT NULL DEFAULT 1,
  `allow_repost` TINYINT(1) NOT NULL DEFAULT 1,
  -- 内容审核
  `audit_status` TINYINT NOT NULL DEFAULT 0 COMMENT '0-待审核 1-通过 2-驳回',
  `audit_reason` VARCHAR(200) DEFAULT NULL,
  `audit_time` DATETIME DEFAULT NULL,
  `audit_user_id` VARCHAR(36) DEFAULT NULL,
  -- 统计数据
  `view_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `real_view_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '真实浏览量（去重）',
  `like_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `comment_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `share_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `favorite_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  -- 推荐与排序
  `weight` INT NOT NULL DEFAULT 0 COMMENT '权重（影响排序）',
  `hot_score` DECIMAL(10,2) NOT NULL DEFAULT 0,
  `quality_score` DECIMAL(5,2) NOT NULL DEFAULT 0 COMMENT '质量分数（AI评分）',
  `is_featured` TINYINT(1) NOT NULL DEFAULT 0,
  `is_top` TINYINT(1) NOT NULL DEFAULT 0,
  `is_hot` TINYINT(1) NOT NULL DEFAULT 0,
  `is_recommend` TINYINT(1) NOT NULL DEFAULT 0,
  -- SEO优化
  `keywords` VARCHAR(200) DEFAULT NULL,
  `description` VARCHAR(500) DEFAULT NULL,
  `slug` VARCHAR(200) DEFAULT NULL,
  -- 付费相关
  `is_paid` TINYINT(1) NOT NULL DEFAULT 0,
  `price` DECIMAL(10,2) NOT NULL DEFAULT 0,
  `free_content` TEXT DEFAULT NULL,
  -- 时间戳
  `publish_time` DATETIME DEFAULT NULL,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `delete_time` DATETIME DEFAULT NULL,
  -- 扩展字段
  `ext_info` JSON DEFAULT NULL,
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_category` (`category_id`),
  INDEX `idx_column` (`column_id`),
  INDEX `idx_featured` (`is_featured`),
  INDEX `idx_hot` (`is_hot`),
  INDEX `idx_slug` (`slug`),
  INDEX `idx_status_publish` (`status`, `publish_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章主表 V3';

-- 文章内容表（内容分离）
DROP TABLE IF EXISTS `article_content`;
CREATE TABLE `article_content` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `article_id` BIGINT UNSIGNED NOT NULL UNIQUE COMMENT '文章ID',
  `content` LONGTEXT NOT NULL COMMENT 'Markdown格式',
  `content_html` LONGTEXT DEFAULT NULL COMMENT 'HTML格式',
  `toc` JSON DEFAULT NULL COMMENT '目录（自动生成）',
  `word_count` INT NOT NULL DEFAULT 0,
  `read_time` INT NOT NULL DEFAULT 0 COMMENT '预计阅读时间（分钟）',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX `idx_article` (`article_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章内容表';

-- 文章草稿表
DROP TABLE IF EXISTS `article_draft`;
CREATE TABLE `article_draft` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `user_id` VARCHAR(36) NOT NULL,
  `article_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联的文章ID（0表示新建）',
  `title` VARCHAR(200) DEFAULT NULL,
  `summary` VARCHAR(500) DEFAULT NULL,
  `cover_image` VARCHAR(500) DEFAULT NULL,
  `content` LONGTEXT DEFAULT NULL,
  `category_id` BIGINT UNSIGNED DEFAULT NULL,
  `column_id` BIGINT UNSIGNED DEFAULT NULL,
  `tags` VARCHAR(500) DEFAULT NULL,
  `topics` VARCHAR(500) DEFAULT NULL,
  `auto_save_count` INT NOT NULL DEFAULT 0 COMMENT '自动保存次数',
  `last_edit_time` DATETIME DEFAULT NULL,
  `is_published` TINYINT(1) NOT NULL DEFAULT 0,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX `idx_user` (`user_id`, `is_published`),
  INDEX `idx_article` (`article_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章草稿表';

-- 文章历史版本表
DROP TABLE IF EXISTS `article_history`;
CREATE TABLE `article_history` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `article_id` BIGINT UNSIGNED NOT NULL,
  `version` INT NOT NULL,
  `title` VARCHAR(200) DEFAULT NULL,
  `content` LONGTEXT DEFAULT NULL,
  `summary` VARCHAR(500) DEFAULT NULL,
  `change_type` TINYINT NOT NULL COMMENT '1-创建 2-编辑 3-发布 4-下线',
  `change_reason` VARCHAR(200) DEFAULT NULL,
  `operator_id` VARCHAR(36) DEFAULT NULL,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX `idx_article_version` (`article_id`, `version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章历史版本表';

-- 文章分类表
DROP TABLE IF EXISTS `article_category`;
CREATE TABLE `article_category` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(50) NOT NULL,
  `parent_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `icon` VARCHAR(200) DEFAULT NULL,
  `sort_order` INT NOT NULL DEFAULT 0,
  `article_count` INT NOT NULL DEFAULT 0,
  `is_show` TINYINT(1) NOT NULL DEFAULT 1,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX `idx_parent` (`parent_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章分类表';

-- 专栏表
DROP TABLE IF EXISTS `article_column`;
CREATE TABLE `article_column` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `user_id` VARCHAR(36) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `description` VARCHAR(500) DEFAULT NULL,
  `cover_image` VARCHAR(500) DEFAULT NULL,
  `article_count` INT NOT NULL DEFAULT 0,
  `subscriber_count` INT NOT NULL DEFAULT 0,
  `is_finished` TINYINT(1) NOT NULL DEFAULT 0,
  `sort_type` TINYINT NOT NULL DEFAULT 1 COMMENT '1-自定义 2-时间正序 3-时间倒序',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1-正常 2-隐藏',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='专栏表';

-- 专栏文章关联表
DROP TABLE IF EXISTS `article_column_rel`;
CREATE TABLE `article_column_rel` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `column_id` BIGINT UNSIGNED NOT NULL,
  `article_id` BIGINT UNSIGNED NOT NULL,
  `sort_order` INT NOT NULL DEFAULT 0,
  `add_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_column_article` (`column_id`, `article_id`),
  INDEX `idx_column_sort` (`column_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='专栏文章关联表';

-- 专栏订阅表
DROP TABLE IF EXISTS `column_subscription`;
CREATE TABLE `column_subscription` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `column_id` BIGINT UNSIGNED NOT NULL COMMENT '专栏ID',
  `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订阅时间',
  INDEX `idx_column_id` (`column_id`),
  INDEX `idx_user_id` (`user_id`),
  UNIQUE KEY `uk_column_user` (`column_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='专栏订阅表';

-- 话题表
DROP TABLE IF EXISTS `topic`;
CREATE TABLE `topic` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL UNIQUE,
  `description` VARCHAR(500) DEFAULT NULL,
  `cover_image` VARCHAR(500) DEFAULT NULL,
  `article_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `follow_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `view_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `is_hot` TINYINT(1) NOT NULL DEFAULT 0,
  `is_recommend` TINYINT(1) NOT NULL DEFAULT 0,
  `category` VARCHAR(50) DEFAULT NULL,
  `creator_id` VARCHAR(36) DEFAULT NULL,
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1-正常 2-隐藏 3-封禁',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX `idx_name` (`name`),
  INDEX `idx_hot` (`is_hot`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='话题表';

-- 文章话题关联表
DROP TABLE IF EXISTS `article_topic_rel`;
CREATE TABLE `article_topic_rel` (
  `article_id` BIGINT UNSIGNED NOT NULL,
  `topic_id` BIGINT UNSIGNED NOT NULL,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`article_id`, `topic_id`),
  INDEX `idx_topic` (`topic_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章话题关联表';

-- 话题关注表
DROP TABLE IF EXISTS `topic_follow`;
CREATE TABLE `topic_follow` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `topic_id` BIGINT UNSIGNED NOT NULL COMMENT '话题ID',
  `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '关注时间',
  INDEX `idx_topic_id` (`topic_id`),
  INDEX `idx_user_id` (`user_id`),
  UNIQUE KEY `uk_topic_user` (`topic_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='话题关注表';

-- 文章标签表
DROP TABLE IF EXISTS `article_tag`;
CREATE TABLE `article_tag` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(50) NOT NULL UNIQUE,
  `article_count` INT NOT NULL DEFAULT 0,
  `follow_count` INT NOT NULL DEFAULT 0,
  `is_hot` TINYINT(1) NOT NULL DEFAULT 0,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章标签表';

-- 文章相关推荐表
DROP TABLE IF EXISTS `article_relation`;
CREATE TABLE `article_relation` (
  `article_id` BIGINT UNSIGNED NOT NULL,
  `related_article_id` BIGINT UNSIGNED NOT NULL,
  `relevance_score` DECIMAL(5,2) NOT NULL DEFAULT 0,
  `relation_type` TINYINT NOT NULL COMMENT '1-同作者 2-同分类 3-同话题 4-算法推荐',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`article_id`, `related_article_id`),
  INDEX `idx_article_score` (`article_id`, `relevance_score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章相关推荐表';

-- 文章统计详情表
DROP TABLE IF EXISTS `article_stats_detail`;
CREATE TABLE `article_stats_detail` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `article_id` BIGINT UNSIGNED NOT NULL UNIQUE,
  `today_view_count` INT NOT NULL DEFAULT 0,
  `week_view_count` INT NOT NULL DEFAULT 0,
  `month_view_count` INT NOT NULL DEFAULT 0,
  `today_like_count` INT NOT NULL DEFAULT 0,
  `today_comment_count` INT NOT NULL DEFAULT 0,
  `today_share_count` INT NOT NULL DEFAULT 0,
  `source_stats` JSON DEFAULT NULL COMMENT '流量来源统计',
  `device_stats` JSON DEFAULT NULL COMMENT '设备统计',
  `region_stats` JSON DEFAULT NULL COMMENT '地域统计',
  `avg_read_progress` DECIMAL(5,2) NOT NULL DEFAULT 0 COMMENT '平均阅读进度（%）',
  `avg_stay_time` INT NOT NULL DEFAULT 0 COMMENT '平均停留时间（秒）',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章统计详情表';

-- ============================================
-- 4. 评论模块（旧版）
-- ============================================

-- 一级评论表
DROP TABLE IF EXISTS `comment_parent`;
CREATE TABLE `comment_parent` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `article_id` VARCHAR(20) DEFAULT NULL,
  `comment` LONGTEXT,
  `comment_time` LONGTEXT,
  `username` VARCHAR(255) DEFAULT NULL,
  `user_id` VARCHAR(36) DEFAULT NULL,
  `phone` VARCHAR(255) DEFAULT NULL,
  `good_count` BIGINT UNSIGNED DEFAULT 0,
  `comment_addr` VARCHAR(255) DEFAULT NULL,
  INDEX `idx_article_id` (`article_id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='一级评论表';

-- 二级评论表
DROP TABLE IF EXISTS `comment_sub_two`;
CREATE TABLE `comment_sub_two` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `parent_comment_id` VARCHAR(20) DEFAULT NULL,
  `comment` LONGTEXT,
  `comment_time` LONGTEXT,
  `username` VARCHAR(255) DEFAULT NULL,
  `user_id` VARCHAR(36) DEFAULT NULL,
  `phone` VARCHAR(255) DEFAULT NULL,
  `good_count` BIGINT UNSIGNED DEFAULT 0,
  `to_username` VARCHAR(255) DEFAULT NULL,
  `to_phone` VARCHAR(255) DEFAULT NULL,
  `to_user_id` VARCHAR(36) DEFAULT NULL,
  `comment_addr` VARCHAR(255) DEFAULT NULL,
  INDEX `idx_parent_comment_id` (`parent_comment_id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='二级评论表';

-- 一级评论点赞表
DROP TABLE IF EXISTS `comment_parent_like`;
CREATE TABLE `comment_parent_like` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `comment_id` BIGINT DEFAULT NULL,
  `user_id` VARCHAR(36) DEFAULT NULL,
  UNIQUE KEY `uk_comment_user` (`comment_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='一级评论点赞表';

-- 二级评论点赞表
DROP TABLE IF EXISTS `comment_sub_two_like`;
CREATE TABLE `comment_sub_two_like` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `comment_id` BIGINT DEFAULT NULL,
  `user_id` VARCHAR(36) DEFAULT NULL,
  UNIQUE KEY `uk_comment_user` (`comment_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='二级评论点赞表';

-- ============================================
-- 5. 评论模块（V3版本 - 企业级）
-- ============================================

-- 评论主表 V3
DROP TABLE IF EXISTS `comment_v3`;
CREATE TABLE `comment_v3` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  -- 所属对象
  `target_type` TINYINT NOT NULL COMMENT '1-文章 2-视频 3-问答 4-动态',
  `target_id` BIGINT UNSIGNED NOT NULL,
  -- 用户信息
  `user_id` VARCHAR(36) NOT NULL,
  `username` VARCHAR(100) DEFAULT NULL COMMENT '用户名（冗余）',
  `user_avatar` VARCHAR(500) DEFAULT NULL COMMENT '用户头像（冗余）',
  -- 评论结构
  `parent_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父评论ID（0表示根评论）',
  `root_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '根评论ID',
  `reply_to_user_id` VARCHAR(36) NOT NULL DEFAULT '',
  `reply_to_comment_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  -- 楼层信息
  `floor_number` INT NOT NULL DEFAULT 0 COMMENT '楼层号',
  `sub_floor_number` INT NOT NULL DEFAULT 0 COMMENT '子楼层号',
  `reply_chain` VARCHAR(1000) DEFAULT NULL COMMENT '回复链路',
  `depth` INT NOT NULL DEFAULT 0 COMMENT '评论深度',
  -- 评论内容
  `content` TEXT NOT NULL,
  `content_type` TINYINT NOT NULL DEFAULT 1 COMMENT '1-文本 2-图片 3-表情包',
  `images` VARCHAR(1000) DEFAULT NULL COMMENT '图片URL（JSON数组）',
  `at_user_ids` VARCHAR(500) DEFAULT NULL COMMENT '@的用户ID列表',
  -- 评论状态
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1-正常 2-审核中 3-已删除 4-已折叠 5-已屏蔽',
  `is_pinned` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否置顶',
  `is_author` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否作者评论',
  `is_hot` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否热评',
  `is_featured` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否精选评论',
  -- 互动数据
  `like_count` INT NOT NULL DEFAULT 0,
  `dislike_count` INT NOT NULL DEFAULT 0,
  `reply_count` INT NOT NULL DEFAULT 0 COMMENT '直接回复数',
  `total_reply_count` INT NOT NULL DEFAULT 0 COMMENT '总回复数',
  -- 热度计算
  `hot_score` DECIMAL(10,2) NOT NULL DEFAULT 0,
  `quality_score` DECIMAL(5,2) NOT NULL DEFAULT 0,
  -- IP与设备
  `ip` VARCHAR(50) DEFAULT NULL,
  `ip_location` VARCHAR(100) DEFAULT NULL,
  `device_type` VARCHAR(50) DEFAULT NULL,
  `user_agent` TEXT DEFAULT NULL,
  -- 审核相关
  `audit_status` TINYINT NOT NULL DEFAULT 0 COMMENT '0-待审核 1-通过 2-不通过',
  `audit_reason` VARCHAR(200) DEFAULT NULL,
  `risk_level` TINYINT NOT NULL DEFAULT 0 COMMENT '0-正常 1-低风险 2-中风险 3-高风险',
  -- 时间戳
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `delete_time` DATETIME DEFAULT NULL,
  -- 扩展字段
  `ext_info` JSON DEFAULT NULL,
  INDEX `idx_target` (`target_type`, `target_id`, `create_time`),
  INDEX `idx_user` (`user_id`),
  INDEX `idx_parent` (`parent_id`),
  INDEX `idx_root` (`root_id`),
  INDEX `idx_floor` (`floor_number`),
  INDEX `idx_hot` (`is_hot`, `hot_score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论主表 V3';

-- 评论互动表
DROP TABLE IF EXISTS `comment_interaction`;
CREATE TABLE `comment_interaction` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `comment_id` BIGINT UNSIGNED NOT NULL,
  `user_id` VARCHAR(36) NOT NULL,
  `action_type` TINYINT NOT NULL COMMENT '1-点赞 2-踩 3-举报',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_comment_user_action` (`comment_id`, `user_id`, `action_type`),
  INDEX `idx_comment` (`comment_id`),
  INDEX `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论互动表';

-- 评论热榜表
DROP TABLE IF EXISTS `comment_hot_list`;
CREATE TABLE `comment_hot_list` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `target_type` TINYINT NOT NULL,
  `target_id` BIGINT UNSIGNED NOT NULL,
  `comment_id` BIGINT UNSIGNED NOT NULL,
  `user_id` VARCHAR(36) DEFAULT NULL,
  `username` VARCHAR(100) DEFAULT NULL,
  `content` TEXT DEFAULT NULL,
  `like_count` INT DEFAULT NULL,
  `rank_position` INT NOT NULL COMMENT '排名位置',
  `hot_score` DECIMAL(10,2) DEFAULT NULL,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_target_comment` (`target_type`, `target_id`, `comment_id`),
  INDEX `idx_target_rank` (`target_id`, `rank_position`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论热榜表';

-- 评论举报表
DROP TABLE IF EXISTS `comment_report`;
CREATE TABLE `comment_report` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `comment_id` BIGINT UNSIGNED NOT NULL,
  `reporter_user_id` VARCHAR(36) NOT NULL,
  `reason_type` TINYINT NOT NULL COMMENT '1-垃圾广告 2-色情低俗 3-政治敏感 4-人身攻击 5-造谣传谣',
  `reason_desc` VARCHAR(500) DEFAULT NULL,
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0-待处理 1-已处理-成立 2-已处理-不成立',
  `handle_result` VARCHAR(200) DEFAULT NULL,
  `handle_user_id` VARCHAR(36) DEFAULT NULL,
  `handle_time` DATETIME DEFAULT NULL,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX `idx_comment` (`comment_id`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论举报表';

-- 评论盖楼表
DROP TABLE IF EXISTS `comment_floor_building`;
CREATE TABLE `comment_floor_building` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `target_type` TINYINT NOT NULL,
  `target_id` BIGINT UNSIGNED NOT NULL,
  `user_id` VARCHAR(36) NOT NULL,
  `comment_ids` VARCHAR(1000) DEFAULT NULL COMMENT '盖楼的评论ID列表（JSON）',
  `floor_count` INT NOT NULL DEFAULT 0,
  `first_comment_time` DATETIME DEFAULT NULL,
  `last_comment_time` DATETIME DEFAULT NULL,
  INDEX `idx_target_user` (`target_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论盖楼表';

-- UP主追评表
DROP TABLE IF EXISTS `comment_author_reply`;
CREATE TABLE `comment_author_reply` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `comment_id` BIGINT UNSIGNED NOT NULL,
  `author_user_id` VARCHAR(36) NOT NULL,
  `content` TEXT NOT NULL,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX `idx_comment` (`comment_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='UP主追评表';

-- 评论表情包表
DROP TABLE IF EXISTS `comment_emotion`;
CREATE TABLE `comment_emotion` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(50) NOT NULL,
  `image_url` VARCHAR(500) NOT NULL,
  `category` VARCHAR(50) DEFAULT NULL,
  `use_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `is_hot` TINYINT(1) NOT NULL DEFAULT 0,
  `sort_order` INT NOT NULL DEFAULT 0,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表情包表';

-- 评论敏感词库
DROP TABLE IF EXISTS `comment_sensitive_word`;
CREATE TABLE `comment_sensitive_word` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `word` VARCHAR(100) NOT NULL UNIQUE,
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '1-一般 2-严重 3-非常严重',
  `action` TINYINT NOT NULL DEFAULT 1 COMMENT '1-替换 2-拦截 3-人工审核',
  `replacement` VARCHAR(100) DEFAULT NULL,
  `category` VARCHAR(50) DEFAULT NULL COMMENT '政治/色情/广告',
  `is_enabled` TINYINT(1) NOT NULL DEFAULT 1,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论敏感词库';

-- 评论统计表
DROP TABLE IF EXISTS `comment_stats`;
CREATE TABLE `comment_stats` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `target_type` TINYINT NOT NULL,
  `target_id` BIGINT UNSIGNED NOT NULL,
  `total_comment_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `today_comment_count` INT NOT NULL DEFAULT 0,
  `root_comment_count` INT NOT NULL DEFAULT 0,
  `avg_comment_length` DECIMAL(10,2) DEFAULT NULL,
  `hot_comment_count` INT NOT NULL DEFAULT 0,
  `last_comment_time` DATETIME DEFAULT NULL,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_target` (`target_type`, `target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论统计表';

-- 评论折叠规则表
DROP TABLE IF EXISTS `comment_fold_rule`;
CREATE TABLE `comment_fold_rule` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `rule_name` VARCHAR(100) DEFAULT NULL,
  `rule_type` TINYINT NOT NULL COMMENT '1-关键词 2-低赞 3-举报数 4-用户等级',
  `rule_config` JSON DEFAULT NULL COMMENT '规则配置',
  `is_enabled` TINYINT(1) NOT NULL DEFAULT 1,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论折叠规则表';

-- ============================================
-- 6. 收藏与通知模块
-- ============================================

-- 用户收藏表
DROP TABLE IF EXISTS `user_favorite`;
CREATE TABLE `user_favorite` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
  `article_id` BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '收藏时间',
  UNIQUE KEY `uk_user_article` (`user_id`, `article_id`),
  INDEX `idx_article_id` (`article_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户收藏表';

-- 通知表
DROP TABLE IF EXISTS `notification`;
CREATE TABLE `notification` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `user_id` VARCHAR(36) NOT NULL COMMENT '接收者ID',
  `type` INT NOT NULL COMMENT '通知类型：1-点赞文章 2-评论文章 3-回复评论 4-关注 5-点赞评论',
  `from_user_id` VARCHAR(36) DEFAULT NULL COMMENT '触发通知的用户ID',
  `from_username` VARCHAR(100) DEFAULT NULL COMMENT '触发通知的用户名',
  `content` VARCHAR(500) DEFAULT NULL COMMENT '通知内容',
  `related_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联ID',
  `related_type` VARCHAR(50) DEFAULT NULL COMMENT '关联类型：article/comment/user',
  `is_read` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已读',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_is_read` (`user_id`, `is_read`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知表';

-- ============================================
-- 7. 标签模块
-- ============================================

-- 标签表
DROP TABLE IF EXISTS `tag`;
CREATE TABLE `tag` (
  `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
  `tag_name` VARCHAR(32) NOT NULL COMMENT '标签名称',
  `sort` INT UNSIGNED DEFAULT NULL COMMENT '排序',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表';

-- ============================================
-- 初始化完成
-- ============================================

-- 插入一些初始数据（可选）

-- 插入默认分类
INSERT INTO `article_category` (`name`, `parent_id`, `sort_order`, `is_show`) VALUES
('技术', 0, 1, 1),
('后端开发', 1, 1, 1),
('前端开发', 1, 2, 1),
('移动开发', 1, 3, 1),
('生活', 0, 2, 1),
('职场', 0, 3, 1);

-- 插入默认话题
INSERT INTO `topic` (`name`, `description`, `category`, `is_hot`, `status`) VALUES
('Go语言', 'Go语言技术交流', '技术', 1, 1),
('Vue3', 'Vue3框架学习', '技术', 1, 1),
('职场经验', '职场生活分享', '职场', 0, 1);

SELECT '✅ 数据库初始化完成！' AS 'Status';
SELECT CONCAT('✅ 共创建了 ', COUNT(*), ' 张表') AS 'Tables Count'
FROM information_schema.tables
WHERE table_schema = 'astronomer';
