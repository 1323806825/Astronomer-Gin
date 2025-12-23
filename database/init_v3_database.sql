-- ==================== Astronomer-Gin 企业级数据库初始化脚本 ====================
-- 版本: V3.0
-- 创建日期: 2025-12-04
-- 说明: 包含所有V3版本的表结构

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ==================== 文章模块（12张表） ====================

-- 1. 文章主表（核心）
DROP TABLE IF EXISTS `article_v3`;
CREATE TABLE `article_v3` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '文章ID',
    `user_id` BIGINT NOT NULL COMMENT '作者ID',

    -- 基础信息
    `title` VARCHAR(200) NOT NULL COMMENT '标题',
    `summary` VARCHAR(500) DEFAULT NULL COMMENT '摘要',
    `cover_image` VARCHAR(500) DEFAULT NULL COMMENT '封面图',
    `content_type` TINYINT DEFAULT 1 COMMENT '内容类型：1-���文 2-视频 3-音频 4-问答',

    -- 分类与标签
    `category_id` BIGINT DEFAULT 0 COMMENT '分类ID',
    `column_id` BIGINT DEFAULT 0 COMMENT '专栏ID',
    `tags` VARCHAR(500) DEFAULT NULL COMMENT '标签（JSON数组）',
    `topics` VARCHAR(500) DEFAULT NULL COMMENT '话题（JSON数组）',

    -- 状态管理
    `status` TINYINT DEFAULT 1 COMMENT '状态：1-已发布 2-审核中 3-审核失败 4-已下线 5-已删除',
    `visibility` TINYINT DEFAULT 1 COMMENT '可见性：1-公开 2-仅粉丝 3-仅好友 4-私密 5-付费',
    `allow_comment` BOOLEAN DEFAULT TRUE COMMENT '是否允许评论',
    `allow_repost` BOOLEAN DEFAULT TRUE COMMENT '是否允许转发',

    -- 内容审核
    `audit_status` TINYINT DEFAULT 0 COMMENT '审核状态：0-待审核 1-通过 2-驳回',
    `audit_reason` VARCHAR(200) DEFAULT NULL COMMENT '审核意见',
    `audit_time` TIMESTAMP NULL DEFAULT NULL COMMENT '审核时间',
    `audit_user_id` BIGINT DEFAULT NULL COMMENT '审核人ID',

    -- 统计数据
    `view_count` BIGINT DEFAULT 0 COMMENT '浏览量',
    `real_view_count` BIGINT DEFAULT 0 COMMENT '真实浏览量（去重）',
    `like_count` BIGINT DEFAULT 0 COMMENT '点赞数',
    `comment_count` BIGINT DEFAULT 0 COMMENT '评论数',
    `share_count` BIGINT DEFAULT 0 COMMENT '分享数',
    `favorite_count` BIGINT DEFAULT 0 COMMENT '收藏数',

    -- 推荐与排序
    `weight` INT DEFAULT 0 COMMENT '权重（影响排序）',
    `hot_score` DECIMAL(10,2) DEFAULT 0 COMMENT '热度分数',
    `quality_score` DECIMAL(5,2) DEFAULT 0 COMMENT '质量分数（AI评分）',
    `is_featured` BOOLEAN DEFAULT FALSE COMMENT '是否精选',
    `is_top` BOOLEAN DEFAULT FALSE COMMENT '是否置顶',
    `is_hot` BOOLEAN DEFAULT FALSE COMMENT '是否热门',
    `is_recommend` BOOLEAN DEFAULT FALSE COMMENT '是否推荐',

    -- SEO优化
    `keywords` VARCHAR(200) DEFAULT NULL COMMENT 'SEO关键词',
    `description` VARCHAR(500) DEFAULT NULL COMMENT 'SEO描述',
    `slug` VARCHAR(200) DEFAULT NULL COMMENT 'URL别名',

    -- 付费相关
    `is_paid` BOOLEAN DEFAULT FALSE COMMENT '是否付费内容',
    `price` DECIMAL(10,2) DEFAULT 0 COMMENT '价格',
    `free_content` TEXT DEFAULT NULL COMMENT '免费预览内容',

    -- 时间戳
    `publish_time` TIMESTAMP NULL DEFAULT NULL COMMENT '发布时间',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` TIMESTAMP NULL DEFAULT NULL COMMENT '删除时间（软删除）',

    -- 扩展字段
    `ext_info` JSON DEFAULT NULL COMMENT '扩展信息',

    INDEX `idx_user_id` (`user_id`, `status`, `publish_time` DESC),
    INDEX `idx_category` (`category_id`, `status`, `hot_score` DESC),
    INDEX `idx_column` (`column_id`, `status`),
    INDEX `idx_status_publish` (`status`, `publish_time` DESC),
    INDEX `idx_hot` (`is_hot`, `hot_score` DESC),
    INDEX `idx_featured` (`is_featured`, `publish_time` DESC),
    UNIQUE INDEX `idx_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章主表（企业级）';

