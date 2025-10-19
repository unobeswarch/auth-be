## Endpoints

### 1. Registrar Usuario
Registra un nuevo usuario en el sistema.

**Endpoint:** `POST /register`

**Request Body:**
```json
{
  "nombre_completo": "string",
  "edad": "number",
  "rol": "string",
  "identificacion": "string",
  "correo": "string",
  "contrasena": "string (mínimo 8 caracteres)",
  "acepta_tratamiento_datos": "boolean"
}
```

**Response exitoso (201):**
```json
{
  "id": "number",
  "mensaje": "Usuario registrado exitosamente",
  "fecha_registro": "timestamp"
}
```
---

### 2. Iniciar Sesión
Autentica un usuario y genera un token JWT.

**Endpoint:** `POST /auth`

**Request Body:**
```json
{
  "correo": "string",
  "contrasena": "string"
}
```

**Response exitoso (200):**
```json
{
  "nombre": "string",
  "token": "string (JWT)",
  "rol": "string",
  "user_id": "number",
  "correo": "string"
}
```
---

### 3. Guardar Foto de Perfil
Permite subir una foto de perfil para el usuario autenticado.

**Endpoint:** `POST /upload`

**Headers:**
```
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**Form Data:**
- `foto`: archivo de imagen (máximo 10 MB)

**Response exitoso (200):**
```json
{
  "mensaje": "Foto guardada correctamente",
  "imagen_url": "/uploads/unique-filename.ext"
}
```
---

### 4. Validar Token y Rol
Valida un token JWT y verifica que el usuario tenga el rol requerido.

**Endpoint:** `POST /validation`

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "required_role": "string"
}
```

**Response exitoso (200):**
```json
{
  "UserID": "string",
  "Email": "string",
  "Role": "string",
  "Name": "string"
}
```
---

### 5. Verificar Existencia de Usuario
Verifica si un usuario existe en la base de datos mediante su ID.

**Endpoint:** `POST /userExists`

**Request Body:**
```json
{
  "user_id": "string"
}
```

**Response exitoso (200):**
```json
{
  "exists": "boolean"
}
```
---

### 6. Obtener Imagen de Usuario
Retorna la imagen de perfil de un usuario específico.

**Endpoint:** `GET /userInfo`

**Query Parameters:**
- `id`: ID del usuario (requerido)

**Response exitoso (200):**
- Retorna el archivo de imagen con el Content-Type apropiado (image/jpeg, image/png, etc.)

### 7. Obtener datos del usuario
Retorna nombre, email y rol del usuario

**Endpoint:** `GET /userImage?id=<user_id>`

**Headers:**
```
Authorization: Bearer <token>
```

**Response exitoso (200):**
```json
{
  "nombre": "string",
	"email":  "string",
	"rol":    "string",
}
```

---

### 7. Obtener Información de Usuario
Obtiene información básica de un usuario específico (nombre completo, identificación y correo).

**Endpoint:** `GET /getUserInfo?id=<user_id>`

**Query Parameters:**
- `id`: ID del usuario (requerido)

**Response exitoso (200):**
```json
{
  "nombre_completo": "string",
  "identificacion": "string", 
  "correo": "string"
}
```

**Errores:**
- **400 Bad Request:**
```json
{
  "error": "Falta parámetro 'id'"
}
```

- **404 Not Found:**
```json
{
  "error": "Usuario no encontrado"
}
```

- **500 Internal Server Error:**
```json
{
  "error": "Error interno del servidor"
}
```
