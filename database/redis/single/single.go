package single

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Conn Redis 單機連線配置
type Conn struct {
	// 連線地址
	Addr string `json:"addr" yaml:"addr"`
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

// New 創建一個新的 Redis 單機客戶端
func New(conn *Conn) (*redis.Client, error) {
	if conn.Addr == "" {
		conn.Addr = "127.0.0.1:6379"
	}

	return redis.NewClient(&redis.Options{
		Addr:            conn.Addr,
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

// Close 關閉 Redis 單機連接
func Close(client *redis.Client) error {
	return client.Close()
}

// Ping 檢查 Redis 單機連接是否正常
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
