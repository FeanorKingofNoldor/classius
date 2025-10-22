package db

import (
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/classius/server/internal/models"
)

var (
	DB    *gorm.DB
	Redis *redis.Client
)

// InitDB initializes the database connection
func InitDB() *gorm.DB {
	// Build DSN from configuration
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		viper.GetString("database.host"),
		viper.GetString("database.user"),
		viper.GetString("database.password"),
		viper.GetString("database.name"),
		viper.GetInt("database.port"),
	)

	// Configure GORM logger
	gormLogger := logger.Default
	if viper.GetString("environment") == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get underlying sql.DB for connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Database connected successfully")

	DB = db
	return db
}

// InitRedis initializes Redis connection
func InitRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	// Test connection
	_, err := client.Ping(client.Context()).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		return nil
	}

	log.Println("✅ Redis connected successfully")
	Redis = client
	return client
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.Tag{},
		&models.UserBook{},
		&models.ReadingProgress{},
		&models.Annotation{},
		&models.Bookmark{},
		&models.SageConversation{},
		// &models.ReadingGroup{},
		// &models.GroupMember{},
		// &models.Discussion{},
		&models.PublishedNote{},
		&models.NoteOverlay{},
		&models.UserSession{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("✅ Database migrations completed")
	return nil
}

// CloseDB closes database connections
func CloseDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error getting underlying sql.DB: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	} else {
		log.Println("✅ Database connection closed")
	}

	if Redis != nil {
		if err := Redis.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		} else {
			log.Println("✅ Redis connection closed")
		}
	}
}

// RunMigrations runs database migrations (legacy method for main.go compatibility)
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return AutoMigrate(DB)
}