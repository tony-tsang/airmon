
# env GOOS=linux GOARCH=arm GOARM=6 go build -o ./airmon ./cmd/airmon/*.go

export GOOS=linux
export GOARCH=arm
export GOARM=7

.PHONY: all

all: airmon spi_test render_test test_dps310

airmon: cmd/airmon/main.go
	go build -o bin/$@ ./cmd/airmon/*.go

spi_test: cmd/spi_test/main.go internal/pkg/uc8159/uc8159.go
	go build -o bin/$@ ./cmd/spi_test/*.go

render_test: cmd/render_test/main.go
	go build -o bin/$@ $<

test_dps310: cmd/test_dps310/main.go
	go build -o bin/$@ $<

clean:
	rm bin/*