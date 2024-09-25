include .env

live/server:
	source .env && go run cmd/app/main.go 

db/up:
	source .env && docker compose --env-file .env -p go-cook-book up db -d
db/migration/up:
	docker compose --env-file .env -p go-cook-book up migrate
db/drop:
	docker compose --env-file .env -p go-cook-book up drop
db/down:
	docker compose --env-file .env -p go-cook-book down
db/migration/create:
	docker run -v $(PWD)/migrations:/migrations migrate/migrate -path=/migrations/ -verbose create -ext sql -dir /migrations $(name)

test/db/up:
	source .env.test && docker compose --env-file .env.test -p test-go-cookbook up db -d
test/db/migration/up:
	source .env.test; docker compose --env-file .env.test -p test-go-cookbook up migrate
test/db/drop:
	source .env.test; docker compose --env-file .env.test -p test-go-cookbook up drop
test/db/down:
	source .env.test; docker compose --env-file .env.test -p test-go-cookbook down
