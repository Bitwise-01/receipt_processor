test: 
	go test ./...

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html coverage.out -o coverage.html


### Running locally will become problematic because no redis
# run: setup
# 	echo "Starting service via terminal"
# 	go run cmd/receipt_processor/main.go

# setup:
# 	go get ./... && go mod tidy 

docker-run: 
	echo "Starting service via Docker"
	docker-compose -f docker-compose.yml down --rmi all && docker-compose -f docker-compose.yml up -d