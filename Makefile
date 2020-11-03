tls-linux:
	@DOCKER_BUILDKIT=1 docker build --build-arg GOOS=linux --build-arg GOARCH=amd64 -t go-build-test .
	@docker container create --name tls-temp go-build-test
	@docker container cp tls-temp:/tls/tls .
	@docker container rm tls-temp

tls-osx:
	@DOCKER_BUILDKIT=1 docker build --build-arg GOOS=darwin --build-arg GOARCH=amd64 -t go-build-test .
	@docker container create --name tls-temp go-build-test
	@docker container cp tls-temp:/tls/tls .
	@docker container rm tls-temp

tls-windows:
	@DOCKER_BUILDKIT=1 docker build --build-arg GOOS=windows --build-arg GOARCH=386 -t go-build-test .
	@docker container create --name tls-temp go-build-test
	@docker container cp tls-temp:/tls/tls.exe .
	@docker container rm tls-temp
