# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.1.0] - 2026-02-15

### Changed

- Switched repository root from `go.mod` to `go.work` for multi-module development
- Upgraded database dependencies across modules:
  - MongoDB driver migrated to `go.mongodb.org/mongo-driver/v2`
  - GORM/PostgreSQL/MySQL/Redis/Valkey dependencies bumped to newer versions
- Updated Valkey modules (`single`, `sentinel`, `cluster`) to use `valkey-go` `ClientOption` API
- Enhanced PostgreSQL module:
  - Added preset support (`supabase_transaction`)
  - Added explicit `GORM` and `PGX` configuration overrides
  - Improved DSN value escaping and runtime parameter ordering
- Added Makefile release automation:
  - Resolve release version from `CHANGELOG.md`
  - Tag changed submodules by commit range (`tag-changed-modules`)

## [v1.0.0] - 2024-03-12

### Added

- MySQL connection management with GORM support
  - Connection pool configuration with customizable settings
  - Read/write timeout settings
  - Master-slave replication support
- PostgreSQL connection management with GORM support
  - Connection pool configuration with customizable settings
  - Statement timeout and idle transaction timeout settings
  - Master-slave replication support
  - Built-in health check with pgx
- Redis connection management
  - Standalone mode with single node support
  - Sentinel mode with high availability
  - Cluster mode with sharding support
  - Connection pool configuration with customizable settings
  - Timeout settings (dial, read, write)
- MongoDB connection management
  - Connection pool configuration with customizable settings
  - Connection timeout settings
  - Authentication support
  - Timeout settings for operations
  - Health check methods
