-- ==================================================================================
-- UserID 从 INT/BIGINT 迁移到 UUID (VARCHAR(36))
-- 数据库迁移脚本
--
-- ⚠️ 重要提示：
-- 1. 执行前务必备份数据库！！！
-- 2. 建议在非业务高峰期执行
-- 3. 建议分批执行，测试无误后再继续
-- 4. 预计执行时间：视数据量而定（建议先在测试环境验证）
-- ==================================================================================

SET FOREIGN_KEY_CHECKS = 0;
SET SQL_SAFE_UPDATES = 0;

-- ==================================================================================
-- 第一步：为 user 表主键生成 UUID
-- ==================================================================================

-- 1.1 添加临时 UUID 列
ALTER TABLE `user` ADD COLUMN `id_uuid` VARCHAR(36) DEFAULT NULL AFTER `id`;

-- 1.2 为现有用户生成 UUID（根据实际情况调整 UUID 生成逻辑）
UPDATE `user` SET `id_uuid` = UUID();

-- 1.3 确保 UUID 不为空
ALTER TABLE `user` MODIFY COLUMN `id_uuid` VARCHAR(36) NOT NULL;

-- 1.4 创建 UUID 索引
CREATE UNIQUE INDEX `uk_id_uuid` ON `user`(`id_uuid`);

