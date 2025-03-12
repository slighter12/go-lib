# go-lib

A collection of Go libraries for database connections and common utilities.

## Features

- Database Connections
  - MySQL: Connection pool management with GORM
  - PostgreSQL: Connection pool management with GORM
  - Redis: Support for standalone, sentinel, and cluster modes
  - MongoDB: Connection management with official driver

## Installation

Each package can be installed independently:

```bash
# MySQL
go get github.com/slighter12/go-lib/database/mysql

# PostgreSQL
go get github.com/slighter12/go-lib/database/postgres

# Redis (choose one)
go get github.com/slighter12/go-lib/database/redis/single    # Standalone mode
go get github.com/slighter12/go-lib/database/redis/sentinel  # Sentinel mode
go get github.com/slighter12/go-lib/database/redis/cluster   # Cluster mode

# MongoDB
go get github.com/slighter12/go-lib/database/mongo
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
}

client := single.New(cfg)
```

### MongoDB

```go
import "github.com/slighter12/go-lib/database/mongo"

cfg := &mongo.DBConn{
    URI:      "mongodb://localhost:27017",
    Database: "dbname",
}

client, err := mongo.New(cfg)
if err != nil {
    log.Fatal(err)
}
```

## License

MIT License