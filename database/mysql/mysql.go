package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
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
}

// ConnectionConfig 定義單個數據庫連接的配置
type ConnectionConfig struct {
	Host     string        `json:"host" yaml:"host"`
	Port     string        `json:"port" yaml:"port"`
	UserName string        `json:"username" yaml:"username"`
	Password string        `json:"password" yaml:"password"`
	Loc      string        `json:"loc" yaml:"loc"`
	Timeout  time.Duration `json:"timeout" yaml:"timeout"`
}

// DSN 生成DSN連接字符串
func (c *ConnectionConfig) DSN(database string) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=%s&timeout=%s",
		c.UserName,
		c.Password,
		c.Host,
		c.Port,
		database,
		c.Loc,
		c.Timeout,
	)
}

// New 創建一個支持讀寫分離的數據庫連接
func New(conn *DBConn) (*gorm.DB, error) {
	if conn.Database == "" {
		return nil, errors.New("database name is required")
	}

	// 創建主庫連接
	masterDSN := conn.Master.DSN(conn.Database)
	dbBase, err := gorm.Open(mysql.Open(masterDSN), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrapf(err, "open database connection: %s", masterDSN)
	}

	// 如果有從庫配置，設置讀寫分離
	if len(conn.Replicas) > 0 {
		var replicas []gorm.Dialector
		for _, replica := range conn.Replicas {
			replicaDSN := replica.DSN(conn.Database)
			replicas = append(replicas, mysql.Open(replicaDSN))
		}

		// 註冊 dbresolver 插件
		err = dbBase.Use(dbresolver.Register(dbresolver.Config{
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		}))
		if err != nil {
			return nil, errors.Wrap(err, "failed to register dbresolver")
		}
	}

	// 獲取底層 SQL DB 對象以設置連接池參數
	sqlDB, err := dbBase.DB()
	if err != nil {
		return nil, errors.Wrap(err, "get connect pool failed")
	}

	// 設置連接池參數
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

	return dbBase, nil
}

// Close 關閉數據庫連接
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get underlying DB")
	}
	return sqlDB.Close()
}

// Ping 檢查數據庫連接是否正常
func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get underlying DB")
	}
	return sqlDB.PingContext(ctx)
}
