# syntax=docker/dockerfile:1.7

FROM golang:1.22-bookworm AS build

WORKDIR /src
ENV GOWORK=off

COPY apps/api/go.mod apps/api/go.sum ./apps/api/
COPY packages/domain/go.mod ./packages/domain/
COPY packages/usecases/go.mod packages/usecases/go.sum ./packages/usecases/
COPY packages/adapters/setlistfm/go.mod ./packages/adapters/setlistfm/
COPY packages/adapters/spotify/go.mod packages/adapters/spotify/go.sum ./packages/adapters/spotify/

RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    cd apps/api && go mod download

COPY apps/api ./apps/api
COPY packages ./packages

RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
    cd apps/api && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gigtape-api .

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/gigtape-api /app/gigtape-api

ENV PORT=8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/gigtape-api"]
