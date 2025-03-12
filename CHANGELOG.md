# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.0] - 2024-03-12

### Added
- MySQL connection management with GORM support
  - Connection pool configuration with customizable settings
  - Automatic retry and failover support
  - Health check methods
- PostgreSQL connection management with GORM support
  - Connection pool configuration with customizable settings
  - Statement timeout and idle transaction timeout settings
  - Health check methods
- Redis connection management
  - Standalone mode with single node support
  - Sentinel mode with high availability
  - Cluster mode with sharding support
  - Connection pool configuration
  - Health check methods
- MongoDB connection management
  - Connection pool configuration with customizable settings
  - Timeout settings for operations
  - Health check methods 