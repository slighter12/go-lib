# go-lib

A collection of Go libraries for database connections and common utilities.

## Requirements

- Go 1.23.6 or higher

## Version

Current stable version: v1.0.0

## Features

- Database Connections
  - MySQL: Connection pool management with GORM
  - PostgreSQL: Connection pool management with GORM
  - Redis: Support for standalone, sentinel, and cluster modes
  - MongoDB: Connection management with official driver

## Installation

Each package can be installed independently. It's recommended to use the latest stable version:

```bash
# MySQL
go get github.com/slighter12/go-lib/database/mysql@v1.0.0

# PostgreSQL
go get github.com/slighter12/go-lib/database/postgres@v1.0.0

# Redis (choose one)
go get github.com/slighter12/go-lib/database/redis/single@v1.0.0    # Standalone mode
go get github.com/slighter12/go-lib/database/redis/sentinel@v1.0.0  # Sentinel mode
go get github.com/slighter12/go-lib/database/redis/cluster@v1.0.0   # Cluster mode

# MongoDB
go get github.com/slighter12/go-lib/database/mongo@v1.0.0
```

## Usage

### MySQL

```go
import "github.com/slighter12/go-lib/database/mysql"

cfg := &mysql.DBConn{
    Host:     "localhost",
    Port:     3306,
    Username: "user",
    Password: "pass",
    Database: "dbname",
    Pool: mysql.Pool{
        MaxIdleConns: 10,
        MaxOpenConns: 100,
        ConnMaxLifetime: time.Hour,
    },
}

db, err := mysql.New(cfg)
if err != nil {
    log.Fatal(err)
}
```

### PostgreSQL

```go
import "github.com/slighter12/go-lib/database/postgres"

cfg := &postgres.DBConn{
    Host:     "localhost",
    Port:     5432,
    Username: "user",
    Password: "pass",
    Database: "dbname",
    Pool: postgres.Pool{
        MaxIdleConns: 10,
        MaxOpenConns: 100,
        ConnMaxLifetime: time.Hour,
    },
}

db, err := postgres.New(cfg)
if err != nil {
    log.Fatal(err)
}
```

### Redis

Standalone mode:
```go
import "github.com/slighter12/go-lib/database/redis/single"

cfg := &single.Conn{
    Host:     "localhost",
    Port:     6379,
    Password: "pass",
    DB:       0,
}

client := single.New(cfg)
```

Sentinel mode:
```go
import "github.com/slighter12/go-lib/database/redis/sentinel"

cfg := &sentinel.Conn{
    MasterName: "mymaster",
    Addrs:      []string{"localhost:26379"},
    Password:   "pass",
    DB:         0,
}

client := sentinel.New(cfg)
```

Cluster mode:
```go
import "github.com/slighter12/go-lib/database/redis/cluster"

cfg := &cluster.Conn{
    Addrs:    []string{"localhost:7000", "localhost:7001"},
    Password: "pass",
}

client := cluster.New(cfg)
```

### MongoDB

```go
import "github.com/slighter12/go-lib/database/mongo"

cfg := &mongo.DBConn{
    URI:      "mongodb://localhost:27017",
    Database: "dbname",
    Pool: mongo.Pool{
        MaxPoolSize: 100,
        MinPoolSize: 10,
    },
}

client, err := mongo.New(cfg)
if err != nil {
    log.Fatal(err)
}
```

## License

MIT License