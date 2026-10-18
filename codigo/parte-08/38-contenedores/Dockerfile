# Etapa 1: compilar. Esta imagen trae Go y pesa cientos de megas,
# pero no llega a producción.
FROM golang:1.27-alpine AS compilacion
WORKDIR /src

# Primero solo las dependencias: si no cambian, Docker reutiliza esta capa.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /api ./cmd/api

# Etapa 2: ejecutar. Sin shell ni gestor de paquetes, con un usuario sin privilegios.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=compilacion /api /api
USER nonroot:nonroot
# Dentro del contenedor hay que escuchar en todas las interfaces: 127.0.0.1
# sería el propio contenedor y Docker no podría hacerle llegar el tráfico.
ENV DIRECCION=0.0.0.0:8080
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=3 CMD ["/api", "-comprobar"]
ENTRYPOINT ["/api"]
