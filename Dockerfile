FROM golang AS builder

# Build root-less, multi-stage image.
# See:
# - https://docs.docker.com/language/golang/build-images/#multi-stage-builds
# For further reference:
# - https://chemidy.medium.com/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
# - https://pythonspeed.com/articles/root-capabilities-docker-security/

WORKDIR /app

# Fetch dependencies separately, so that this is cached
# by Docker if the go.mod or go.sum files are not changed.
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY additional_robots.txt ./
COPY LICENSE ./

RUN CGO_ENABLED=0 GOOS=linux go build -o robots-txt-extender

FROM gcr.io/distroless/static-debian12 AS release
LABEL org.opencontainers.image.authors="denis@golubev.dev"

WORKDIR /
COPY --from=builder /app/robots-txt-extender .
# We need the default robots.txt
COPY --from=builder /app/additional_robots.txt .

# As we are not using a root user, we need to set the port to a non-privileged one.
ENV PORT=8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/robots-txt-extender"]
