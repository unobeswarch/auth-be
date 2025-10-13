# BusinessLogic Service

Microservicio **BusinessLogic** desarrollado en Go.  
Se encarga de orquestar la comunicación entre frontend, prediagnostic y otros componentes del sistema, 
exponiendo endpoints **REST** y **GraphQL**.

---

## 📂 Estructura del Proyecto

```bash
/businesslogic
│── cmd/
│   └── server/                  # main.go vive aquí (entrypoint del microservicio)
│
│── internal/                    # Código interno, no expuesto a otros módulos
│   ├── config/                  # Configuración (archivos .env, variables globales, setup de GraphQL/REST)
│   ├── server/                  # Inicialización de servidores REST y GraphQL
│   ├── handlers/                # Lógica de endpoints (REST y GraphQL resolvers)
│   ├── services/                # Lógica de negocio (orquestación entre componentes externos)
│   ├── clients/                 # Conexiones HTTP/GraphQL a otros microservicios
│   ├── models/                  # Definición de estructuras de datos
│   └── utils/                   # Utilidades comunes (JWT parsing, logging, errores)
│
│── graph/                       # Archivos autogenerados por gqlgen para GraphQL
│   ├── schema.graphqls          # Definición del esquema GraphQL
│   ├── generated.go
│   └── resolvers.go
│
│── pkg/                         # Librerías reutilizables (ej: middlewares)
│
│── go.mod                       # Definición del módulo y dependencias
│── go.sum                       # Checksum de dependencias
