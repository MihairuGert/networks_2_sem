TARGET = "./out/snake_game"

all:
	go mod tidy
	go build -o $(TARGET)
	$(TARGET)
generate:
	protoc --proto_path=./proto --go_out=./internal/domain --go_opt=paths=source_relative snakes.proto