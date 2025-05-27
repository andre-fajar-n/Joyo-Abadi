# Gunakan image golang yang sesuai
FROM golang:1.23 AS builder

# Set working directory di dalam container
WORKDIR /app

# Copy go.mod dan go.sum untuk menginstal dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy seluruh source code ke dalam container
COPY . .

# Build aplikasi
RUN go build -o main .

# Gunakan image minimalis untuk menjalankan aplikasi
FROM alpine:latest

# Install dependencies untuk menjalankan aplikasi
RUN apk --no-cache add ca-certificates

# Copy aplikasi hasil build dari stage builder
COPY --from=builder /app/main /main

# Tentukan port yang digunakan aplikasi
EXPOSE 3000

# Jalankan aplikasi
CMD ["/main"]
