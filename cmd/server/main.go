package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/growth-tracker-pro/backend/internal/config"
	"github.com/growth-tracker-pro/backend/internal/handler"
	"github.com/growth-tracker-pro/backend/internal/models"
	"github.com/growth-tracker-pro/backend/internal/repository"
	"github.com/growth-tracker-pro/backend/internal/service"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 加载配置
	cfg := config.LoadDefault()

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 初始化Redis
	redisClient := initRedis(cfg)

	// 初始化仓储
	repo := repository.NewMySQLRepository(db)
	cache := repository.NewRedisCache(redisClient)

	// 初始化服务
	svc := service.NewService(repo, cache, cfg)

	// 初始化处理器
	h := handler.NewHandler(svc)

	// 创建Gin引擎
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 注册路由
	h.RegisterRoutes(r)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         cfg.Server.GetAddr(),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("服务器启动在 %s", cfg.Server.GetAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}

	log.Println("服务器已关闭")
}

// initDatabase 初始化数据库
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.GetDSN()

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Child{},
		&models.Record{},
		&models.Subscription{},
		&models.Family{},
		&models.FamilyMember{},
		&models.FamilyChild{},
		&models.LabReport{},
		&models.AIConversation{},
	)
}

// initRedis 初始化Redis
func initRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}
