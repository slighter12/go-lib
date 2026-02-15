package single

import (
	"net"
	"time"

	"github.com/valkey-io/valkey-go"
)

const (
	_defaultBlockingPoolSize    = 25              // 默認連接池大小
	_defaultDialTimeout         = 5 * time.Second // 默認連接超時時間
	_defaultConnWriteTimeout    = 3 * time.Second // 默認讀寫超時時間
	_defaultBlockingPoolCleanup = 5 * time.Minute // 默認連接池清理時間
)

// Conn Valkey 單機連線配置
type Conn struct {
	Address             string        `json:"address" yaml:"address"`
	Username            string        `json:"username" yaml:"username"`
	Password            string        `json:"password" yaml:"password"`
	DB                  int           `json:"db" yaml:"db"`
	DialTimeout         time.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	ConnWriteTimeout    time.Duration `json:"connWriteTimeout" yaml:"connWriteTimeout"`
	BlockingPoolSize    int           `json:"blockingPoolSize" yaml:"blockingPoolSize"`
	BlockingPoolCleanup time.Duration `json:"blockingPoolCleanup" yaml:"blockingPoolCleanup"`
	BlockingPoolMinSize int           `json:"blockingPoolMinSize" yaml:"blockingPoolMinSize"`
}

// New 創建一個新的 Valkey 單機客戶端
func New(conn *Conn) (valkey.Client, error) {
	// 使用默認值
	blockingPoolSize := _defaultBlockingPoolSize
	if conn.BlockingPoolSize > 0 {
		blockingPoolSize = conn.BlockingPoolSize
	}

	dialTimeout := _defaultDialTimeout
	if conn.DialTimeout > 0 {
		dialTimeout = conn.DialTimeout
	}

	connWriteTimeout := _defaultConnWriteTimeout
	if conn.ConnWriteTimeout > 0 {
		connWriteTimeout = conn.ConnWriteTimeout
	}

	blockingPoolCleanup := _defaultBlockingPoolCleanup
	if conn.BlockingPoolCleanup > 0 {
		blockingPoolCleanup = conn.BlockingPoolCleanup
	}

	return valkey.NewClient(valkey.ClientOption{
		InitAddress:         []string{conn.Address},
		Username:            conn.Username,
		Password:            conn.Password,
		SelectDB:            conn.DB,
		Dialer:              net.Dialer{Timeout: dialTimeout},
		ConnWriteTimeout:    connWriteTimeout,
		BlockingPoolSize:    blockingPoolSize,
		BlockingPoolCleanup: blockingPoolCleanup,
		BlockingPoolMinSize: conn.BlockingPoolMinSize,
	})
}
