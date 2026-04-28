package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/growth-tracker-pro-backend/internal/models"
	"gorm.io/gorm"
)

// Repository 仓储接口
type Repository interface {
	// 用户
	GetUserByOpenID(ctx context.Context, openID string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error

	// 宝宝
	GetChildrenByUserID(ctx context.Context, userID string) ([]models.Child, error)
	GetChildByID(ctx context.Context, id string) (*models.Child, error)
	GetActiveChild(ctx context.Context, userID string) (*models.Child, error)
	CreateChild(ctx context.Context, child *models.Child) error
	UpdateChild(ctx context.Context, child *models.Child) error
	DeleteChild(ctx context.Context, id string) error
	SetActiveChild(ctx context.Context, userID, childID string) error

	// 记录
	GetRecordsByChildID(ctx context.Context, childID string, startDate, endDate string, page, pageSize int) ([]models.Record, int64, error)
	GetRecordByID(ctx context.Context, id string) (*models.Record, error)
	CreateRecord(ctx context.Context, record *models.Record) error
	UpdateRecord(ctx context.Context, record *models.Record) error
	DeleteRecord(ctx context.Context, id string) error

	// 订阅
	GetSubscriptionByUserID(ctx context.Context, userID string) (*models.Subscription, error)
	CreateSubscription(ctx context.Context, sub *models.Subscription) error
	UpdateSubscription(ctx context.Context, sub *models.Subscription) error
	IncrementAIUsage(ctx context.Context, userID string) error

	// 家庭
	GetFamilyByID(ctx context.Context, id string) (*models.Family, error)
	GetFamilyByInviteCode(ctx context.Context, code string) (*models.Family, error)
	GetFamilyByUserID(ctx context.Context, userID string) (*models.Family, error)
	CreateFamily(ctx context.Context, family *models.Family) error
	AddFamilyMember(ctx context.Context, member *models.FamilyMember) error
	RemoveFamilyMember(ctx context.Context, familyID, memberID string) error
	UpdateMemberRole(ctx context.Context, memberID, role string) error
	GetFamilyMembers(ctx context.Context, familyID string) ([]models.FamilyMember, error)

	// 化验单
	GetLabReportsByChildID(ctx context.Context, childID string) ([]models.LabReport, error)
	CreateLabReport(ctx context.Context, report *models.LabReport) error

	// AI对话
	GetConversationBySessionID(ctx context.Context, sessionID string) (*models.AIConversation, error)
	CreateConversation(ctx context.Context, conv *models.AIConversation) error
	UpdateConversation(ctx context.Context, conv *models.AIConversation) error
}

// MySQLRepository MySQL仓储实现
type MySQLRepository struct {
	db *gorm.DB
}

// NewMySQLRepository 创建MySQL仓储
func NewMySQLRepository(db *gorm.DB) Repository {
	return &MySQLRepository{db: db}
}

// ========== 用户操作 ==========

func (r *MySQLRepository) GetUserByOpenID(ctx context.Context, openID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("open_id = ?", openID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *MySQLRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Children").Preload("Subscription").First(&user, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *MySQLRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *MySQLRepository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// ========== 宝宝操作 ==========

func (r *MySQLRepository) GetChildrenByUserID(ctx context.Context, userID string) ([]models.Child, error) {
	var children []models.Child
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&children).Error
	return children, err
}

func (r *MySQLRepository) GetChildByID(ctx context.Context, id string) (*models.Child, error) {
	var child models.Child
	err := r.db.WithContext(ctx).First(&child, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &child, nil
}

func (r *MySQLRepository) GetActiveChild(ctx context.Context, userID string) (*models.Child, error) {
	var child models.Child
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).First(&child).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &child, nil
}

func (r *MySQLRepository) CreateChild(ctx context.Context, child *models.Child) error {
	return r.db.WithContext(ctx).Create(child).Error
}

func (r *MySQLRepository) UpdateChild(ctx context.Context, child *models.Child) error {
	return r.db.WithContext(ctx).Save(child).Error
}

func (r *MySQLRepository) DeleteChild(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Child{}, "id = ?", id).Error
}

func (r *MySQLRepository) SetActiveChild(ctx context.Context, userID, childID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先取消所有激活状态
		if err := tx.Model(&models.Child{}).Where("user_id = ?", userID).Update("is_active", false).Error; err != nil {
			return err
		}
		// 设置指定宝宝为激活状态
		return tx.Model(&models.Child{}).Where("id = ? AND user_id = ?", childID, userID).Update("is_active", true).Error
	})
}

