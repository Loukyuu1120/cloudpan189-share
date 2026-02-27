package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/taskengine"
	"go.uber.org/zap"

	"github.com/glebarez/sqlite"
	"github.com/xxcheng123/cloudpan189-share/internal/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func useSQLiteDB(c *configs.Config) (db *gorm.DB, err error) {
	dir := filepath.Dir(c.DBFile)
	if err = os.MkdirAll(dir, 0766); err != nil {
		return
	}

	db, err = gorm.Open(sqlite.Open(c.DBFile), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to open SQLite database")
	}
	db.Exec("PRAGMA journal_mode = WAL;")
	db.Exec("PRAGMA synchronous = NORMAL;")
	db.Exec("PRAGMA busy_timeout = 5000;")

	if err = db.Use(new(TracePlugin)); err != nil {
		return nil, errors.Wrap(err, "failed to register trace plugin")
	}

	return db, nil
}

func useMySqlDB(c *configs.Config) (db *gorm.DB, err error) {
	if c.MySQL == nil {
		return nil, errors.New("MySQL configuration is required when using MySQL database")
	}

	// 构建 MySQL DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQL.User,
		c.MySQL.Pass,
		c.MySQL.Host,
		c.MySQL.Port,
		c.MySQL.DBName,
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to MySQL database")
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	if err = db.Use(new(TracePlugin)); err != nil {
		return nil, errors.Wrap(err, "failed to register trace plugin")
	}

	return db, nil
}

func usePostgresDB(c *configs.Config) (db *gorm.DB, err error) {
	if c.Postgres == nil {
		return nil, errors.New("PostgreSQL configuration is required when using PostgreSQL database")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		c.Postgres.Host,
		c.Postgres.User,
		c.Postgres.Pass,
		c.Postgres.DBName,
		c.Postgres.Port,
		c.Postgres.SSLMode,
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to PostgreSQL database")
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	if err = db.Use(new(TracePlugin)); err != nil {
		return nil, errors.Wrap(err, "failed to register trace plugin")
	}

	return db, nil
}

func connectDB(c *configs.Config) (db *gorm.DB, err error) {
	switch c.DBType {
	case "mysql":
		return useMySqlDB(c)
	case "postgres", "postgresql":
		return usePostgresDB(c)
	case "sqlite":
		return useSQLiteDB(c)
	default:
		return useSQLiteDB(c) // 默认使用 SQLite
	}
}

func assignShared(db *gorm.DB) (err error) {
	var setting = new(models.Setting)
	if err = db.First(setting).Error; err != nil {
		return err
	}

	shared.SaltKey = setting.SaltKey
	shared.BaseURL = setting.BaseURL
	shared.EnableAuth = setting.EnableAuth
	shared.SettingAddition = setting.Addition

	var mediaConfig = new(models.MediaConfig)
	if err = db.First(mediaConfig).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	shared.MediaConfig = mediaConfig

	return nil
}

func initTaskEngine(logger *zap.Logger, cfg *configs.TaskEngineConfig) taskengine.TaskEngine {
	opts := []taskengine.EngineOption{
		taskengine.WithLogger(logger.Named("task_engine")),
	}

	if cfg != nil {
		if cfg.WorkerCount > 0 {
			opts = append(opts, taskengine.EngineOption{
				Options: []taskengine.OptionFunc{
					taskengine.WithWorkerCount(cfg.WorkerCount),
				},
			})
		}
		if cfg.BufferSize > 0 {
			opts = append(opts, taskengine.EngineOption{
				Options: []taskengine.OptionFunc{
					taskengine.WithBufferSize(cfg.BufferSize),
				},
			})
		}
		if cfg.ProcessTimeout > 0 {
			opts = append(opts, taskengine.EngineOption{
				Options: []taskengine.OptionFunc{
					taskengine.WithProcessTimeout(time.Duration(cfg.ProcessTimeout) * time.Second),
				},
			})
		}
		if cfg.MaxRetry >= 0 {
			opts = append(opts, taskengine.EngineOption{
				Options: []taskengine.OptionFunc{
					taskengine.WithMaxRetry(cfg.MaxRetry),
				},
			})
		}
	}

	return taskengine.NewTaskEngine(opts...)
}
