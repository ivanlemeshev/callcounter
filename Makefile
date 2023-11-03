test:
	go test --race -v .

bench:
	go test -bench=. -benchmem -count 1
