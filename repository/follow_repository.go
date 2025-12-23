package repository

import (
	"astronomer-gin/model"
	"gorm.io/gorm"
)

type FollowRepository interface {
	// 关注操作
	Follow(userID, followUserID string) error
	Unfollow(userID, followUserID string) error
	IsFollowing(userID, followUserID string) bool

	// 列表查询
	GetFollowers(userID string, page, pageSize int) ([]string, int64, error)
	GetFollowing(userID string, page, pageSize int) ([]string, int64, error)

	// 统计
	GetFollowerCount(userID string) (int64, error)
	GetFollowingCount(userID string) (int64, error)

	// 好友相关（互关即好友，类似抖音）
	IsFriend(userID, targetUserID string) bool
	GetFriendsList(userID string, page, pageSize int) ([]string, int64, error)
	GetFriendsCount(userID string) (int64, error)

	// 拉黑操作
	Block(userID, blockUserID string) error
	Unblock(userID, blockUserID string) error
	IsBlocked(userID, blockUserID string) bool
	GetBlockList(userID string) ([]string, error)
}

type followRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepository{db: db}
}

// Follow 关注用户
func (r *followRepository) Follow(userID, followUserID string) error {
	follow := &model.UserFollow{
		UserID:       userID,
		FollowUserID: followUserID,
	}
	return r.db.Create(follow).Error
}

// Unfollow 取消关注
func (r *followRepository) Unfollow(userID, followUserID string) error {
	return r.db.Where("user_id = ? AND follow_user_id = ?", userID, followUserID).
		Delete(&model.UserFollow{}).Error
}

// IsFollowing 检查是否已关注
func (r *followRepository) IsFollowing(userID, followUserID string) bool {
	var count int64
	r.db.Model(&model.UserFollow{}).
		Where("user_id = ? AND follow_user_id = ?", userID, followUserID).
		Count(&count)
	return count > 0
}

// GetFollowers 获取粉丝列表（分页）
func (r *followRepository) GetFollowers(userID string, page, pageSize int) ([]string, int64, error) {
	var total int64
	var follows []model.UserFollow

	// 获取总数
	if err := r.db.Model(&model.UserFollow{}).
		Where("follow_user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Where("follow_user_id = ?", userID).
		Order("create_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	// 提取用户ID
	userIDs := make([]string, len(follows))
	for i, f := range follows {
		userIDs[i] = f.UserID
	}

	return userIDs, total, nil
}

// GetFollowing 获取关注列表（分页）
func (r *followRepository) GetFollowing(userID string, page, pageSize int) ([]string, int64, error) {
	var total int64
	var follows []model.UserFollow

	// 获取总数
	if err := r.db.Model(&model.UserFollow{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).
		Order("create_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	// 提取用户ID
	userIDs := make([]string, len(follows))
	for i, f := range follows {
		userIDs[i] = f.FollowUserID
	}

	return userIDs, total, nil
}

// GetFollowerCount 获取粉丝数
func (r *followRepository) GetFollowerCount(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserFollow{}).
		Where("follow_user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetFollowingCount 获取关注数
func (r *followRepository) GetFollowingCount(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserFollow{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// Block 拉黑用户
func (r *followRepository) Block(userID, blockUserID string) error {
	block := &model.UserBlock{
		UserID:      userID,
		BlockUserID: blockUserID,
	}
	return r.db.Create(block).Error
}

// Unblock 取消拉黑
func (r *followRepository) Unblock(userID, blockUserID string) error {
	return r.db.Where("user_id = ? AND block_user_id = ?", userID, blockUserID).
		Delete(&model.UserBlock{}).Error
}

// IsBlocked 检查是否已拉黑
func (r *followRepository) IsBlocked(userID, blockUserID string) bool {
	var count int64
	r.db.Model(&model.UserBlock{}).
		Where("user_id = ? AND block_user_id = ?", userID, blockUserID).
		Count(&count)
	return count > 0
}

// GetBlockList 获取拉黑列表
func (r *followRepository) GetBlockList(userID string) ([]string, error) {
	var blocks []model.UserBlock
	if err := r.db.Where("user_id = ?", userID).
		Order("create_time DESC").
		Find(&blocks).Error; err != nil {
		return nil, err
	}

	userIDs := make([]string, len(blocks))
	for i, b := range blocks {
		userIDs[i] = b.BlockUserID
	}

	return userIDs, nil
}

// ==================== 好友相关方法 ====================

// IsFriend 检查是否为好友（互相关注）
func (r *followRepository) IsFriend(userID, targetUserID string) bool {
	// 检查双方是否互相关注
	return r.IsFollowing(userID, targetUserID) && r.IsFollowing(targetUserID, userID)
}

// GetFriendsList 获取好友列表（互相关注的用户）
func (r *followRepository) GetFriendsList(userID string, page, pageSize int) ([]string, int64, error) {
	// 查询我关注的人
	var myFollowing []model.UserFollow
	if err := r.db.Where("user_id = ?", userID).
		Find(&myFollowing).Error; err != nil {
		return nil, 0, err
	}

	// 提取我关注的所有用户ID
	followingIDs := make([]string, len(myFollowing))
	for i, f := range myFollowing {
		followingIDs[i] = f.FollowUserID
	}

	if len(followingIDs) == 0 {
		return []string{}, 0, nil
	}

	// 查询在我关注的人中，也关注了我的人（即好友）
	var mutualFollows []model.UserFollow
	query := r.db.Model(&model.UserFollow{}).
		Where("user_id IN ? AND follow_user_id = ?", followingIDs, userID)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&mutualFollows).Error; err != nil {
		return nil, 0, err
	}

	// 提取好友ID
	friendIDs := make([]string, len(mutualFollows))
	for i, f := range mutualFollows {
		friendIDs[i] = f.UserID
	}

	return friendIDs, total, nil
}

// GetFriendsCount 获取好友数量
func (r *followRepository) GetFriendsCount(userID string) (int64, error) {
	// 查询我关注的人
	var myFollowing []model.UserFollow
	if err := r.db.Where("user_id = ?", userID).
		Find(&myFollowing).Error; err != nil {
		return 0, err
	}

	// 提取我关注的所有用户ID
	followingIDs := make([]string, len(myFollowing))
	for i, f := range myFollowing {
		followingIDs[i] = f.FollowUserID
	}

	if len(followingIDs) == 0 {
		return 0, nil
	}

	// 统计在我关注的人中，也关注了我的人数
	var count int64
	err := r.db.Model(&model.UserFollow{}).
		Where("user_id IN ? AND follow_user_id = ?", followingIDs, userID).
		Count(&count).Error

	return count, err
}