// ========== 记录操作 ==========

func (r *MySQLRepository) GetRecordsByChildID(ctx context.Context, childID string, startDate, endDate string, page, pageSize int) ([]models.Record, int64, error) {
	var records []models.Record
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Record{}).Where("child_id = ?", childID)

	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("date DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}

func (r *MySQLRepository) GetRecordByID(ctx context.Context, id string) (*models.Record, error) {
	var record models.Record
	err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *MySQLRepository) CreateRecord(ctx context.Context, record *models.Record) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *MySQLRepository) UpdateRecord(ctx context.Context, record *models.Record) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *MySQLRepository) DeleteRecord(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Record{}, "id = ?", id).Error
}

// ========== 订阅操作 ==========

func (r *MySQLRepository) GetSubscriptionByUserID(ctx context.Context, userID string) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *MySQLRepository) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *MySQLRepository) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *MySQLRepository) IncrementAIUsage(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("user_id = ? AND status = ?", userID, "active").
		Updates(map[string]interface{}{
			"ai_used":   gorm.Expr("ai_used + 1"),
			"updated_at": time.Now(),
		}).Error
}

// ========== 家庭操作 ==========

func (r *MySQLRepository) GetFamilyByID(ctx context.Context, id string) (*models.Family, error) {
	var family models.Family
	err := r.db.WithContext(ctx).Preload("Members").First(&family, "family_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &family, nil
}

func (r *MySQLRepository) GetFamilyByInviteCode(ctx context.Context, code string) (*models.Family, error) {
	var family models.Family
	err := r.db.WithContext(ctx).Preload("Members").First(&family, "invite_code = ?", code).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &family, nil
}

func (r *MySQLRepository) GetFamilyByUserID(ctx context.Context, userID string) (*models.Family, error) {
	var member models.FamilyMember
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.GetFamilyByID(ctx, member.FamilyID)
}

func (r *MySQLRepository) CreateFamily(ctx context.Context, family *models.Family) error {
	return r.db.WithContext(ctx).Create(family).Error
}

func (r *MySQLRepository) AddFamilyMember(ctx context.Context, member *models.FamilyMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *MySQLRepository) RemoveFamilyMember(ctx context.Context, familyID, memberID string) error {
	return r.db.WithContext(ctx).Delete(&models.FamilyMember{}, "family_id = ? AND id = ?", familyID, memberID).Error
}

func (r *MySQLRepository) UpdateMemberRole(ctx context.Context, memberID, role string) error {
	return r.db.WithContext(ctx).Model(&models.FamilyMember{}).Where("id = ?", memberID).Update("role", role).Error
}

func (r *MySQLRepository) GetFamilyMembers(ctx context.Context, familyID string) ([]models.FamilyMember, error) {
	var members []models.FamilyMember
	err := r.db.WithContext(ctx).Where("family_id = ?", familyID).Find(&members).Error
	return members, err
}

// ========== 化验单操作 ==========

func (r *MySQLRepository) GetLabReportsByChildID(ctx context.Context, childID string) ([]models.LabReport, error) {
	var reports []models.LabReport
	err := r.db.WithContext(ctx).Where("child_id = ?", childID).Order("created_at DESC").Find(&reports).Error
	return reports, err
}

func (r *MySQLRepository) CreateLabReport(ctx context.Context, report *models.LabReport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

// ========== AI对话操作 ==========

func (r *MySQLRepository) GetConversationBySessionID(ctx context.Context, sessionID string) (*models.AIConversation, error) {
	var conv models.AIConversation
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&conv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

func (r *MySQLRepository) CreateConversation(ctx context.Context, conv *models.AIConversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *MySQLRepository) UpdateConversation(ctx context.Context, conv *models.AIConversation) error {
	return r.db.WithContext(ctx).Save(conv).Error
}

// ========== Redis缓存 ==========

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Set 存储数据
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取数据
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete 删除数据
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Incr 递增
func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// Expire 设置过期时间
func (c *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// 生成缓存Key
func CacheKey(parts ...string) string {
	return fmt.Sprintf("growth-tracker:%s", join(parts, ":"))
}

func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
