package sentinel

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Conn Redis 哨兵連線配置
type Conn struct {
	// 哨兵節點地址
	SentinelAddrs []string `json:"sentinelAddrs" yaml:"sentinelAddrs"`
	// 主節點名稱
	MasterName string `json:"masterName" yaml:"masterName"`
	// 認證信息
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	// 資料庫編號
	DB int `json:"db" yaml:"db"`
	// 超時設定
	DialTimeout  time.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout"`
	// 連線池設定
	PoolSize        int           `json:"poolSize"  yaml:"poolSize"`
	MinIdleConns    int           `json:"minIdleConns" yaml:"minIdleConns"`
	MaxIdleConns    int           `json:"maxIdleConns" yaml:"maxIdleConns"`
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime" yaml:"connMaxIdleTime"`
}

// New 創建一個新的 Redis 哨兵客戶端
func New(conn *Conn) (*redis.Client, error) {
	return redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:      conn.MasterName,
		SentinelAddrs:   conn.SentinelAddrs,
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
	}), nil
}

// Close 關閉 Redis 哨兵連接
func Close(client *redis.Client) error {
	return client.Close()
}

// Ping 檢查 Redis 哨兵連接是否正常
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
