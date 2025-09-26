TARGET = "./out/snake_game"

all:
	go build -o $(TARGET)
	$(TARGET)
generate:
	protoc --proto_path=./proto --go_out=./generated --go_opt=paths=source_relative snakes.proto