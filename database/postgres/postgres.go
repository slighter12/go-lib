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

// DBConn 整合了主庫和從庫的配置
type DBConn struct {
	// 主庫配置
	Master ConnectionConfig `json:"master" yaml:"master"`

	// 從庫配置列表
	Replicas []ConnectionConfig `json:"replicas" yaml:"replicas"`

	// 連接池配置
	MaxIdleConns    int           `json:"maxIdleConns" yaml:"maxIdleConns"`
	MaxOpenConns    int           `json:"maxOpenConns" yaml:"maxOpenConns"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime" yaml:"connMaxLifetime"`

	// 數據庫名稱
	Database string `json:"database" yaml:"database"`

	// PostgreSQL 特有配置
	Schema     string `json:"schema" yaml:"schema"`
	SearchPath string `json:"searchPath" yaml:"searchPath"`
	SSLMode    string `json:"sslMode" yaml:"sslMode"`

	// PostgreSQL 超時設置
	StatementTimeout                time.Duration `json:"statementTimeout" yaml:"statementTimeout"`                               // statement_timeout
	LockTimeout                     time.Duration `json:"lockTimeout" yaml:"lockTimeout"`                                         // lock_timeout
	IdleInTransactionSessionTimeout time.Duration `json:"idleInTransactionSessionTimeout" yaml:"idleInTransactionSessionTimeout"` // idle_in_transaction_session_timeout

	// pgx 特有配置
	ApplicationName   string            `json:"applicationName" yaml:"applicationName"`
	RuntimeParams     map[string]string `json:"runtimeParams" yaml:"runtimeParams"`
	HealthCheckPeriod time.Duration     `json:"healthCheckPeriod" yaml:"healthCheckPeriod"`

	// Preset 配置 - 預設的連線行為
	// 可選值: "", PresetSupabaseTransaction
	Preset Preset `json:"preset" yaml:"preset"`
	// GORM 配置 - 覆寫 preset 的行為
	GORM *GORMConfig `json:"gorm" yaml:"gorm"`
	// PGX 配置 - 驅動層配置
	PGX *PGXConfig `json:"pgx" yaml:"pgx"`
}

// ConnectionConfig 定義單個數據庫連接的配置
type ConnectionConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	UserName string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// GORMConfig 定義 GORM 層的行為配置
type GORMConfig struct {
	// 跳過 GORM 預設的隱式交易
	// 適用於 transaction pooler 如 Supabase/pgBouncer
	SkipDefaultTransaction *bool `json:"skipDefaultTransaction" yaml:"skipDefaultTransaction"`
	// 停用 server-side prepared statements
	// 適用於 transaction pooler
	PrepareStmt *bool `json:"prepareStmt" yaml:"prepareStmt"`
}

// PGXConfig 定義 pgx 驅動層的配置
type PGXConfig struct {
	// Statement cache 容量，設為 0 可完全停用
	// 適用於 transaction pooler
	StatementCacheCap *int `json:"statementCacheCap" yaml:"statementCacheCap"`
}

// DSN 生成PostgreSQL連接字符串
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

	// 添加超時設置
	if cfg.StatementTimeout > 0 {
		appendDSNParam(&dsn, "statement_timeout", strconv.FormatInt(cfg.StatementTimeout.Milliseconds(), 10))
	}
	if cfg.LockTimeout > 0 {
		appendDSNParam(&dsn, "lock_timeout", strconv.FormatInt(cfg.LockTimeout.Milliseconds(), 10))
	}
	if cfg.IdleInTransactionSessionTimeout > 0 {
		appendDSNParam(&dsn, "idle_in_transaction_session_timeout", strconv.FormatInt(cfg.IdleInTransactionSessionTimeout.Milliseconds(), 10))
	}

	// 添加 pgx 特有參數
	if cfg.ApplicationName != "" {
		appendDSNParam(&dsn, "application_name", cfg.ApplicationName)
	}

	// 添加運行時參數
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

// New 創建一個新的 PostgreSQL 數據庫連接
func New(conn *DBConn) (*gorm.DB, error) {
	preset := resolvePreset(conn.Preset)
	if conn.Preset != "" && preset == "" {
		log.Printf("postgres: unknown preset %q, using default behavior", conn.Preset)
	}

	// 創建主庫連接
	masterConfig, err := pgxpool.ParseConfig(conn.Master.DSN(conn))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse master config")
	}

	// 設置 pgx 特有配置
	if conn.HealthCheckPeriod > 0 {
		masterConfig.HealthCheckPeriod = conn.HealthCheckPeriod
	}

	// 套用 PGX 配置
	applyPGXConfig(masterConfig.ConnConfig, conn, preset)

	masterDB := stdlib.OpenDB(*masterConfig.ConnConfig)
	// 套用 GORM 配置
	gormConfig := &gorm.Config{}
	applyGORMConfig(gormConfig, conn, preset)
	dbBase, err := gorm.Open(postgres.New(postgres.Config{
		Conn: masterDB,
	}), gormConfig)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create master connection")
	}

	// 如果有從庫配置，設置讀寫分離
	if len(conn.Replicas) > 0 {
		var replicas []gorm.Dialector
		for _, replica := range conn.Replicas {
			replicaConfig, err := pgxpool.ParseConfig(replica.DSN(conn))
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse replica config")
			}

			// 設置 pgx 特有配置
			if conn.HealthCheckPeriod > 0 {
				replicaConfig.HealthCheckPeriod = conn.HealthCheckPeriod
			}
			applyPGXConfig(replicaConfig.ConnConfig, conn, preset)

			replicaDB := stdlib.OpenDB(*replicaConfig.ConnConfig)
			replicas = append(replicas, postgres.New(postgres.Config{
				Conn: replicaDB,
			}))
		}

		// 註冊 dbresolver 插件
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

// applyGORMConfig 套用 GORM 配置
func applyGORMConfig(cfg *gorm.Config, conn *DBConn, preset Preset) {
	if conn.GORM == nil && preset == "" {
		return
	}
	// 解析 preset
	skipDefaultTx := cfg.SkipDefaultTransaction
	prepareStmt := cfg.PrepareStmt
	if preset == PresetSupabaseTransaction {
		skipDefaultTx = true
		prepareStmt = false
	}
	// 覆寫為 explicit config
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

// applyPGXConfig 套用 PGX 配置
func applyPGXConfig(cfg *pgx.ConnConfig, conn *DBConn, preset Preset) {
	hasPreset := preset != ""
	hasExplicitStatementCacheCap := conn.PGX != nil && conn.PGX.StatementCacheCap != nil
	if !hasPreset && !hasExplicitStatementCacheCap {
		return
	}
	// 解析 preset
	statementCacheCap := cfg.StatementCacheCapacity
	if preset == PresetSupabaseTransaction {
		statementCacheCap = 0
	}
	// 覆寫為 explicit config
	if conn.PGX != nil {
		if conn.PGX.StatementCacheCap != nil {
			statementCacheCap = *conn.PGX.StatementCacheCap
		}
	}

	// 設置 statement cache capacity
	// 設為 0 表示停用 statement cache
	cfg.StatementCacheCapacity = statementCacheCap
	if statementCacheCap == 0 && (preset == PresetSupabaseTransaction || hasExplicitStatementCacheCap) {
		// 對 transaction pooler 關閉自動 prepare 行為
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
