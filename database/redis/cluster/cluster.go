package cluster

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// Conn Redis 集群連線配置
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

// New 創建一個新的 Redis 集群客戶端
func New(conn *Conn) (*redis.ClusterClient, error) {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           conn.Address,
		Username:        conn.Username,
		Password:        conn.Password,
		DialTimeout:     conn.DialTimeout,
		ReadTimeout:     conn.ReadTimeout,
		WriteTimeout:    conn.WriteTimeout,
		PoolSize:        conn.PoolSize,
		MinIdleConns:    conn.MinIdleConns,
		MaxIdleConns:    conn.MaxIdleConns,
		ConnMaxIdleTime: conn.ConnMaxIdleTime,
	}), nil
}
