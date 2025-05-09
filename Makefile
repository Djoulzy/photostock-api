all:
	make main
	make rebase
	make clearThumb
	make rename

clean:
	go clean -i -cache -modcache

main:
	swag init -g cmd/IS2/main.go
	go build -o IS2 cmd/IS2/*.go

install:
	go install github.com/swaggo/swag/cmd/swag@latest

rebase: cmd/rebase/main.go model/DB.go database/main.go utils/main.go
	go build -o rebase cmd/rebase/*

clearThumb: cmd/clearThumb/main.go utils/main.go
	go build -o clearThumb cmd/clearThumb/*

rename: cmd/rename/main.go database/main.go
	go build -o rename cmd/rename/*