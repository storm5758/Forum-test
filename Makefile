.PHONY: generate
generate:
	buf mod update
	buf generate


MIGRATIONS_DIR=./migrations
.PHONY: migration
migration:
	goose -dir=${MIGRATIONS_DIR} create $(NAME) sql 