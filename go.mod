module github.com/slighter12/go-lib

go 1.23.6

replace (
	github.com/slighter12/go-lib/database/mongo => ./database/mongo
	github.com/slighter12/go-lib/database/mysql => ./database/mysql
	github.com/slighter12/go-lib/database/postgres => ./database/postgres
	github.com/slighter12/go-lib/database/redis/cluster => ./database/redis/cluster
	github.com/slighter12/go-lib/database/redis/sentinel => ./database/redis/sentinel
	github.com/slighter12/go-lib/database/redis/single => ./database/redis/single
)

require github.com/valkey-io/valkey-go v1.0.68

require golang.org/x/sys v0.31.0 // indirect
