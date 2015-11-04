.PHONY: test, demo

DEFAULT_POSTGRES_CONN	:=postgres://postgres:postgres@localhost/hero_test?sslmode=disable
DEFAULT_DIALECT			:=postgres

ifeq "$(origin DB_CONN)" "undefined"
DB_CONN=$(DEFAULT_POSTGRES_CONN)
endif

ifeq "$(origin DB_DIALECT)" "undefined"
DB_DIALECT=$(DEFAULT_DIALECT)
endif

test:
	@DB_CONN=$(DB_CONN) DB_DIALECT=$(DB_DIALECT) go test -v -cover

check:
	@go vet
	@golint

convey:
	@DB_CONN=$(DB_CONN) DB_DIALECT=$(DB_DIALECT) goconvey
	

testapp:
	@go test -v ./cmd/hero

clean:
	@go clean ./cmd/hero
	@go clean
	
deps: 
	@go get github.com/golang/lint/golint
	@go get github.com/smartystreets/goconvey	
	
server:
	@go run cmd/hero/hero.go s --migrate config_dev.json
	
travis: check
	@$HOME/gopath/bin/goveralls -service=travis-ci -repotoken=$COVERALLS
	
demo:
	@go get github.com/stretchr/codecs/...
	@go get github.com/stretchr/gomniauth
	@cd demo && go run main.go
	