build:
	go build -o gowatch ./gowatch.go
	go build -o gomod ./gomod.go

clean:
	go clean
	rm -f gowatch gomod
