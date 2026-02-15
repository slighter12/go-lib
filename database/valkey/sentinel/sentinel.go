package sentinel

import (
	"net"
	"time"

	"github.com/valkey-io/valkey-go"
)

const (
	_defaultBlockingPoolSize    = 25              // Default connection pool size.
	_defaultDialTimeout         = 5 * time.Second // Default connection timeout.
	_defaultConnWriteTimeout    = 3 * time.Second // Default read/write timeout.
	_defaultBlockingPoolCleanup = 5 * time.Minute // Default connection pool cleanup interval.
)

// Conn Valkey sentinel connection config.
type Conn struct {
	MasterName          string        `json:"masterName" yaml:"masterName"`
	Address             []string      `json:"address" yaml:"address"`
	Username            string        `json:"username" yaml:"username"`
	Password            string        `json:"password" yaml:"password"`
	SentinelUsername    string        `json:"sentinelUsername" yaml:"sentinelUsername"`
	SentinelPassword    string        `json:"sentinelPassword" yaml:"sentinelPassword"`
	DB                  int           `json:"db" yaml:"db"`
	DialTimeout         time.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	ConnWriteTimeout    time.Duration `json:"connWriteTimeout" yaml:"connWriteTimeout"`
	BlockingPoolSize    int           `json:"blockingPoolSize" yaml:"blockingPoolSize"`
	BlockingPoolCleanup time.Duration `json:"blockingPoolCleanup" yaml:"blockingPoolCleanup"`
	BlockingPoolMinSize int           `json:"blockingPoolMinSize" yaml:"blockingPoolMinSize"`
}

// New creates a new Valkey sentinel client.
func New(conn *Conn) (valkey.Client, error) {
	// Use defaults.
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

	sentinelUsername := conn.Username
	if conn.SentinelUsername != "" {
		sentinelUsername = conn.SentinelUsername
	}

	sentinelPassword := conn.Password
	if conn.SentinelPassword != "" {
		sentinelPassword = conn.SentinelPassword
	}

	dialer := net.Dialer{Timeout: dialTimeout}

	return valkey.NewClient(valkey.ClientOption{
		InitAddress:         conn.Address,
		Username:            conn.Username,
		Password:            conn.Password,
		SelectDB:            conn.DB,
		Dialer:              dialer,
		ConnWriteTimeout:    connWriteTimeout,
		BlockingPoolSize:    blockingPoolSize,
		BlockingPoolCleanup: blockingPoolCleanup,
		BlockingPoolMinSize: conn.BlockingPoolMinSize,
		Sentinel: valkey.SentinelOption{
			MasterSet: conn.MasterName,
			Dialer:    dialer,
			Username:  sentinelUsername,
			Password:  sentinelPassword,
		},
	})
}
