package cluster

import (
	"time"

	"github.com/valkey-io/valkey-go"
)

const (
	_defaultPoolSize     = 25              // 默認連接池大小
	_defaultMinIdleConns = 10              // 默認最小空閒連接數
	_defaultMaxIdleConns = 25              // 默認最大空閒連接數
	_defaultDialTimeout  = 5 * time.Second // 默認連接超時時間
	_defaultReadTimeout  = 3 * time.Second // 默認讀取超時時間
	_defaultWriteTimeout = 3 * time.Second // 默認寫入超時時間
	_defaultConnMaxIdle  = 5 * time.Minute // 默認連接最大空閒時間
)

// Conn Valkey 集群連線配置
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

// New 創建一個新的 Valkey 集群客戶端
func New(conn *Conn) valkey.ClusterClient {
	// 使用默認值
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

	return valkey.NewClusterClient(valkey.ClusterOptions{
		InitAddress:     conn.Address,
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
