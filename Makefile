.PHONY: db-up db-down

build:
	cd cmd/gophermart && go build -buildmode=exe -ldflags="-s -w" -o gophermart
	cd ../..

test:
	gophermarttest \
	  -test.v -test.run=^TestGophermart$ \
	  -gophermart-binary-path=cmd/gophermart/gophermart \
	  -gophermart-host=localhost \
	  -gophermart-port=8080 \
	  -gophermart-database-uri="postgresql://postgres:postgres@localhost:5432/postgres" \
	  -accrual-binary-path=cmd/accrual/accrual_darwin_amd64 \
	  -accrual-host=localhost \
	  -accrual-port=8089 \
	  -accrual-database-uri="postgresql://postgres:postgres@localhost:5432/postgres"

coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

vet:
	go vet -vettool=$(which statictest) ./...

lint:
	golangci-lint run

db-up:
	docker build -t gophermarttest-postgres -f Dockerfile-test-db .
	docker run -d --name gophermarttest-postgres -p 5432:5432 gophermarttest-postgres

db-down:
	docker stop gophermarttest-postgres
	docker rm gophermarttest-postgres
