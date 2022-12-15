.PHONY: build sbuild clean

build:
	go build -o image_uploader .

sbuild:
	GOOS=linux GOARCH=amd64 go build -o image_uploader .

clean:
	@rm -rf ./image_uploader
