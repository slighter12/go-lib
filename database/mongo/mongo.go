package mongo

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

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
	if conn == nil {
		return nil, errors.New("mongo connection config is required")
	}

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
		ApplyURI(connectionURI(conn)).
		SetMaxPoolSize(maxPoolSize).
		SetMinPoolSize(minPoolSize).
		SetMaxConnIdleTime(maxConnIdleTime)

	if conn.ConnectTimeout > 0 {
		clientOptions.SetConnectTimeout(conn.ConnectTimeout)
	}

	// Establish connection.
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	return client, nil
}

func connectionURI(conn *DBConn) string {
	var uri strings.Builder
	uri.WriteString("mongodb://")
	if conn.Username != "" || conn.Password != "" {
		uri.WriteString(url.UserPassword(conn.Username, conn.Password).String())
		uri.WriteByte('@')
	}
	uri.WriteString(strings.Join(conn.Hosts, ","))
	uri.WriteByte('/')
	uri.WriteString(url.PathEscape(conn.AuthDB))

	options := url.Values{}
	for key, value := range conn.Options {
		options.Set(key, value)
	}
	if encoded := options.Encode(); encoded != "" {
		uri.WriteByte('?')
		uri.WriteString(encoded)
	}
	return uri.String()
}
