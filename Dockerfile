FROM golang:1.19 as base
RUN go install github.com/cosmtrek/air@latest
WORKDIR /build_api
COPY . .

FROM base as dev

# TODO: For prod copy a built server from base and set an entrypoint, prob change to a distroless image as well.
FROM base as prod

