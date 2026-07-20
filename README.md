# go-lib

A collection of independently installable Go database connection libraries.

## Development status

The latest tagged release is v1.2.0.

## Requirements

- Go 1.25 or higher for database modules
- Go 1.26 or higher for the source-stack module and this workspace

## Installation

Install the latest tagged release for the package you need:

```bash
go get github.com/slighter12/go-lib/database/mysql@v1.2.0
go get github.com/slighter12/go-lib/database/postgres@v1.2.0

go get github.com/slighter12/go-lib/database/redis/single@v1.2.0
go get github.com/slighter12/go-lib/database/redis/sentinel@v1.2.0
go get github.com/slighter12/go-lib/database/redis/cluster@v1.2.0

go get github.com/slighter12/go-lib/database/valkey/single@v1.2.0
go get github.com/slighter12/go-lib/database/valkey/sentinel@v1.2.0
go get github.com/slighter12/go-lib/database/valkey/cluster@v1.2.0

go get github.com/slighter12/go-lib/database/mongo@v1.2.0
```

### Source stack

This package requires Go 1.26 or higher:

```bash
go get github.com/slighter12/go-lib/errors/stack@latest
```

Use `stack.With(err)` to capture an error's source stack. Retrieve it with
`errors.AsType[stack.Provider]` only where your logger decides it should be
rendered.

Use `stack.Replace(err, appErr)` when an application layer must expose a domain
error while preserving the original cause and any captured source stack.
Both errors remain searchable with the standard `errors` package.

## Usage

### MySQL

```go
cfg := &mysql.DBConn{
	Database: "app",
	Master: mysql.ConnectionConfig{
		Host: "localhost", Port: "3306", UserName: "user", Password: "pass",
		Loc: "UTC", Timeout: 5 * time.Second,
	},
	MaxIdleConns: 10,
	MaxOpenConns: 100,
	ConnMaxLifetime: time.Hour,
}

db, err := mysql.New(cfg)
if err != nil {
	log.Fatal(err)
}
```

### PostgreSQL

```go
cfg := &postgres.DBConn{
	Database: "app",
	Master: postgres.ConnectionConfig{
		Host: "localhost", Port: "5432", UserName: "user", Password: "pass",
	},
	MaxIdleConns: 10,
	MaxOpenConns: 100,
	ConnMaxLifetime: time.Hour,
}

db, err := postgres.New(cfg)
if err != nil {
	log.Fatal(err)
}
```

### Redis

```go
client, err := single.New(&single.Conn{
	Address: "localhost:6379",
	Password: "pass",
})
if err != nil {
	log.Fatal(err)
}
defer client.Close()
```

Use `sentinel.New` with `sentinel.Conn{MasterName, Address}` for Sentinel, or `cluster.New` with `cluster.Conn{Address}` for a cluster.

### Valkey

```go
client, err := single.New(&single.Conn{
	Address: "localhost:6379",
	Password: "pass",
})
if err != nil {
	log.Fatal(err)
}
defer client.Close()
```

Use `sentinel.New` with `sentinel.Conn{MasterName, Address}` for Sentinel, or `cluster.New` with `cluster.Conn{Address}` for a cluster.

### MongoDB

```go
client, err := mongo.New(&mongo.DBConn{
	Hosts:    []string{"localhost:27017"},
	Username: "user",
	Password: "pass",
	AuthDB:   "admin",
})
if err != nil {
	log.Fatal(err)
}
defer client.Disconnect(context.Background())
```

## Readiness and cleanup

`New` configures a client or pool but does not verify database readiness. Use the underlying driver's context-aware Ping method where readiness is required, and close the returned client or underlying `*sql.DB` during shutdown.

## License

MIT License
