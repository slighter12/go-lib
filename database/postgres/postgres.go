package postgres

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const (
	_defaultMaxOpenConns     = 25
	_defaultMaxIdleConns     = 25
	_defaultMaxLifeTime      = 5 * time.Minute
	_defaultSlowSQLThreshold = 200 * time.Millisecond
)

// DBConn 整合了主庫和從庫的配置
type DBConn struct {
	// 主庫配置
	Master ConnectionConfig `json:"master" yaml:"master"`

	// 從庫配置列表
	Replicas []ConnectionConfig `json:"replicas" yaml:"replicas"`

	// 連接池配置
	MaxIdleConns    int           `json:"maxIdleConns" yaml:"maxIdleConns"`
	MaxOpenConns    int           `json:"maxOpenConns" yaml:"maxOpenConns"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime" yaml:"connMaxLifetime"`

	// 數據庫名稱
	Database string `json:"database" yaml:"database"`

	// PostgreSQL 特有配置
	Schema     string `json:"schema" yaml:"schema"`
	SearchPath string `json:"searchPath" yaml:"searchPath"`
	SSLMode    string `json:"sslMode" yaml:"sslMode"`

	// PostgreSQL 超時設置
	StatementTimeout                time.Duration `json:"statementTimeout" yaml:"statementTimeout"`                               // statement_timeout
	LockTimeout                     time.Duration `json:"lockTimeout" yaml:"lockTimeout"`                                         // lock_timeout
	IdleInTransactionSessionTimeout time.Duration `json:"idleInTransactionSessionTimeout" yaml:"idleInTransactionSessionTimeout"` // idle_in_transaction_session_timeout

	// pgx 特有配置
	ApplicationName   string            `json:"applicationName" yaml:"applicationName"`
	RuntimeParams     map[string]string `json:"runtimeParams" yaml:"runtimeParams"`
	HealthCheckPeriod time.Duration     `json:"healthCheckPeriod" yaml:"healthCheckPeriod"`
}

// ConnectionConfig 定義單個數據庫連接的配置
type ConnectionConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	UserName string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// DSN 生成PostgreSQL連接字符串
func (c *ConnectionConfig) DSN(cfg *DBConn) string {
	sslMode := "disable"
	if cfg.SSLMode != "" {
		sslMode = cfg.SSLMode
	}

	searchPath := "public"
	if cfg.SearchPath != "" {
		searchPath = cfg.SearchPath
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		c.Host,
		c.Port,
		c.UserName,
		c.Password,
		cfg.Database,
		sslMode,
		searchPath,
	)

	// 添加超時設置
	if cfg.StatementTimeout > 0 {
		dsn += fmt.Sprintf(" statement_timeout=%d", cfg.StatementTimeout.Milliseconds())
	}
	if cfg.LockTimeout > 0 {
		dsn += fmt.Sprintf(" lock_timeout=%d", cfg.LockTimeout.Milliseconds())
	}
	if cfg.IdleInTransactionSessionTimeout > 0 {
		dsn += fmt.Sprintf(" idle_in_transaction_session_timeout=%d", cfg.IdleInTransactionSessionTimeout.Milliseconds())
	}

	// 添加 pgx 特有參數
	if cfg.ApplicationName != "" {
		dsn += fmt.Sprintf(" application_name=%s", cfg.ApplicationName)
	}

	// 添加運行時參數
	for key, value := range cfg.RuntimeParams {
		dsn += fmt.Sprintf(" %s=%s", key, value)
	}

	return dsn
}

func setupConnPool(db *gorm.DB, conn *DBConn) error {
	sqlDB, err := db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get underlying DB")
	}

	maxIdleConns := _defaultMaxIdleConns
	if conn.MaxIdleConns > 0 {
		maxIdleConns = conn.MaxIdleConns
	}

	maxOpenConns := _defaultMaxOpenConns
	if conn.MaxOpenConns > 0 {
		maxOpenConns = conn.MaxOpenConns
	}

	maxLifeTime := _defaultMaxLifeTime
	if conn.ConnMaxLifetime > 0 {
		maxLifeTime = conn.ConnMaxLifetime
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(maxLifeTime)

	return nil
}

// New 創建一個新的 PostgreSQL 數據庫連接
func New(conn *DBConn) (*gorm.DB, error) {
	// 創建主庫連接
	masterConfig, err := pgxpool.ParseConfig(conn.Master.DSN(conn))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse master config")
	}

	// 設置 pgx 特有配置
	if conn.HealthCheckPeriod > 0 {
		masterConfig.HealthCheckPeriod = conn.HealthCheckPeriod
	}

	masterDB := stdlib.OpenDB(*masterConfig.ConnConfig)
	dbBase, err := gorm.Open(postgres.New(postgres.Config{
		Conn: masterDB,
	}), &gorm.Config{})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create master connection")
	}

	// 如果有從庫配置，設置讀寫分離
	if len(conn.Replicas) > 0 {
		var replicas []gorm.Dialector
		for _, replica := range conn.Replicas {
			replicaConfig, err := pgxpool.ParseConfig(replica.DSN(conn))
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse replica config")
			}

			// 設置 pgx 特有配置
			if conn.HealthCheckPeriod > 0 {
				replicaConfig.HealthCheckPeriod = conn.HealthCheckPeriod
			}

			replicaDB := stdlib.OpenDB(*replicaConfig.ConnConfig)
			replicas = append(replicas, postgres.New(postgres.Config{
				Conn: replicaDB,
			}))
		}

		// 註冊 dbresolver 插件
		err = dbBase.Use(dbresolver.Register(dbresolver.Config{
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		}).SetConnMaxIdleTime(time.Hour))

		if err != nil {
			return nil, errors.Wrap(err, "failed to register dbresolver")
		}
	}

	if err := setupConnPool(dbBase, conn); err != nil {
		return nil, err
	}

	return dbBase, nil
}
