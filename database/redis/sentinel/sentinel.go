package sentinel

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// Conn Redis 哨兵連線配置
type Conn struct {
	MasterName      string        `json:"masterName" yaml:"masterName"`
	Address         []string      `json:"address" yaml:"address"`
	Username        string        `json:"username" yaml:"username"`
	Password        string        `json:"password" yaml:"password"`
	DB              int           `json:"db" yaml:"db"`
	DialTimeout     time.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	ReadTimeout     time.Duration `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout    time.Duration `json:"writeTimeout" yaml:"writeTimeout"`
	PoolSize        int           `json:"poolSize"  yaml:"poolSize"`
	MinIdleConns    int           `json:"minIdleConns" yaml:"minIdleConns"`
	MaxIdleConns    int           `json:"maxIdleConns" yaml:"maxIdleConns"`
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime" yaml:"connMaxIdleTime"`
}

// New 創建一個新的 Redis 哨兵客戶端
func New(conn *Conn) *redis.Client {
	return redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:      conn.MasterName,
		SentinelAddrs:   conn.Address,
		Username:        conn.Username,
		Password:        conn.Password,
		DB:              conn.DB,
		DialTimeout:     conn.DialTimeout,
		ReadTimeout:     conn.ReadTimeout,
		WriteTimeout:    conn.WriteTimeout,
		PoolSize:        conn.PoolSize,
		MinIdleConns:    conn.MinIdleConns,
		MaxIdleConns:    conn.MaxIdleConns,
		ConnMaxIdleTime: conn.ConnMaxIdleTime,
	})
}
