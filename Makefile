
# env GOOS=linux GOARCH=arm GOARM=6 go build -o ./airmon ./cmd/airmon/*.go

GOOS=linux
GOARCH=arm
GOARM=6

.PHONY: all

all: airmon spi_test

airmon: cmd/airmon/main.go
	go build -o $@ ./cmd/airmon/*.go

spi_test: cmd/spi_test/main.go
	go build -o $@ ./cmd/spi_test/*.go

clean:
	rm airmon spi_test