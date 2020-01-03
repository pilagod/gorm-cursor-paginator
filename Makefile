test:
	go test -v

test-env-up:
	docker-compose up -d

test-env-down:
	docker-compose down -v

test-coverage:
	go test --coverprofile=c.out.tmp
	cat c.out.tmp | grep -v "test_.*\.go" > c.out
	rm c.out.tmp
	go tool cover -html=c.out