package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/unobeswarch/businesslogic/internal/models"
	"github.com/unobeswarch/businesslogic/internal/services"
)

type ValidateTokenRequest struct {
	RequiredRole string `json:"required_role,omitempty"`
}

type UserExistsRequest struct {
	UserID string `json:"user_id"`
}

func HandlerRegistrarUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Metodo no permitido",
		})
		return
	}

	var usuario models.User
	err := json.NewDecoder(r.Body).Decode(&usuario)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Problema durante la conversion a JSON",
		})
		return
	}

	id, fecha, err := services.RegistrarUsuario(usuario)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch err {
		case services.ErrDatosEnviados:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "VALIDATION_ERROR",
				"mensaje": "Datos de entrada inválidos",
			})
		case services.ErrUsuarioExistente:
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "USER_ALREADY_EXISTS",
				"mensaje": "Ya existe un usuario con este correo o identificación",
			})
		case services.ErrTratamientoDatos:
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "BUSINESS_RULE_VIOLATION",
				"mensaje": "Debe aceptar el tratamiento de datos personales",
			})
		default:
			// Log del error específico para depuración
			fmt.Printf("Error específico durante registro: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":         "INTERNAL_ERROR",
				"mensaje":       "Error interno del servidor",
				"codigo_error":  "REG_001",
				"error_detalle": err.Error(),
			})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             id,
		"mensaje":        "Usuario registrado exitosamente",
		"fecha_registro": fecha,
	})
}

func HandlerIniciarSesion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Metodo no permitido",
		})
		return
	}

	var datos map[string]interface{}

	err := json.NewDecoder(r.Body).Decode(&datos)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Problema durante la conversion del JSON",
		})
		return
	}

	nombre, id, correo, rol, token, err := services.IniciarSesion(datos["correo"].(string), datos["contrasena"].(string))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"nombre":  nombre,
		"token":   token,
		"rol":     rol,
		"user_id": id,
		"correo":  correo,
	})
}

func HandlerValidacion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Metodo no permitido",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"error": "token de autorización requerido"}`, http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, `{"error": "formato de token inválido"}`, http.StatusUnauthorized)
		return
	}
	token := parts[1]

	authService := services.NewAuthService()
	claims, err := authService.ValidateJWT(token)
	if err != nil {
		http.Error(w, `{"error": "token inválido"}`, http.StatusUnauthorized)
		return
	}

	resp := services.UserClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		Name:   claims.Name,
	}

	json.NewEncoder(w).Encode(resp)
}

func HandlerGuardarFotoPerfil(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Falta header de autorización", http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error procesando formulario: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("foto")
	if err != nil {
		http.Error(w, "Error leyendo archivo: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	authService := &services.AuthService{
		Key: []byte("asfqwr1242t1weg"),
	}

	imagenURL, err := authService.GuardarFoto(r.Context(), authHeader, file, fileHeader)
	if err != nil {
		http.Error(w, "Error guardando foto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje":    "Foto guardada correctamente",
		"imagen_url": imagenURL,
	})
}

func HandlerValidarTokenYRol(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"error": "token de autorización requerido"}`, http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, `{"error": "formato de token inválido"}`, http.StatusUnauthorized)
		return
	}
	token := parts[1]

	var req ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "body inválido"}`, http.StatusBadRequest)
		return
	}

	if req.RequiredRole == "" {
		http.Error(w, `{"error": "campo 'required_role' es obligatorio"}`, http.StatusBadRequest)
		return
	}

	authService := services.NewAuthService()

	claims, err := authService.ValidateJWT(token)
	if err != nil {
		http.Error(w, `{"error": "token inválido"}`, http.StatusUnauthorized)
		return
	}

	if claims.Role != req.RequiredRole {
		http.Error(w, fmt.Sprintf(`{"error": "acceso denegado: se requiere rol '%s'"}`, req.RequiredRole), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(claims)
}

func HandlerUserExists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Método no permitido",
		})
		return
	}

	var req UserExistsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Body inválido o mal formado",
		})
		return
	}

	if req.UserID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "El campo 'user_id' es obligatorio",
		})
		return
	}

	authService := services.NewAuthService()

	exists, err := authService.UserExists(r.Context(), req.UserID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exists": exists,
	})
}

func HandlerObtenerImagenUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Falta parámetro 'id'"})
		return
	}

	authService := services.NewAuthService()
	imagePath, err := authService.RetornarFoto(r.Context(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(err.Error(), "no encontrado") {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Usuario no encontrado"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error obteniendo imagen: " + err.Error()})
		}
		return
	}

	file, err := os.Open("." + imagePath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "No se pudo abrir la imagen"})
		return
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		http.Error(w, "Error leyendo imagen", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(buffer)

	file.Seek(0, 0)

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)

}

func HandlerObtenerUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	authService := services.NewAuthService()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "token de autorización requerido", http.StatusUnauthorized)
		return
	}

	nombre, email, rol, err := authService.RetornarUsuario(r.Context(), authHeader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := map[string]string{
		"nombre": nombre,
		"email":  email,
		"rol":    rol,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func HandlerGetUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Método no permitido",
		})
		return
	}

	// Obtener userID de query params
	userID := r.URL.Query().Get("id")
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Falta parámetro 'id'",
		})
		return
	}

	authService := services.NewAuthService()
	userInfo, err := authService.GetUserInfo(r.Context(), userID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(err.Error(), "no encontrado") {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Usuario no encontrado",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Error interno del servidor",
			})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userInfo)
}
