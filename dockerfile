# Build the GO image
FROM golang:1.25 AS builder

# Set the working directory
WORKDIR /app

# Copy the dependencies for cache
COPY /go.mod /go.sum ./

RUN go mod download

# Copy the source code
COPY . .

# Build the static binary (no CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go

# Final small image to just run the compiled binary
FROM alpine:latest AS final

# Set the work directory for the final image
WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /app/main .

COPY /template ./template

EXPOSE 4000

CMD [ "./main" ]