-- 1.5 创建映射临时表（用于后续关联查询）
CREATE TABLE IF NOT EXISTS `user_id_mapping` (
    `old_id` INT NOT NULL,
    `new_uuid` VARCHAR(36) NOT NULL,
    PRIMARY KEY (`old_id`),
    UNIQUE KEY `uk_new_uuid` (`new_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User ID映射表（迁移临时表）';

-- 1.6 填充映射表
INSERT INTO `user_id_mapping` (`old_id`, `new_uuid`)
SELECT `id`, `id_uuid` FROM `user`;


-- ==================================================================================
-- 第二步：更新所有外键表的 user_id 字段
-- ==================================================================================

-- 2.1 article 表
ALTER TABLE `article` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `article` a
INNER JOIN `user_id_mapping` m ON a.`user_id` = CAST(m.`old_id` AS CHAR)
SET a.`user_id_new` = m.`new_uuid`;
-- 确认更新成功后删除旧列
ALTER TABLE `article` DROP COLUMN `user_id`;
ALTER TABLE `article` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `article` ADD INDEX `idx_user_id` (`user_id`);

-- 2.2 article_v3 表
ALTER TABLE `article_v3` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
ALTER TABLE `article_v3` ADD COLUMN `audit_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `audit_user_id`;

UPDATE `article_v3` a
INNER JOIN `user_id_mapping` m ON a.`user_id` = m.`new_uuid`
SET a.`user_id_new` = m.`new_uuid`;

UPDATE `article_v3` a
INNER JOIN `user_id_mapping` m ON a.`audit_user_id` = m.`new_uuid`
SET a.`audit_user_id_new` = m.`new_uuid`
WHERE a.`audit_user_id` IS NOT NULL AND a.`audit_user_id` != '';

ALTER TABLE `article_v3` DROP COLUMN `user_id`;
ALTER TABLE `article_v3` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `article_v3` ADD INDEX `idx_user_id` (`user_id`);

ALTER TABLE `article_v3` DROP COLUMN `audit_user_id`;
ALTER TABLE `article_v3` CHANGE COLUMN `audit_user_id_new` `audit_user_id` VARCHAR(36) DEFAULT NULL;

-- 2.3 article_draft 表
ALTER TABLE `article_draft` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `article_draft` a
INNER JOIN `user_id_mapping` m ON a.`user_id` = m.`new_uuid`
SET a.`user_id_new` = m.`new_uuid`;
ALTER TABLE `article_draft` DROP COLUMN `user_id`;
ALTER TABLE `article_draft` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `article_draft` ADD INDEX `idx_user` (`user_id`);

-- 2.4 article_column 表
ALTER TABLE `article_column` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `article_column` a
INNER JOIN `user_id_mapping` m ON a.`user_id` = m.`new_uuid`
SET a.`user_id_new` = m.`new_uuid`;
ALTER TABLE `article_column` DROP COLUMN `user_id`;
ALTER TABLE `article_column` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `article_column` ADD INDEX `idx_user` (`user_id`);

-- 2.5 article_history 表
ALTER TABLE `article_history` ADD COLUMN `operator_id_new` VARCHAR(36) DEFAULT NULL AFTER `operator_id`;
UPDATE `article_history` a
INNER JOIN `user_id_mapping` m ON a.`operator_id` = m.`new_uuid`
SET a.`operator_id_new` = m.`new_uuid`
WHERE a.`operator_id` IS NOT NULL AND a.`operator_id` != '';
ALTER TABLE `article_history` DROP COLUMN `operator_id`;
ALTER TABLE `article_history` CHANGE COLUMN `operator_id_new` `operator_id` VARCHAR(36) DEFAULT NULL;

-- 2.6 topic 表
ALTER TABLE `topic` ADD COLUMN `creator_id_new` VARCHAR(36) DEFAULT NULL AFTER `creator_id`;
UPDATE `topic` t
INNER JOIN `user_id_mapping` m ON t.`creator_id` = m.`new_uuid`
SET t.`creator_id_new` = m.`new_uuid`
WHERE t.`creator_id` IS NOT NULL AND t.`creator_id` != '';
ALTER TABLE `topic` DROP COLUMN `creator_id`;
ALTER TABLE `topic` CHANGE COLUMN `creator_id_new` `creator_id` VARCHAR(36) DEFAULT NULL;

-- 2.7 comment_parent 表
ALTER TABLE `comment_parent` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `comment_parent` c
INNER JOIN `user_id_mapping` m ON c.`user_id` = m.`new_uuid`
SET c.`user_id_new` = m.`new_uuid`;
ALTER TABLE `comment_parent` DROP COLUMN `user_id`;
ALTER TABLE `comment_parent` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;

-- 2.8 comment_sub_two 表
ALTER TABLE `comment_sub_two` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
ALTER TABLE `comment_sub_two` ADD COLUMN `to_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `to_user_id`;

UPDATE `comment_sub_two` c
INNER JOIN `user_id_mapping` m ON c.`user_id` = m.`new_uuid`
SET c.`user_id_new` = m.`new_uuid`;

UPDATE `comment_sub_two` c
INNER JOIN `user_id_mapping` m ON c.`to_user_id` = m.`new_uuid`
SET c.`to_user_id_new` = m.`new_uuid`
WHERE c.`to_user_id` IS NOT NULL AND c.`to_user_id` != '';

ALTER TABLE `comment_sub_two` DROP COLUMN `user_id`;
ALTER TABLE `comment_sub_two` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;

ALTER TABLE `comment_sub_two` DROP COLUMN `to_user_id`;
ALTER TABLE `comment_sub_two` CHANGE COLUMN `to_user_id_new` `to_user_id` VARCHAR(36) DEFAULT NULL;

-- 2.9 comment_parent_like 表
ALTER TABLE `comment_parent_like` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `comment_parent_like` c
INNER JOIN `user_id_mapping` m ON c.`user_id` = m.`new_uuid`
SET c.`user_id_new` = m.`new_uuid`;
ALTER TABLE `comment_parent_like` DROP COLUMN `user_id`;
ALTER TABLE `comment_parent_like` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;

-- 2.10 comment_sub_two_like 表
ALTER TABLE `comment_sub_two_like` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `comment_sub_two_like` c
INNER JOIN `user_id_mapping` m ON c.`user_id` = m.`new_uuid`
SET c.`user_id_new` = m.`new_uuid`;
ALTER TABLE `comment_sub_two_like` DROP COLUMN `user_id`;
ALTER TABLE `comment_sub_two_like` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;

-- 2.11 comment_v3 表
ALTER TABLE `comment_v3` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
ALTER TABLE `comment_v3` ADD COLUMN `reply_to_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `reply_to_user_id`;

UPDATE `comment_v3` c
INNER JOIN `user_id_mapping` m ON c.`user_id` = m.`new_uuid`
SET c.`user_id_new` = m.`new_uuid`;

UPDATE `comment_v3` c
INNER JOIN `user_id_mapping` m ON c.`reply_to_user_id` = m.`new_uuid`
SET c.`reply_to_user_id_new` = m.`new_uuid`
WHERE c.`reply_to_user_id` IS NOT NULL AND c.`reply_to_user_id` != '';

ALTER TABLE `comment_v3` DROP COLUMN `user_id`;
ALTER TABLE `comment_v3` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `comment_v3` ADD INDEX `idx_user` (`user_id`);

ALTER TABLE `comment_v3` DROP COLUMN `reply_to_user_id`;
ALTER TABLE `comment_v3` CHANGE COLUMN `reply_to_user_id_new` `reply_to_user_id` VARCHAR(36) DEFAULT '';

-- 2.12 comment_interaction 表
ALTER TABLE `comment_interaction` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `comment_interaction` c
INNER JOIN `user_id_mapping` m ON c.`user_id` = m.`new_uuid`
SET c.`user_id_new` = m.`new_uuid`;

-- 删除旧的唯一索引
DROP INDEX `uk_comment_user_action` ON `comment_interaction`;
ALTER TABLE `comment_interaction` DROP COLUMN `user_id`;
ALTER TABLE `comment_interaction` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;

-- 重建唯一索引
CREATE UNIQUE INDEX `uk_comment_user_action` ON `comment_interaction`(`comment_id`, `user_id`, `action_type`);
CREATE INDEX `idx_user` ON `comment_interaction`(`user_id`);

-- 2.13 comment_hot_list 表
ALTER TABLE `comment_hot_list` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `comment_hot_list` c
INNER JOIN `user_id_mapping` m ON c.`user_id` = m.`new_uuid`
SET c.`user_id_new` = m.`new_uuid`
WHERE c.`user_id` IS NOT NULL AND c.`user_id` != '';
ALTER TABLE `comment_hot_list` DROP COLUMN `user_id`;
ALTER TABLE `comment_hot_list` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) DEFAULT NULL;

-- 2.14 comment_report 表
ALTER TABLE `comment_report` ADD COLUMN `reporter_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `reporter_user_id`;
ALTER TABLE `comment_report` ADD COLUMN `handle_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `handle_user_id`;

UPDATE `comment_report` c
INNER JOIN `user_id_mapping` m ON c.`reporter_user_id` = m.`new_uuid`
SET c.`reporter_user_id_new` = m.`new_uuid`;

UPDATE `comment_report` c
INNER JOIN `user_id_mapping` m ON c.`handle_user_id` = m.`new_uuid`
SET c.`handle_user_id_new` = m.`new_uuid`
WHERE c.`handle_user_id` IS NOT NULL AND c.`handle_user_id` != '';

ALTER TABLE `comment_report` DROP COLUMN `reporter_user_id`;
ALTER TABLE `comment_report` CHANGE COLUMN `reporter_user_id_new` `reporter_user_id` VARCHAR(36) NOT NULL;

ALTER TABLE `comment_report` DROP COLUMN `handle_user_id`;
ALTER TABLE `comment_report` CHANGE COLUMN `handle_user_id_new` `handle_user_id` VARCHAR(36) DEFAULT NULL;

-- 2.15 comment_floor_building 表
ALTER TABLE `comment_floor_building` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `comment_floor_building` c
INNER JOIN `user_id_mapping` m ON c.`user_id` = m.`new_uuid`
SET c.`user_id_new` = m.`new_uuid`;
ALTER TABLE `comment_floor_building` DROP COLUMN `user_id`;
ALTER TABLE `comment_floor_building` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `comment_floor_building` ADD INDEX `idx_target_user` (`target_id`, `user_id`);

-- 2.16 comment_author_reply 表
ALTER TABLE `comment_author_reply` ADD COLUMN `author_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `author_user_id`;
UPDATE `comment_author_reply` c
INNER JOIN `user_id_mapping` m ON c.`author_user_id` = m.`new_uuid`
SET c.`author_user_id_new` = m.`new_uuid`;
ALTER TABLE `comment_author_reply` DROP COLUMN `author_user_id`;
ALTER TABLE `comment_author_reply` CHANGE COLUMN `author_user_id_new` `author_user_id` VARCHAR(36) NOT NULL;

-- 2.17 user_follow 表
ALTER TABLE `user_follow` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
ALTER TABLE `user_follow` ADD COLUMN `follow_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `follow_user_id`;

UPDATE `user_follow` f
INNER JOIN `user_id_mapping` m ON f.`user_id` = m.`new_uuid`
SET f.`user_id_new` = m.`new_uuid`;

UPDATE `user_follow` f
INNER JOIN `user_id_mapping` m ON f.`follow_user_id` = m.`new_uuid`
SET f.`follow_user_id_new` = m.`new_uuid`;

ALTER TABLE `user_follow` DROP COLUMN `user_id`;
ALTER TABLE `user_follow` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `user_follow` ADD INDEX `idx_user_id` (`user_id`);

ALTER TABLE `user_follow` DROP COLUMN `follow_user_id`;
ALTER TABLE `user_follow` CHANGE COLUMN `follow_user_id_new` `follow_user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `user_follow` ADD INDEX `idx_follow_user_id` (`follow_user_id`);

-- 2.18 user_block 表
ALTER TABLE `user_block` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
ALTER TABLE `user_block` ADD COLUMN `block_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `block_user_id`;

UPDATE `user_block` b
INNER JOIN `user_id_mapping` m ON b.`user_id` = m.`new_uuid`
SET b.`user_id_new` = m.`new_uuid`;

UPDATE `user_block` b
INNER JOIN `user_id_mapping` m ON b.`block_user_id` = m.`new_uuid`
SET b.`block_user_id_new` = m.`new_uuid`;

ALTER TABLE `user_block` DROP COLUMN `user_id`;
ALTER TABLE `user_block` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `user_block` ADD INDEX `idx_user_id` (`user_id`);

ALTER TABLE `user_block` DROP COLUMN `block_user_id`;
ALTER TABLE `user_block` CHANGE COLUMN `block_user_id_new` `block_user_id` VARCHAR(36) NOT NULL;

-- 2.19 article_star (like) 表
ALTER TABLE `article_star` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `article_star` a
INNER JOIN `user_id_mapping` m ON a.`user_id` = m.`new_uuid`
SET a.`user_id_new` = m.`new_uuid`;
ALTER TABLE `article_star` DROP COLUMN `user_id`;
ALTER TABLE `article_star` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `article_star` ADD INDEX `idx_user_id` (`user_id`);

-- 2.20 user_favorite 表
ALTER TABLE `user_favorite` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `user_favorite` f
INNER JOIN `user_id_mapping` m ON f.`user_id` = m.`new_uuid`
SET f.`user_id_new` = m.`new_uuid`;
ALTER TABLE `user_favorite` DROP COLUMN `user_id`;
ALTER TABLE `user_favorite` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `user_favorite` ADD INDEX `idx_user_id` (`user_id`);

-- 2.21 user_chat 表
ALTER TABLE `user_chat` ADD COLUMN `from_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `from_user_id`;
ALTER TABLE `user_chat` ADD COLUMN `to_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `to_user_id`;

UPDATE `user_chat` c
INNER JOIN `user_id_mapping` m ON c.`from_user_id` = m.`new_uuid`
SET c.`from_user_id_new` = m.`new_uuid`;

UPDATE `user_chat` c
INNER JOIN `user_id_mapping` m ON c.`to_user_id` = m.`new_uuid`
SET c.`to_user_id_new` = m.`new_uuid`;

ALTER TABLE `user_chat` DROP COLUMN `from_user_id`;
ALTER TABLE `user_chat` CHANGE COLUMN `from_user_id_new` `from_user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `user_chat` ADD INDEX `idx_from_user` (`from_user_id`);

ALTER TABLE `user_chat` DROP COLUMN `to_user_id`;
ALTER TABLE `user_chat` CHANGE COLUMN `to_user_id_new` `to_user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `user_chat` ADD INDEX `idx_to_user` (`to_user_id`);

-- 2.22 notification 表
ALTER TABLE `notification` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
ALTER TABLE `notification` ADD COLUMN `from_user_id_new` VARCHAR(36) DEFAULT NULL AFTER `from_user_id`;

UPDATE `notification` n
INNER JOIN `user_id_mapping` m ON n.`user_id` = m.`new_uuid`
SET n.`user_id_new` = m.`new_uuid`;

UPDATE `notification` n
INNER JOIN `user_id_mapping` m ON n.`from_user_id` = m.`new_uuid`
SET n.`from_user_id_new` = m.`new_uuid`
WHERE n.`from_user_id` IS NOT NULL AND n.`from_user_id` != '';

ALTER TABLE `notification` DROP COLUMN `user_id`;
ALTER TABLE `notification` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `notification` ADD INDEX `idx_user_id` (`user_id`);

ALTER TABLE `notification` DROP COLUMN `from_user_id`;
ALTER TABLE `notification` CHANGE COLUMN `from_user_id_new` `from_user_id` VARCHAR(36) DEFAULT NULL;

-- 2.23 column_subscription 表
ALTER TABLE `column_subscription` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `column_subscription` c
INNER JOIN `user_id_mapping` m ON c.`user_id` = m.`new_uuid`
SET c.`user_id_new` = m.`new_uuid`;
ALTER TABLE `column_subscription` DROP COLUMN `user_id`;
ALTER TABLE `column_subscription` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `column_subscription` ADD INDEX `idx_user_id` (`user_id`);

-- 2.24 topic_follow 表
ALTER TABLE `topic_follow` ADD COLUMN `user_id_new` VARCHAR(36) DEFAULT NULL AFTER `user_id`;
UPDATE `topic_follow` t
INNER JOIN `user_id_mapping` m ON t.`user_id` = m.`new_uuid`
SET t.`user_id_new` = m.`new_uuid`;
ALTER TABLE `topic_follow` DROP COLUMN `user_id`;
ALTER TABLE `topic_follow` CHANGE COLUMN `user_id_new` `user_id` VARCHAR(36) NOT NULL;
ALTER TABLE `topic_follow` ADD INDEX `idx_user_id` (`user_id`);


-- ==================================================================================
-- 第三步：替换 user 表主键
-- ==================================================================================

-- 3.1 删除旧主键
ALTER TABLE `user` DROP PRIMARY KEY;

-- 3.2 删除旧 id 列
ALTER TABLE `user` DROP COLUMN `id`;

-- 3.3 将 id_uuid 重命名为 id
ALTER TABLE `user` CHANGE COLUMN `id_uuid` `id` VARCHAR(36) NOT NULL;

-- 3.4 设置新主键
ALTER TABLE `user` ADD PRIMARY KEY (`id`);

-- 3.5 删除临时创建的唯一索引（如果主键已创建）
-- ALTER TABLE `user` DROP INDEX `uk_id_uuid`;


-- ==================================================================================
-- 第四步：清理工作
-- ==================================================================================

-- 4.1 删除映射临时表（可选：建议保留一段时间以备回滚）
-- DROP TABLE IF EXISTS `user_id_mapping`;

-- 4.2 验证数据完整性
-- SELECT COUNT(*) FROM `user` WHERE `id` IS NULL OR `id` = '';
-- SELECT COUNT(*) FROM `article` WHERE `user_id` IS NULL OR `user_id` = '';
-- SELECT COUNT(*) FROM `user_follow` WHERE `user_id` IS NULL OR `follow_user_id` IS NULL;

-- 4.3 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;
SET SQL_SAFE_UPDATES = 1;


-- ==================================================================================
-- 第五步：验证脚本（可选）
-- ==================================================================================

-- 验证 user 表
-- SELECT id, phone, username FROM `user` LIMIT 10;

-- 验证关联表
-- SELECT a.id, a.title, u.username
-- FROM article a
-- INNER JOIN user u ON a.user_id = u.id
-- LIMIT 10;

-- 验证关注关系
-- SELECT f.id, u1.username AS follower, u2.username AS following
-- FROM user_follow f
-- INNER JOIN user u1 ON f.user_id = u1.id
-- INNER JOIN user u2 ON f.follow_user_id = u2.id
-- LIMIT 10;


-- ==================================================================================
-- 回滚脚本（紧急情况使用）
-- ==================================================================================

/*
-- ⚠️ 仅在迁移失败需要回滚时使用！！！

-- 使用映射表回滚（需要 user_id_mapping 表存在）
SET FOREIGN_KEY_CHECKS = 0;

-- 回滚示例（以 article 表为例）
ALTER TABLE `article` ADD COLUMN `user_id_old` INT DEFAULT NULL;
UPDATE `article` a
INNER JOIN `user_id_mapping` m ON a.`user_id` = m.`new_uuid`
SET a.`user_id_old` = m.`old_id`;
ALTER TABLE `article` DROP COLUMN `user_id`;
ALTER TABLE `article` CHANGE COLUMN `user_id_old` `user_id` INT NOT NULL;

-- ... 依次回滚其他表

-- 回滚 user 表主键
ALTER TABLE `user` ADD COLUMN `id_old` INT AUTO_INCREMENT FIRST, ADD KEY(`id_old`);
UPDATE `user` u
INNER JOIN `user_id_mapping` m ON u.`id` = m.`new_uuid`
SET u.`id_old` = m.`old_id`;
ALTER TABLE `user` DROP PRIMARY KEY;
ALTER TABLE `user` DROP COLUMN `id`;
ALTER TABLE `user` CHANGE COLUMN `id_old` `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY;

SET FOREIGN_KEY_CHECKS = 1;
*/


-- ==================================================================================
-- 执行建议
-- ==================================================================================
/*
1. 分批执行建议：
   - 先执行第一步（生成 UUID 和映射表）
   - 逐个执行第二步中的表更新（每更新 2-3 个表验证一次）
   - 最后执行第三步（替换主键）

2. 监控要点：
   - 执行过程中监控表锁定情况
   - 检查磁盘空间（临时列会占用额外空间）
   - 记录每步执行时间

3. 测试环境验证：
   - 先在测试环境完整执行一遍
   - 验证应用程序功能正常
   - 验证查询性能未受影响

4. 生产环境执行：
   - 选择业务低峰期
   - 提前通知用户系统维护
   - 准备回滚预案
   - 保留 user_id_mapping 表至少 1 周

5. 执行后验证：
   - 运行第五步的验证 SQL
   - 检查应用日志是否有错误
   - 抽查核心功能是否正常
*/
