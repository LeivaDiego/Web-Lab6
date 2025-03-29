FROM golang:1.23

WORKDIR /app

# Copiar dependencias primero
COPY go.mod .
COPY go.sum .

# Descargar dependencias iniciales
RUN go mod tidy

# Copiar el resto del código (incluye main.go y docs/)
COPY . .

# Instalar swag CLI para generar documentación
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Descargar http-swagger y swag
RUN go get github.com/swaggo/http-swagger@latest
RUN go get github.com/swaggo/swag@latest

# Compilar el binario
RUN go build -o server .

# Exponer el puerto de la API
EXPOSE 8080

CMD ["./server"]
