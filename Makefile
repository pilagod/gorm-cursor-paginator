test:
	go test -v ./...

test-cursor:
	go test -v ./cursor

test-paginator:
	go test -v ./paginator

test-env-up:
	docker-compose up -d

test-env-down:
	docker-compose down -v

test-coverage:
	go test --coverprofile=c.out ./...
	go tool cover -html=c.out
