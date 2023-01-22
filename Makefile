
# env GOOS=linux GOARCH=arm GOARM=6 go build -o ./airmon ./cmd/airmon/*.go

export GOOS=linux
export GOARCH=arm
export GOARM=6

.PHONY: all

all: bin/airmon bin/spi_test bin/render_test bin/test_dps310

bin/airmon: cmd/airmon/main.go go.mod
	go build -o $@ ./cmd/airmon/*.go

bin/spi_test: cmd/spi_test/main.go internal/pkg/uc8159/uc8159.go go.mod
	go build -o $@ ./cmd/spi_test/*.go

bin/render_test: cmd/render_test/main.go
	go build -o $@ $<

bin/test_dps310: cmd/test_dps310/main.go internal/pkg/dps310/dps310.go go.mod
	go build -o $@ $<

clean:
	rm bin/*
