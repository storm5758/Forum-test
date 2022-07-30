module github.com/moguchev/BD-Forum

go 1.17

replace github.com/moguchev/BD-Forum/pkg/api => ./pkg/api

replace github.com/moguchev/BD-Forum/pkg/api/models => ./pkg/api/models

require (
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.0
	github.com/jackc/pgtype v1.11.0
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v4 v4.16.1
	github.com/jmoiron/sqlx v1.3.5
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
	google.golang.org/genproto v0.0.0-20220728213248-dd149ef739b9
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/Masterminds/squirrel v1.5.3 // indirect
	github.com/georgysavva/scany v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgconn v1.12.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/puddle v1.2.1 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e // indirect
	golang.org/x/sys v0.0.0-20220610221304-9f5ed59c137d // indirect
	golang.org/x/text v0.3.7 // indirect
)
