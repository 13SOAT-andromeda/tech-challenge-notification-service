FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/bootstrap ./cmd/lambda

FROM public.ecr.aws/lambda/provided:al2023
COPY --from=builder /app/bootstrap /var/runtime/bootstrap
COPY --from=builder /app/bootstrap /var/task/bootstrap
CMD ["bootstrap"]
