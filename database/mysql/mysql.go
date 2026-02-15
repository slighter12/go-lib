package mysql

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const (
	_defaultMaxOpenConns = 25
	_defaultMaxIdleConns = 25
	_defaultMaxLifeTime  = 5 * time.Minute
)

// DBConn combines primary and replica configurations.
type DBConn struct {
	// Primary configuration.
	Master ConnectionConfig `json:"master" yaml:"master"`

	// Replica configuration list.
	Replicas []ConnectionConfig `json:"replicas" yaml:"replicas"`

	// Connection pool settings.
	MaxIdleConns    int           `json:"maxIdleConns" yaml:"maxIdleConns"`
	MaxOpenConns    int           `json:"maxOpenConns" yaml:"maxOpenConns"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime" yaml:"connMaxLifetime"`

	// Database name.
	Database string `json:"database" yaml:"database"`

	// MySQL timeout settings.
	ReadTimeout     time.Duration `json:"readTimeout" yaml:"readTimeout"`         // read_timeout
	WriteTimeout    time.Duration `json:"writeTimeout" yaml:"writeTimeout"`       // write_timeout
	NetReadTimeout  time.Duration `json:"netReadTimeout" yaml:"netReadTimeout"`   // net_read_timeout
	NetWriteTimeout time.Duration `json:"netWriteTimeout" yaml:"netWriteTimeout"` // net_write_timeout
	WaitTimeout     time.Duration `json:"waitTimeout" yaml:"waitTimeout"`         // wait_timeout
}

// ConnectionConfig defines the configuration for a single database connection.
type ConnectionConfig struct {
	Host     string        `json:"host" yaml:"host"`
	Port     string        `json:"port" yaml:"port"`
	UserName string        `json:"username" yaml:"username"`
	Password string        `json:"password" yaml:"password"`
	Loc      string        `json:"loc" yaml:"loc"`
	Timeout  time.Duration `json:"timeout" yaml:"timeout"` // Connection timeout.
}

// DSN generates a MySQL DSN string.
func (c *ConnectionConfig) DSN(cfg *DBConn) string {
	var dsn strings.Builder
	dsn.WriteString(c.UserName)
	dsn.WriteByte(':')
	dsn.WriteString(c.Password)
	dsn.WriteString("@tcp(")
	dsn.WriteString(c.Host)
	dsn.WriteByte(':')
	dsn.WriteString(c.Port)
	dsn.WriteString(")/")
	dsn.WriteString(cfg.Database)
	dsn.WriteByte('?')

	dsn.WriteString("charset=utf8mb4&parseTime=True&loc=")
	dsn.WriteString(c.Loc)
	dsn.WriteString("&timeout=")
	dsn.WriteString(c.Timeout.String())

	// Apply timeout settings.
	if cfg.ReadTimeout > 0 {
		dsn.WriteString("&readTimeout=")
		dsn.WriteString(cfg.ReadTimeout.String())
	}
	if cfg.WriteTimeout > 0 {
		dsn.WriteString("&writeTimeout=")
		dsn.WriteString(cfg.WriteTimeout.String())
	}
	if cfg.NetReadTimeout > 0 {
		dsn.WriteString("&net_read_timeout=")
		dsn.WriteString(strconv.FormatInt(int64(cfg.NetReadTimeout/time.Second), 10))
	}
	if cfg.NetWriteTimeout > 0 {
		dsn.WriteString("&net_write_timeout=")
		dsn.WriteString(strconv.FormatInt(int64(cfg.NetWriteTimeout/time.Second), 10))
	}
	if cfg.WaitTimeout > 0 {
		dsn.WriteString("&wait_timeout=")
		dsn.WriteString(strconv.FormatInt(int64(cfg.WaitTimeout/time.Second), 10))
	}

	return dsn.String()
}

// New creates a new database connection with read/write splitting.
func New(conn *DBConn) (*gorm.DB, error) {
	if conn.Database == "" {
		return nil, errors.New("database name is required")
	}

	// Create primary connection.
	masterDSN := conn.Master.DSN(conn)
	dbBase, err := gorm.Open(mysql.Open(masterDSN), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrapf(err, "open database connection: %s", masterDSN)
	}

	// Configure read/write splitting when replicas are provided.
	if len(conn.Replicas) > 0 {
		var replicas []gorm.Dialector
		for _, replica := range conn.Replicas {
			replicaDSN := replica.DSN(conn)
			replicas = append(replicas, mysql.Open(replicaDSN))
		}

		// Register dbresolver plugin.
		err = dbBase.Use(dbresolver.Register(dbresolver.Config{
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		}))
		if err != nil {
			return nil, errors.Wrap(err, "failed to register dbresolver")
		}
	}

	// Get underlying SQL DB object to configure pool settings.
	sqlDB, err := dbBase.DB()
	if err != nil {
		return nil, errors.Wrap(err, "get connect pool failed")
	}

	// Configure connection pool settings.
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
