package cluster

import (
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	_defaultPoolSize     = 25              // Default connection pool size.
	_defaultMinIdleConns = 10              // Default minimum idle connections.
	_defaultMaxIdleConns = 25              // Default maximum idle connections.
	_defaultDialTimeout  = 5 * time.Second // Default connection timeout.
	_defaultReadTimeout  = 3 * time.Second // Default read timeout.
	_defaultWriteTimeout = 3 * time.Second // Default write timeout.
	_defaultConnMaxIdle  = 5 * time.Minute // Default max idle connection duration.
)

// Conn Redis cluster connection config.
type Conn struct {
	Address         []string      `json:"address" yaml:"address"`
	Username        string        `json:"username" yaml:"username"`
	Password        string        `json:"password" yaml:"password"`
	DialTimeout     time.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	ReadTimeout     time.Duration `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout    time.Duration `json:"writeTimeout" yaml:"writeTimeout"`
	PoolSize        int           `json:"poolSize"  yaml:"poolSize"`
	MinIdleConns    int           `json:"minIdleConns" yaml:"minIdleConns"`
	MaxIdleConns    int           `json:"maxIdleConns" yaml:"maxIdleConns"`
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime" yaml:"connMaxIdleTime"`
}

// New creates a new Redis cluster client.
func New(conn *Conn) *redis.ClusterClient {
	// Use defaults.
	poolSize := _defaultPoolSize
	if conn.PoolSize > 0 {
		poolSize = conn.PoolSize
	}

	minIdleConns := _defaultMinIdleConns
	if conn.MinIdleConns > 0 {
		minIdleConns = conn.MinIdleConns
	}

	maxIdleConns := _defaultMaxIdleConns
	if conn.MaxIdleConns > 0 {
		maxIdleConns = conn.MaxIdleConns
	}

	dialTimeout := _defaultDialTimeout
	if conn.DialTimeout > 0 {
		dialTimeout = conn.DialTimeout
	}

	readTimeout := _defaultReadTimeout
	if conn.ReadTimeout > 0 {
		readTimeout = conn.ReadTimeout
	}

	writeTimeout := _defaultWriteTimeout
	if conn.WriteTimeout > 0 {
		writeTimeout = conn.WriteTimeout
	}

	connMaxIdleTime := _defaultConnMaxIdle
	if conn.ConnMaxIdleTime > 0 {
		connMaxIdleTime = conn.ConnMaxIdleTime
	}

	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           conn.Address,
		Username:        conn.Username,
		Password:        conn.Password,
		DialTimeout:     dialTimeout,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		PoolSize:        poolSize,
		MinIdleConns:    minIdleConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxIdleTime: connMaxIdleTime,
	})
}
