FRONT_END_BINARY=frontApp
BROKER_BINARY=brokerApp
AUTH_BINARY=authApp
LOGGER_BINARY=loggerServiceApp
MAIL_BINARY=mailerApp
LISTENER_BINARY=listenerApp
MENU_BINARY=menuApp

# SERVICES = authentication-service menu-service broker-service logger-service mail-service listener-service
SERVICES = authentication-service

## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_broker build_auth build_listener build_menu # build_logger build_mail
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build -d
	@echo "Docker images built and started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker-compose down
	@echo "Done!"

## build_broker: builds the broker binary as a linux executable
build_broker:
	@echo "Building broker binary..."
	cd ../broker-service && env GOOS=linux CGO_ENABLED=0 go build -o ${BROKER_BINARY} ./cmd/api
	@echo "Done!"

## build_logger: builds the logger binary as a linux executable
build_logger:
	@echo "Building logger binary..."
	cd ../logger-service && env GOOS=linux CGO_ENABLED=0 go build -o ${LOGGER_BINARY} ./cmd/api
	@echo "Done!"

## build_listener: builds the listener binary as a linux executable
build_listener:
	@echo "Building listener binary..."
	cd ../listener-service && env GOOS=linux CGO_ENABLED=0 go build -o ${LISTENER_BINARY} .
	@echo "Done!"

## build_auth: builds the auth binary as a linux executable
build_auth:
	@echo "Building auth binary..."
	cd ../authentication-service && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTH_BINARY} ./cmd/api
	@echo "Done!" 

## build_menu: builds the menu binary as a linux executable
build_menu:
	@echo "Building auth binary..."
	cd ../menu-service && env GOOS=linux CGO_ENABLED=0 go build -o ${MENU_BINARY} ./cmd/api
	@echo "Done!" 

## build_mail: builds the mail binary as a linux executable
build_mail:
	@echo "Building mail binary..."
	cd ../mail-service && env GOOS=linux CGO_ENABLED=0 go build -o ${MAIL_BINARY} ./cmd/api
	@echo "Done!"

## generate_protos: generates the proto files for all services
generate_protos:
	@echo "Generating proto files for all services..."
	for service in ${SERVICES}; do \
		rm -f ../$$service/pb/*.go; \
		protoc --proto_path=../$$service/protos --go_out=../$$service/pb --go_opt=paths=source_relative \
		--go-grpc_out=../$$service/pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=../$$service/pb --grpc-gateway_opt=paths=source_relative \
		../$$service/protos/*.proto; \
	done
	@echo "Proto files generated for all services!"