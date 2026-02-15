package postgres

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const (
	_defaultMaxOpenConns = 25
	_defaultMaxIdleConns = 25
	_defaultMaxLifeTime  = 5 * time.Minute
)

type Preset string

const (
	PresetSupabaseTransaction Preset = "supabase_transaction"
)

// DBConn combines primary and replica configurations.
type DBConn struct {
	// Primary configuration.
	Master ConnectionConfig `json:"master" yaml:"master"`

	// Replica configuration list.
	Replicas []ConnectionConfig `json:"replicas" yaml:"replicas"`

	// Connection pool settings.
	MaxIdleConns    int           `json:"maxIdleConns" yaml:"maxIdleConns"`
	MaxOpenConns    int           `json:"maxOpenConns" yaml:"maxOpenConns"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime" yaml:"connMaxLifetime"`

	// Database name.
	Database string `json:"database" yaml:"database"`

	// PostgreSQL-specific settings.
	Schema     string `json:"schema" yaml:"schema"`
	SearchPath string `json:"searchPath" yaml:"searchPath"`
	SSLMode    string `json:"sslMode" yaml:"sslMode"`

	// PostgreSQL timeout settings.
	StatementTimeout                time.Duration `json:"statementTimeout" yaml:"statementTimeout"`                               // statement_timeout
	LockTimeout                     time.Duration `json:"lockTimeout" yaml:"lockTimeout"`                                         // lock_timeout
	IdleInTransactionSessionTimeout time.Duration `json:"idleInTransactionSessionTimeout" yaml:"idleInTransactionSessionTimeout"` // idle_in_transaction_session_timeout

	// pgx-specific settings.
	ApplicationName   string            `json:"applicationName" yaml:"applicationName"`
	RuntimeParams     map[string]string `json:"runtimeParams" yaml:"runtimeParams"`
	HealthCheckPeriod time.Duration     `json:"healthCheckPeriod" yaml:"healthCheckPeriod"`

	// Preset configuration for default connection behavior.
	// Optional values: "", PresetSupabaseTransaction.
	Preset Preset `json:"preset" yaml:"preset"`
	// GORM settings that override preset behavior.
	GORM *GORMConfig `json:"gorm" yaml:"gorm"`
	// PGX settings at the driver layer.
	PGX *PGXConfig `json:"pgx" yaml:"pgx"`
}

// ConnectionConfig defines the configuration for a single database connection.
type ConnectionConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	UserName string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// GORMConfig defines behavior settings at the GORM layer.
type GORMConfig struct {
	// Skip GORM default implicit transactions.
	// Useful for transaction poolers such as Supabase/pgBouncer.
	SkipDefaultTransaction *bool `json:"skipDefaultTransaction" yaml:"skipDefaultTransaction"`
	// Disable server-side prepared statements.
	// Useful for transaction poolers.
	PrepareStmt *bool `json:"prepareStmt" yaml:"prepareStmt"`
}

// PGXConfig defines settings at the pgx driver layer.
type PGXConfig struct {
	// Statement cache capacity; set to 0 to disable entirely.
	// Useful for transaction poolers.
	StatementCacheCap *int `json:"statementCacheCap" yaml:"statementCacheCap"`
}

// DSN generates a PostgreSQL connection string.
func (c *ConnectionConfig) DSN(cfg *DBConn) string {
	sslMode := "disable"
	if cfg.SSLMode != "" {
		sslMode = cfg.SSLMode
	}

	searchPath := "public"
	if cfg.SearchPath != "" {
		searchPath = cfg.SearchPath
	}

	var dsn strings.Builder
	appendDSNParam(&dsn, "host", c.Host)
	appendDSNParam(&dsn, "port", c.Port)
	appendDSNParam(&dsn, "user", c.UserName)
	appendDSNParam(&dsn, "password", c.Password)
	appendDSNParam(&dsn, "dbname", cfg.Database)
	appendDSNParam(&dsn, "sslmode", sslMode)
	appendDSNParam(&dsn, "search_path", searchPath)

	// Apply timeout settings.
	if cfg.StatementTimeout > 0 {
		appendDSNParam(&dsn, "statement_timeout", strconv.FormatInt(cfg.StatementTimeout.Milliseconds(), 10))
	}
	if cfg.LockTimeout > 0 {
		appendDSNParam(&dsn, "lock_timeout", strconv.FormatInt(cfg.LockTimeout.Milliseconds(), 10))
	}
	if cfg.IdleInTransactionSessionTimeout > 0 {
		appendDSNParam(&dsn, "idle_in_transaction_session_timeout", strconv.FormatInt(cfg.IdleInTransactionSessionTimeout.Milliseconds(), 10))
	}

	// Add pgx-specific parameters.
	if cfg.ApplicationName != "" {
		appendDSNParam(&dsn, "application_name", cfg.ApplicationName)
	}

	// Add runtime parameters.
	if len(cfg.RuntimeParams) > 0 {
		keys := make([]string, 0, len(cfg.RuntimeParams))
		for key := range cfg.RuntimeParams {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appendDSNParam(&dsn, key, cfg.RuntimeParams[key])
		}
	}

	return dsn.String()
}

func appendDSNParam(dsn *strings.Builder, key, value string) {
	if dsn.Len() > 0 {
		dsn.WriteByte(' ')
	}
	dsn.WriteString(key)
	dsn.WriteByte('=')
	dsn.WriteString(escapeDSNValue(value))
}

