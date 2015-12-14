.PHONY: test, demo

DEFAULT_POSTGRES_CONN	:=postgres://postgres:postgres@localhost/hero_test?sslmode=disable
DEFAULT_DIALECT			:=postgres

ifeq "$(origin DB_CONN)" "undefined"
DB_CONN=$(DEFAULT_POSTGRES_CONN)
endif

ifeq "$(origin DB_DIALECT)" "undefined"
DB_DIALECT=$(DEFAULT_DIALECT)
endif

test: check testapp
	@DB_CONN=$(DB_CONN) DB_DIALECT=$(DB_DIALECT) go test -v -cover

check:
	@go vet
	@golint

convey:
	@DB_CONN=$(DB_CONN) DB_DIALECT=$(DB_DIALECT) goconvey -port 8000
	

testapp:
	@go test -v ./cmd/hero

clean:
	@go clean ./cmd/hero
	@go clean
	
deps: 
	@go get github.com/golang/lint/golint
	@go get github.com/smartystreets/goconvey	
	@go get github.com/mitchellh/gox
	
server:
	@go run cmd/hero/hero.go s --migrate -dev config_dev.json
	
	
demo:
	@go get github.com/stretchr/codecs/...
	@go get github.com/stretchr/gomniauth
	@cd demo && go run main.go

dist:
	@gox -output="bin/{{.Dir}}v$(VERSION)_{{.OS}}_{{.Arch}}/{{.Dir}}" ./cmd/hero