-- 2. 文章内容表（内容分离）
DROP TABLE IF EXISTS `article_content`;
CREATE TABLE `article_content` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `article_id` BIGINT NOT NULL UNIQUE COMMENT '文章ID',
    `content` LONGTEXT NOT NULL COMMENT '文章内容（Markdown）',
    `content_html` LONGTEXT DEFAULT NULL COMMENT '文章内容（HTML）',
    `toc` JSON DEFAULT NULL COMMENT '目录（自动生成）',
    `word_count` INT DEFAULT 0 COMMENT '字数',
    `read_time` INT DEFAULT 0 COMMENT '预计阅读时间（分钟）',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `update_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_article` (`article_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章内容表';

-- 3. 文章草稿表（独立管理）
DROP TABLE IF EXISTS `article_draft`;
CREATE TABLE `article_draft` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL COMMENT '作者ID',
    `article_id` BIGINT DEFAULT 0 COMMENT '关联的文章ID（0表示新建）',

    -- 草稿内容
    `title` VARCHAR(200) DEFAULT NULL COMMENT '标题',
    `summary` VARCHAR(500) DEFAULT NULL COMMENT '摘要',
    `cover_image` VARCHAR(500) DEFAULT NULL COMMENT '封面图',
    `content` LONGTEXT DEFAULT NULL COMMENT '内容',
    `category_id` BIGINT DEFAULT NULL,
    `column_id` BIGINT DEFAULT NULL,
    `tags` VARCHAR(500) DEFAULT NULL COMMENT '标签',
    `topics` VARCHAR(500) DEFAULT NULL COMMENT '话题',

    -- 草稿管理
    `auto_save_count` INT DEFAULT 0 COMMENT '自动保存次数',
    `last_edit_time` TIMESTAMP NULL DEFAULT NULL COMMENT '最后编辑时间',
    `is_published` BOOLEAN DEFAULT FALSE COMMENT '是否已发布',

    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `update_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX `idx_user` (`user_id`, `is_published`, `update_time` DESC),
    INDEX `idx_article` (`article_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章草稿表';

-- 4. 文章历史版本表（版本控制）
DROP TABLE IF EXISTS `article_history`;
CREATE TABLE `article_history` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `article_id` BIGINT NOT NULL COMMENT '文章ID',
    `version` INT NOT NULL COMMENT '版本号',

    -- 快照数据
    `title` VARCHAR(200) DEFAULT NULL COMMENT '标题',
    `content` LONGTEXT DEFAULT NULL COMMENT '内容',
    `summary` VARCHAR(500) DEFAULT NULL COMMENT '摘要',

    -- 变更信息
    `change_type` TINYINT DEFAULT NULL COMMENT '变更类型：1-创建 2-编辑 3-发布 4-下线',
    `change_reason` VARCHAR(200) DEFAULT NULL COMMENT '变更原因',
    `operator_id` BIGINT DEFAULT NULL COMMENT '操作人ID',

    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_article_version` (`article_id`, `version` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章历史版本';

-- 5. 文章分类表
DROP TABLE IF EXISTS `article_category`;
CREATE TABLE `article_category` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL COMMENT '分类名称',
    `parent_id` BIGINT DEFAULT 0 COMMENT '父分类ID（支持多级分类）',
    `icon` VARCHAR(200) DEFAULT NULL COMMENT '分类图标',
    `sort_order` INT DEFAULT 0 COMMENT '排序',
    `article_count` INT DEFAULT 0 COMMENT '文章数',
    `is_show` BOOLEAN DEFAULT TRUE COMMENT '是否显示',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_parent` (`parent_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章分类';

-- 6. 专栏表（系列文章）
DROP TABLE IF EXISTS `article_column`;
CREATE TABLE `article_column` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL COMMENT '作者ID',
    `name` VARCHAR(100) NOT NULL COMMENT '专栏名称',
    `description` VARCHAR(500) DEFAULT NULL COMMENT '专栏简介',
    `cover_image` VARCHAR(500) DEFAULT NULL COMMENT '封面图',

    `article_count` INT DEFAULT 0 COMMENT '文章数',
    `subscriber_count` INT DEFAULT 0 COMMENT '订阅数',

    `is_finished` BOOLEAN DEFAULT FALSE COMMENT '是否完结',
    `sort_type` TINYINT DEFAULT 1 COMMENT '排序方式：1-自定义 2-时间正序 3-时间倒序',

    `status` TINYINT DEFAULT 1 COMMENT '状态：1-正常 2-隐藏',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `update_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX `idx_user` (`user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章专栏';

-- 7. 专栏-文章关联表
DROP TABLE IF EXISTS `article_column_rel`;
CREATE TABLE `article_column_rel` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `column_id` BIGINT NOT NULL COMMENT '专栏ID',
    `article_id` BIGINT NOT NULL COMMENT '文章ID',
    `sort_order` INT DEFAULT 0 COMMENT '排序',
    `add_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_column_article` (`column_id`, `article_id`),
    INDEX `idx_column_sort` (`column_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='专栏文章关联';

-- 8. 话题表（类似微博话题）
DROP TABLE IF EXISTS `topic`;
CREATE TABLE `topic` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL UNIQUE COMMENT '话题名称（#前端开发）',
    `description` VARCHAR(500) DEFAULT NULL COMMENT '话题描述',
    `cover_image` VARCHAR(500) DEFAULT NULL COMMENT '话题封面',

    `article_count` BIGINT DEFAULT 0 COMMENT '文章数',
    `follow_count` BIGINT DEFAULT 0 COMMENT '关注数',
    `view_count` BIGINT DEFAULT 0 COMMENT '浏览量',

    `is_hot` BOOLEAN DEFAULT FALSE COMMENT '是否热门',
    `is_recommend` BOOLEAN DEFAULT FALSE COMMENT '是否推荐',

    `category` VARCHAR(50) DEFAULT NULL COMMENT '话题分类',
    `creator_id` BIGINT DEFAULT NULL COMMENT '创建者ID',

    `status` TINYINT DEFAULT 1 COMMENT '状态：1-正常 2-隐藏 3-封禁',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_hot` (`is_hot`, `article_count` DESC),
    INDEX `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='话题表';

-- 9. 文章-话题关联表
DROP TABLE IF EXISTS `article_topic_rel`;
CREATE TABLE `article_topic_rel` (
    `article_id` BIGINT NOT NULL COMMENT '文章ID',
    `topic_id` BIGINT NOT NULL COMMENT '话题ID',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`article_id`, `topic_id`),
    INDEX `idx_topic` (`topic_id`, `create_time` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章话题关联';

-- 10. 文章标签表
DROP TABLE IF EXISTS `article_tag`;
CREATE TABLE `article_tag` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL UNIQUE COMMENT '标签名称',
    `article_count` INT DEFAULT 0 COMMENT '文章数',
    `follow_count` INT DEFAULT 0 COMMENT '关注数',
    `is_hot` BOOLEAN DEFAULT FALSE COMMENT '是否热门',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章标签';

-- 11. 文章相关推荐关联表（机器学习推荐）
DROP TABLE IF EXISTS `article_relation`;
CREATE TABLE `article_relation` (
    `article_id` BIGINT NOT NULL COMMENT '文章ID',
    `related_article_id` BIGINT NOT NULL COMMENT '相关文章ID',
    `relevance_score` DECIMAL(5,2) DEFAULT 0 COMMENT '相关度分数',
    `relation_type` TINYINT DEFAULT NULL COMMENT '关联类型：1-同作者 2-同分类 3-同话题 4-算法推荐',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`article_id`, `related_article_id`),
    INDEX `idx_article_score` (`article_id`, `relevance_score` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章相关推荐';

-- 12. 文章统计详情表（分离热数据）
DROP TABLE IF EXISTS `article_stats_detail`;
CREATE TABLE `article_stats_detail` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `article_id` BIGINT NOT NULL UNIQUE COMMENT '文章ID',

    -- UV/PV统计
    `today_view_count` INT DEFAULT 0 COMMENT '今日浏览',
    `week_view_count` INT DEFAULT 0 COMMENT '本周浏览',
    `month_view_count` INT DEFAULT 0 COMMENT '本月浏览',

    -- 互动统计
    `today_like_count` INT DEFAULT 0 COMMENT '今日点赞',
    `today_comment_count` INT DEFAULT 0 COMMENT '今日评论',
    `today_share_count` INT DEFAULT 0 COMMENT '今日分享',

    -- 来源统计（JSON）
    `source_stats` JSON DEFAULT NULL COMMENT '流量来源统计',
    `device_stats` JSON DEFAULT NULL COMMENT '设备统计',
    `region_stats` JSON DEFAULT NULL COMMENT '地域统计',

    -- 阅读完成率
    `avg_read_progress` DECIMAL(5,2) DEFAULT 0 COMMENT '平均阅读进度（%）',
    `avg_stay_time` INT DEFAULT 0 COMMENT '平均停留时间（秒）',

    `update_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章统计详情';

-- ==================== 评论模块（10张表） ====================

-- 1. 评论主表（统一评论表）
DROP TABLE IF EXISTS `comment_v3`;
CREATE TABLE `comment_v3` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '评论ID',

    -- 所属对象
    `target_type` TINYINT NOT NULL COMMENT '评论对象类型：1-文章 2-视频 3-问答 4-动态',
    `target_id` BIGINT NOT NULL COMMENT '评论对象ID',

    -- 用户信息
    `user_id` BIGINT NOT NULL COMMENT '评论者ID',
    `username` VARCHAR(100) DEFAULT NULL COMMENT '用户名（冗余）',
    `user_avatar` VARCHAR(500) DEFAULT NULL COMMENT '用户头像（冗余）',

    -- 评论结构（核心！）
    `parent_id` BIGINT DEFAULT 0 COMMENT '父评论ID（0表示根评论）',
    `root_id` BIGINT DEFAULT 0 COMMENT '根评论ID（方便查询整个评论树）',
    `reply_to_user_id` BIGINT DEFAULT 0 COMMENT '回复的用户ID',
    `reply_to_comment_id` BIGINT DEFAULT 0 COMMENT '回复的评论ID',

    -- 楼层信息（关键！）
    `floor_number` INT DEFAULT 0 COMMENT '楼层号（1楼、2楼...）',
    `sub_floor_number` INT DEFAULT 0 COMMENT '子楼层号（1-1、1-2...）',
    `reply_chain` VARCHAR(1000) DEFAULT NULL COMMENT '回复链路（JSON数组：[id1, id2, id3]）',
    `depth` INT DEFAULT 0 COMMENT '评论深度（0-根评论 1-一级回复 2-二级回复...）',

    -- 评论内容
    `content` TEXT NOT NULL COMMENT '评论内容',
    `content_type` TINYINT DEFAULT 1 COMMENT '内容类型：1-文本 2-图片 3-表情包',
    `images` VARCHAR(1000) DEFAULT NULL COMMENT '图片URL（JSON数组）',
    `at_user_ids` VARCHAR(500) DEFAULT NULL COMMENT '@的用户ID列表（JSON）',

    -- 评论状态
    `status` TINYINT DEFAULT 1 COMMENT '状态：1-正常 2-审核中 3-已删除 4-已折叠 5-已屏蔽',
    `is_pinned` BOOLEAN DEFAULT FALSE COMMENT '是否置顶（UP主置顶）',
    `is_author` BOOLEAN DEFAULT FALSE COMMENT '是否作者评论',
    `is_hot` BOOLEAN DEFAULT FALSE COMMENT '是否热评',
    `is_featured` BOOLEAN DEFAULT FALSE COMMENT '是否精选评论',

    -- 互动数据
    `like_count` INT DEFAULT 0 COMMENT '点赞数',
    `dislike_count` INT DEFAULT 0 COMMENT '踩数',
    `reply_count` INT DEFAULT 0 COMMENT '回复数（直接回复数）',
    `total_reply_count` INT DEFAULT 0 COMMENT '总回复数（包含子回复）',

    -- 热度计算
    `hot_score` DECIMAL(10,2) DEFAULT 0 COMMENT '热度分数（用于排序）',
    `quality_score` DECIMAL(5,2) DEFAULT 0 COMMENT '质量分数',

    -- IP与设备
    `ip` VARCHAR(50) DEFAULT NULL COMMENT '评论IP',
    `ip_location` VARCHAR(100) DEFAULT NULL COMMENT 'IP地理位置',
    `device_type` VARCHAR(50) DEFAULT NULL COMMENT '设备类型',
    `user_agent` TEXT DEFAULT NULL COMMENT 'User Agent',

    -- 审核相关
    `audit_status` TINYINT DEFAULT 0 COMMENT '审核状态：0-待审核 1-通过 2-不通过',
    `audit_reason` VARCHAR(200) DEFAULT NULL COMMENT '审核原因',
    `risk_level` TINYINT DEFAULT 0 COMMENT '风险等级：0-正常 1-低风险 2-中风险 3-高风险',

    -- 时间戳
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '评论时间',
    `update_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `delete_time` TIMESTAMP NULL DEFAULT NULL COMMENT '删除时间',

    -- 扩展字段
    `ext_info` JSON DEFAULT NULL COMMENT '扩展信息',

    INDEX `idx_target` (`target_type`, `target_id`, `status`, `create_time` DESC),
    INDEX `idx_user` (`user_id`, `create_time` DESC),
    INDEX `idx_root` (`root_id`, `floor_number`, `sub_floor_number`),
    INDEX `idx_parent` (`parent_id`, `create_time`),
    INDEX `idx_hot` (`target_id`, `is_hot`, `hot_score` DESC),
    INDEX `idx_floor` (`target_id`, `floor_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论主表（企业级）';

-- 2. 评论互动表（点赞/踩）
DROP TABLE IF EXISTS `comment_interaction`;
CREATE TABLE `comment_interaction` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `comment_id` BIGINT NOT NULL COMMENT '评论ID',
    `user_id` BIGINT NOT NULL COMMENT '用户ID',
    `action_type` TINYINT NOT NULL COMMENT '互动类型：1-点赞 2-踩 3-举报',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_comment_user_action` (`comment_id`, `user_id`, `action_type`),
    INDEX `idx_user` (`user_id`, `action_type`, `create_time` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论互动表';

-- 3. 评论热榜表（热评缓存）
DROP TABLE IF EXISTS `comment_hot_list`;
CREATE TABLE `comment_hot_list` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `target_type` TINYINT NOT NULL COMMENT '对象类型',
    `target_id` BIGINT NOT NULL COMMENT '对象ID',
    `comment_id` BIGINT NOT NULL COMMENT '评论ID',

    -- 热评信息（冗余，减少JOIN）
    `user_id` BIGINT DEFAULT NULL COMMENT '评论者ID',
    `username` VARCHAR(100) DEFAULT NULL COMMENT '用户名',
    `content` TEXT DEFAULT NULL COMMENT '评论内容',
    `like_count` INT DEFAULT NULL COMMENT '点赞数',

    `rank_position` INT DEFAULT NULL COMMENT '排名位置',
    `hot_score` DECIMAL(10,2) DEFAULT NULL COMMENT '热度分数',

    `update_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY `uk_target_comment` (`target_type`, `target_id`, `comment_id`),
    INDEX `idx_target_rank` (`target_type`, `target_id`, `rank_position`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='热评榜单';

-- 4. 评论举报表
DROP TABLE IF EXISTS `comment_report`;
CREATE TABLE `comment_report` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `comment_id` BIGINT NOT NULL COMMENT '评论ID',
    `reporter_user_id` BIGINT NOT NULL COMMENT '举报人ID',

    `reason_type` TINYINT DEFAULT NULL COMMENT '举报原因：1-垃圾广告 2-色情低俗 3-政治敏感 4-人身攻击 5-造谣传谣',
    `reason_desc` VARCHAR(500) DEFAULT NULL COMMENT '举报详情',

    `status` TINYINT DEFAULT 0 COMMENT '处理状态：0-待处理 1-已处理-成立 2-已处理-不成立',
    `handle_result` VARCHAR(200) DEFAULT NULL COMMENT '处理结果',
    `handle_user_id` BIGINT DEFAULT NULL COMMENT '处理人ID',
    `handle_time` TIMESTAMP NULL DEFAULT NULL COMMENT '处理时间',

    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_comment` (`comment_id`, `status`),
    INDEX `idx_status` (`status`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论举报表';

-- 5. 评论盖楼表（连续评论）
DROP TABLE IF EXISTS `comment_floor_building`;
CREATE TABLE `comment_floor_building` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `target_type` TINYINT NOT NULL COMMENT '对象类型',
    `target_id` BIGINT NOT NULL COMMENT '对象ID',
    `user_id` BIGINT NOT NULL COMMENT '用户ID',

    `comment_ids` VARCHAR(1000) DEFAULT NULL COMMENT '盖楼的评论ID列表（JSON）',
    `floor_count` INT DEFAULT 0 COMMENT '盖楼数量',

    `first_comment_time` TIMESTAMP NULL DEFAULT NULL COMMENT '首次评论时间',
    `last_comment_time` TIMESTAMP NULL DEFAULT NULL COMMENT '最后评论时间',

    INDEX `idx_target_user` (`target_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论盖楼记录';

-- 6. UP主追评表（作者对评论的补充）
DROP TABLE IF EXISTS `comment_author_reply`;
CREATE TABLE `comment_author_reply` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `comment_id` BIGINT NOT NULL COMMENT '原评论ID',
    `author_user_id` BIGINT NOT NULL COMMENT '作者ID',

    `content` TEXT NOT NULL COMMENT '追评��容',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX `idx_comment` (`comment_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='UP主追评';

-- 7. 评论表情包表（神评配图）
DROP TABLE IF EXISTS `comment_emotion`;
CREATE TABLE `comment_emotion` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL COMMENT '表情名称',
    `image_url` VARCHAR(500) NOT NULL COMMENT '表情图片URL',
    `category` VARCHAR(50) DEFAULT NULL COMMENT '分类',
    `use_count` BIGINT DEFAULT 0 COMMENT '使用次数',
    `is_hot` BOOLEAN DEFAULT FALSE COMMENT '是否热门',
    `sort_order` INT DEFAULT 0 COMMENT '排序',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表情包';

-- 8. 评论敏感词库（内容审核）
DROP TABLE IF EXISTS `comment_sensitive_word`;
CREATE TABLE `comment_sensitive_word` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `word` VARCHAR(100) NOT NULL UNIQUE COMMENT '敏感词',
    `level` TINYINT DEFAULT 1 COMMENT '等级：1-一般 2-严重 3-非常严重',
    `action` TINYINT DEFAULT 1 COMMENT '处理动作：1-替换 2-拦截 3-人工审核',
    `replacement` VARCHAR(100) DEFAULT NULL COMMENT '替换词',
    `category` VARCHAR(50) DEFAULT NULL COMMENT '分类：政治/色情/广告',
    `is_enabled` BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论敏感词库';

-- 9. 评论统计表（分离热数据）
DROP TABLE IF EXISTS `comment_stats`;
CREATE TABLE `comment_stats` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `target_type` TINYINT NOT NULL COMMENT '对象类型',
    `target_id` BIGINT NOT NULL COMMENT '对象ID',

    `total_comment_count` BIGINT DEFAULT 0 COMMENT '总评论数',
    `today_comment_count` INT DEFAULT 0 COMMENT '今日评论数',
    `root_comment_count` INT DEFAULT 0 COMMENT '根评论数',

    `avg_comment_length` DECIMAL(10,2) DEFAULT NULL COMMENT '平均评论长度',
    `hot_comment_count` INT DEFAULT 0 COMMENT '热评数量',

    `last_comment_time` TIMESTAMP NULL DEFAULT NULL COMMENT '最后评论时间',
    `update_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY `uk_target` (`target_type`, `target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论统计表';

-- 10. 评论折叠规则表（智能折叠低质评论）
DROP TABLE IF EXISTS `comment_fold_rule`;
CREATE TABLE `comment_fold_rule` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `rule_name` VARCHAR(100) DEFAULT NULL COMMENT '规则名称',
    `rule_type` TINYINT DEFAULT NULL COMMENT '规则类型：1-关键词 2-低赞 3-举报数 4-用户等级',
    `rule_config` JSON DEFAULT NULL COMMENT '规则配置',
    `is_enabled` BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    `create_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论折叠规则';

-- ==================== 初始化敏感词数据 ====================
-- 插入一些常见敏感词示例（实际生产环境需要更完整的词库）
INSERT INTO `comment_sensitive_word` (`word`, `level`, `action`, `replacement`, `category`, `is_enabled`) VALUES
('傻逼', 3, 1, '**', '脏话', TRUE),
('fuck', 3, 1, '****', '脏话', TRUE),
('垃圾', 2, 1, '**', '脏话', TRUE),
('广告', 2, 3, NULL, '广告', TRUE),
('加微信', 2, 2, NULL, '广告', TRUE),
('刷单', 2, 2, NULL, '广告', TRUE);

SET FOREIGN_KEY_CHECKS = 1;

-- 执行完成提示
SELECT '✅ 数据库初始化完成！共创建22张表。' AS message;