func escapeDSNValue(value string) string {
	if value == "" {
		return "''"
	}

	needsQuote := strings.ContainsAny(value, " \t\r\n'\\")
	if !needsQuote {
		return value
	}

	var escaped strings.Builder
	escaped.Grow(len(value) + 2)
	escaped.WriteByte('\'')
	for _, ch := range value {
		if ch == '\'' || ch == '\\' {
			escaped.WriteByte('\\')
		}
		escaped.WriteRune(ch)
	}
	escaped.WriteByte('\'')
	return escaped.String()
}

func setupConnPool(db *gorm.DB, conn *DBConn) error {
	sqlDB, err := db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get underlying DB")
	}

	maxIdleConns := _defaultMaxIdleConns
	if conn.MaxIdleConns > 0 {
		maxIdleConns = conn.MaxIdleConns
	}

	maxOpenConns := _defaultMaxOpenConns
	if conn.MaxOpenConns > 0 {
		maxOpenConns = conn.MaxOpenConns
	}

	maxLifeTime := _defaultMaxLifeTime
	if conn.ConnMaxLifetime > 0 {
		maxLifeTime = conn.ConnMaxLifetime
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(maxLifeTime)

	return nil
}

// New creates a new PostgreSQL database connection.
func New(conn *DBConn) (*gorm.DB, error) {
	preset := resolvePreset(conn.Preset)
	if conn.Preset != "" && preset == "" {
		log.Printf("postgres: unknown preset %q, using default behavior", conn.Preset)
	}

	// Create primary connection.
	masterConfig, err := pgxpool.ParseConfig(conn.Master.DSN(conn))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse master config")
	}

	// Set pgx-specific settings.
	if conn.HealthCheckPeriod > 0 {
		masterConfig.HealthCheckPeriod = conn.HealthCheckPeriod
	}

	// Apply PGX settings.
	applyPGXConfig(masterConfig.ConnConfig, conn, preset)

	masterDB := stdlib.OpenDB(*masterConfig.ConnConfig)
	// Apply GORM settings.
	gormConfig := &gorm.Config{}
	applyGORMConfig(gormConfig, conn, preset)
	dbBase, err := gorm.Open(postgres.New(postgres.Config{
		Conn: masterDB,
	}), gormConfig)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create master connection")
	}

	// Configure read/write splitting when replicas are provided.
	if len(conn.Replicas) > 0 {
		var replicas []gorm.Dialector
		for _, replica := range conn.Replicas {
			replicaConfig, err := pgxpool.ParseConfig(replica.DSN(conn))
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse replica config")
			}

			// Set pgx-specific settings.
			if conn.HealthCheckPeriod > 0 {
				replicaConfig.HealthCheckPeriod = conn.HealthCheckPeriod
			}
			applyPGXConfig(replicaConfig.ConnConfig, conn, preset)

			replicaDB := stdlib.OpenDB(*replicaConfig.ConnConfig)
			replicas = append(replicas, postgres.New(postgres.Config{
				Conn: replicaDB,
			}))
		}

		// Register dbresolver plugin.
		err = dbBase.Use(dbresolver.Register(dbresolver.Config{
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		}).SetConnMaxIdleTime(time.Hour))

		if err != nil {
			return nil, errors.Wrap(err, "failed to register dbresolver")
		}
	}

	if err := setupConnPool(dbBase, conn); err != nil {
		return nil, err
	}

	return dbBase, nil
}

// applyGORMConfig applies GORM settings.
func applyGORMConfig(cfg *gorm.Config, conn *DBConn, preset Preset) {
	if conn.GORM == nil && preset == "" {
		return
	}
	// Resolve preset.
	skipDefaultTx := cfg.SkipDefaultTransaction
	prepareStmt := cfg.PrepareStmt
	if preset == PresetSupabaseTransaction {
		skipDefaultTx = true
		prepareStmt = false
	}
	// Override with explicit config.
	if conn.GORM != nil {
		if conn.GORM.SkipDefaultTransaction != nil {
			skipDefaultTx = *conn.GORM.SkipDefaultTransaction
		}
		if conn.GORM.PrepareStmt != nil {
			prepareStmt = *conn.GORM.PrepareStmt
		}
	}
	cfg.SkipDefaultTransaction = skipDefaultTx
	cfg.PrepareStmt = prepareStmt
}

// applyPGXConfig applies PGX settings.
func applyPGXConfig(cfg *pgx.ConnConfig, conn *DBConn, preset Preset) {
	hasPreset := preset != ""
	hasExplicitStatementCacheCap := conn.PGX != nil && conn.PGX.StatementCacheCap != nil
	if !hasPreset && !hasExplicitStatementCacheCap {
		return
	}
	// Resolve preset.
	statementCacheCap := cfg.StatementCacheCapacity
	if preset == PresetSupabaseTransaction {
		statementCacheCap = 0
	}
	// Override with explicit config.
	if conn.PGX != nil {
		if conn.PGX.StatementCacheCap != nil {
			statementCacheCap = *conn.PGX.StatementCacheCap
		}
	}

	// Set statement cache capacity.
	// A value of 0 disables the statement cache.
	cfg.StatementCacheCapacity = statementCacheCap
	if statementCacheCap == 0 && (preset == PresetSupabaseTransaction || hasExplicitStatementCacheCap) {
		// Disable auto-prepare behavior for transaction poolers.
		cfg.DefaultQueryExecMode = pgx.QueryExecModeExec
	}
}

func resolvePreset(preset Preset) Preset {
	switch preset {
	case "", PresetSupabaseTransaction:
		return preset
	default:
		return ""
	}
}
