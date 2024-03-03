FROM golang AS builder

# Build root-less image.
# See:
# - https://chemidy.medium.com/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
# - https://pythonspeed.com/articles/root-capabilities-docker-security/
RUN useradd --create-home robots-txt-extender-user
WORKDIR /home/robots-txt-extender-user
# Ensure correct permissions on the built binary.
USER robots-txt-extender-user

# Fetch dependencies separately, so that this is cached
# by Docker if the go.mod or go.sum files are not changed.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o robots-txt-extender

FROM scratch
LABEL org.opencontainers.image.authors="denis@golubev.dev"

# We need the user created in the builder image in this second stage as well.
COPY --from=builder /etc/passwd /etc/passwd
USER robots-txt-extender-user

WORKDIR /home/robots-txt-extender-user
COPY --from=builder /home/robots-txt-extender-user/robots-txt-extender .

# TODO: DEBUG
COPY --from=builder /usr/bin/ls /usr/bin/ls

ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/home/robots-txt-extender-user/robots-txt-extender"]
