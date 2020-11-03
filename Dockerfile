FROM golang:1.15 AS download-dependencies

ARG GOOS
ARG GOARCH

WORKDIR /tls

COPY . .

RUN go mod download

FROM download-dependencies as tls-builder

ARG GOOS
ARG GOARCH

ENV CGO_ENABLED=0

# ldflags pass options to go tool link. -s removes symbol table,
# and -w does not generate DWARF debugging info, resulting in a ~30% smaller
# binary
RUN GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" apps/translit-svc/main.go

RUN mv main tls || mv main.exe tls.exe