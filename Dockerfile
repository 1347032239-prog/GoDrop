FROM golang:1.27-bookworm AS build

WORKDIR /src

COPY go.mod ./
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/godrop .

FROM scratch

COPY --from=build /out/godrop /godrop

USER 65532:65532

ENV GODROP_ADDR=0.0.0.0:8080

EXPOSE 8080

ENTRYPOINT ["/godrop"]
CMD ["serve"]
