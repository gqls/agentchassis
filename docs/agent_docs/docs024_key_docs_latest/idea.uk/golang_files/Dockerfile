# idea.uk service — Go build (engine + billing + delivery)
FROM golang:1.22 AS build
WORKDIR /src
COPY idea-go/ ./
# stdlib-only; no module fetch needed
ENV GOPROXY=off GOTOOLCHAIN=local
RUN go build -o /idea .

FROM gcr.io/distroless/base-debian12
COPY --from=build /idea /idea
# Persist the JSON store on a mounted volume:
#   docker run -v idea_data:/data -e IDEA_DB_PATH=/data/idea_orders.json ...
ENV IDEA_DB_PATH=/data/idea_orders.json PORT=8080
VOLUME ["/data"]
EXPOSE 8080
ENTRYPOINT ["/idea"]
