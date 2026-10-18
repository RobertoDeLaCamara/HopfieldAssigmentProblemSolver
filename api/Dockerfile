# Build stage
FROM golang:1.27-alpine AS builder

# Instalar dependencias necesarias
RUN apk add --no-cache git

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Construir la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Runtime stage
FROM alpine:3.24

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates

# Crear usuario no-root
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Establecer directorio de trabajo
WORKDIR /app

# Copiar el binario desde el stage de build
COPY --from=builder /app/main .

# Cambiar propietario del archivo
RUN chown appuser:appgroup main

# Cambiar a usuario no-root
USER appuser

# Exponer puerto
EXPOSE 8080

# El contenedor está listo cuando /health responde (el puerto lo fija PORT, 8080 por defecto)
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- "http://localhost:${PORT:-8080}/health" >/dev/null || exit 1

# Comando por defecto
CMD ["./main"]
