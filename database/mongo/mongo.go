package mongo

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	_defaultMaxPoolSize     = 100
	_defaultMinPoolSize     = 10
	_defaultMaxConnIdleTime = 5 * time.Minute
)

// DBConn MongoDB connection config
type DBConn struct {
	Hosts           []string          `json:"hosts" yaml:"hosts"`
	Username        string            `json:"username" yaml:"username"`
	Password        string            `json:"password" yaml:"password"`
	AuthDB          string            `json:"authDB" yaml:"authDB"`
	MaxPoolSize     uint64            `json:"maxPoolSize" yaml:"maxPoolSize"`
	MinPoolSize     uint64            `json:"minPoolSize" yaml:"minPoolSize"`
	MaxConnIdleTime time.Duration     `json:"maxConnIdleTime" yaml:"maxConnIdleTime"`
	ConnectTimeout  time.Duration     `json:"connectTimeout" yaml:"connectTimeout"`
	Options         map[string]string `json:"options" yaml:"options"`
}

// New creates a new MongoDB client
func New(conn *DBConn) (*mongo.Client, error) {
	// Build connection URI.
	var auth, optionsStr string
	if conn.Username != "" && conn.Password != "" {
		auth = fmt.Sprintf("%s:%s@", conn.Username, conn.Password)
	}

	hosts := strings.Join(conn.Hosts, ",")
	authDB := conn.AuthDB

	// Handle extra options.
	if len(conn.Options) > 0 {
		var params []string
		for key, value := range conn.Options {
			params = append(params, fmt.Sprintf("%s=%s", key, url.QueryEscape(value)))
		}
		optionsStr = "?" + strings.Join(params, "&")
	}

	uri := fmt.Sprintf("mongodb://%s%s/%s%s", auth, hosts, authDB, optionsStr)

	// Configure connection pool settings.
	maxPoolSize := conn.MaxPoolSize
	if maxPoolSize == 0 {
		maxPoolSize = _defaultMaxPoolSize
	}

	minPoolSize := conn.MinPoolSize
	if minPoolSize == 0 {
		minPoolSize = _defaultMinPoolSize
	}

	maxConnIdleTime := _defaultMaxConnIdleTime
	if conn.MaxConnIdleTime > 0 {
		maxConnIdleTime = conn.MaxConnIdleTime
	}

	// Configure client options.
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(maxPoolSize).
		SetMinPoolSize(minPoolSize).
		SetMaxConnIdleTime(maxConnIdleTime)

	if conn.ConnectTimeout > 0 {
		clientOptions.SetConnectTimeout(conn.ConnectTimeout)
	}

	// Establish connection.
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "mongo connect failed")
	}

	return client, nil
}
