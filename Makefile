all:
	make main
	make rebase
	make clearThumb
	make rename

clean:
	go clean -i -cache -modcache

main:
	swag init -g cmd/IS2/main.go
	go build -o /go/bin/app/IS2 cmd/IS2/*.go

install:
	go install github.com/swaggo/swag/cmd/swag@latest

rebase: cmd/rebase/main.go model/DB.go database/main.go utils/main.go
	go build -o /go/bin/app/rebase cmd/rebase/*

clearThumb: cmd/clearThumb/main.go utils/main.go
	go build -o /go/bin/app/clearThumb cmd/clearThumb/*

rename: cmd/rename/main.go database/main.go
	go build -o /go/bin/app/rename cmd/rename/*