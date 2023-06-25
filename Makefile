test:
	docker build -t gia-common-tester .
	docker run --rm gia-common-tester

coverage-report:
	go test -v -coverprofile cover.out ./...
	go tool cover -html=cover.out -o cover.html
	open cover.html