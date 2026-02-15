package single

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

// Conn Valkey single-node connection config.
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

// New creates a new Valkey single-node client.
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
