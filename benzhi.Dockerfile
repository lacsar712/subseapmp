FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/subseapmp ./cmd/subseapmp

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/subseapmp /subseapmp
ENTRYPOINT ["/subseapmp"]