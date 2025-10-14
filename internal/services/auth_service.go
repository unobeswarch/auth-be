package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/unobeswarch/businesslogic/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsuarioExistente = errors.New("USER_ALREADY_EXISTS")
	ErrDatosEnviados    = errors.New("VALIDATION_ERROR")
	ErrTratamientoDatos = errors.New("BUSINESS_RULE_VIOLATION")
)

type AuthService struct {
	Key []byte
}

type UserClaims struct {
	UserID string
	Email  string
	Role   string
	Name   string
}

func RegistrarUsuario(u models.User) (int, time.Time, error) {
	if u.NombreCompleto == "" || u.Correo == "" || u.Contrasena == "" || len(u.Contrasena) < 8 {
		return 0, time.Time{}, ErrDatosEnviados
	}

	if !u.AceptaTratamientoDatos {
		return 0, time.Time{}, ErrTratamientoDatos
	}

	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=123 dbname=blogic_db sslmode=disable")
	// Si falla, probar con URL encoding (línea comentada abajo)
	// db, err := sql.Open("postgres", "postgres://postgres:BDatosPost0912%2B@localhost:5432/blogic_db?sslmode=disable")
	// Original password (comentada):
	// db, err := sql.Open("postgres", "postgres://postgres:123@localhost:5432/blogic_db?sslmode=disable")
	if err != nil {
		return 0, time.Time{}, err
	}
	defer db.Close()

	var existe bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE correo=$1 OR identificacion=$2)", u.Correo, u.Identificacion).Scan(&existe)
	if err != nil {
		return 0, time.Time{}, err
	}
	if existe {
		return 0, time.Time{}, ErrUsuarioExistente
	}

	hash_contrasena, err := bcrypt.GenerateFromPassword([]byte(u.Contrasena), bcrypt.DefaultCost)
	if err != nil {
		return 0, time.Time{}, err
	}

	query := `
		INSERT INTO usuarios 
		(nombre_completo, edad, rol, identificacion, correo, contrasena, acepta_tratamiento_datos)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, fecha_creacion
	`

	var id int
	var fechaCreacion time.Time

	err = db.QueryRow(
		query,
		u.NombreCompleto,
		u.Edad,
		u.Rol,
		u.Identificacion,
		u.Correo,
		hash_contrasena,
		u.AceptaTratamientoDatos,
	).Scan(&id, &fechaCreacion)
	if err != nil {
		return 0, time.Time{}, err
	}

	return id, fechaCreacion, nil
}

func IniciarSesion(correo string, contrasena string) (string, int, string, string, string, error) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=123 dbname=blogic_db sslmode=disable")
	// Original:
	// db, err := sql.Open("postgres", "postgres://postgres:123@localhost:5432/blogic_db?sslmode=disable")
	if err != nil {
		return "", 0, "", "", "", err
	}
	defer db.Close()

	var (
		id_usuario         int
		contrasena_usuario string
		correo_usuario     string
		rol_usuario        string
		nombre_completo    string
	)

	query := `SELECT nombre_completo, id, correo, contrasena, rol FROM usuarios WHERE correo=$1`

	err = db.QueryRow(query, correo).Scan(&nombre_completo, &id_usuario, &correo_usuario, &contrasena_usuario, &rol_usuario)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, "", "", "", fmt.Errorf("usuario no encontrado")
		}
		return "", 0, "", "", "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(contrasena_usuario), []byte(contrasena))
	if err != nil {
		return "", 0, "", "", "", err
	}

	var auth AuthService = AuthService{
		Key: []byte("asfqwr1242t1weg"),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id_usuario":      id_usuario,
		"email":           correo_usuario,
		"rol":             rol_usuario,
		"nombre_completo": nombre_completo,
		"exp":             time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(auth.Key)
	if err != nil {
		return "", 0, "", "", "", err
	}

	return nombre_completo, id_usuario, correo_usuario, rol_usuario, tokenString, nil
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// UserExists verifica si el usuario existe en la base de datos relacional
func (s *AuthService) UserExists(ctx context.Context, userID string) (bool, error) {
	// Intentar primera con connection string sin URL encoding
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=123 dbname=blogic_db sslmode=disable")
	// Si falla, probar con URL encoding (línea comentada abajo)
	// db, err := sql.Open("postgres", "postgres://postgres:BDatosPost0912%2B@localhost:5432/blogic_db?sslmode=disable")
	// Original password (comentada):
	// db, err := sql.Open("postgres", "postgres://postgres:123@localhost:5432/blogic_db?sslmode=disable")
	if err != nil {
		return false, err
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE id=$1)", userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ValidateTokenAndRole valida el token de autorización y verifica el rol
func (s *AuthService) ValidateTokenAndRole(ctx context.Context, authHeader string, requiredRole string) (*UserClaims, error) {
	if authHeader == "" {
		return nil, errors.New("token de autorización requerido")
	}

	// Extraer el token del header "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("formato de token inválido")
	}

	token := parts[1]

	userClaims, err := s.ValidateJWT(token)
	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}

	// Verificar rol
	if userClaims.Role != requiredRole {
		return nil, fmt.Errorf("acceso denegado: se requiere rol %s, pero el usuario tiene rol %s",
			requiredRole, userClaims.Role)
	}

	return userClaims, nil
}

func (s *AuthService) ValidateJWT(token string) (*UserClaims, error) {
	var auth AuthService = AuthService{
		Key: []byte("asfqwr1242t1weg"),
	}

	tkn, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return auth.Key, nil
	})
	if err != nil {
		return nil, err
	}

	claims := tkn.Claims.(jwt.MapClaims)

	return &UserClaims{
		UserID: fmt.Sprintf("%v", claims["id_usuario"]),
		Email:  fmt.Sprintf("%v", claims["email"]),
		Role:   fmt.Sprintf("%v", claims["rol"]),
		Name:   fmt.Sprintf("%v", claims["nombre_completo"]),
	}, nil
}

func (s *AuthService) GuardarFoto(ctx context.Context, authHeader string, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	db, err := sql.Open("postgres", "postgres://postgres:123@localhost:5432/blogic_db?sslmode=disable")

	if err != nil {
		return "", err
	}

	defer db.Close()

	if authHeader == "" {
		return "", errors.New("token de autorización requerido")
	}

	// Extraer el token del header "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("formato de token inválido")
	}

	token := parts[1]

	userClaims, err := s.ValidateJWT(token)

	uploadDir := "./uploads"
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("error creando carpeta de uploads: %w", err)
	}

	extension := filepath.Ext(fileHeader.Filename)
	uniqueFileName := fmt.Sprintf("%s%s", uuid.New().String(), extension)
	filePath := filepath.Join(uploadDir, uniqueFileName)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("error creando archivo destino: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", fmt.Errorf("error copiando archivo: %w", err)
	}

	imagenURL := fmt.Sprintf("/uploads/%s", uniqueFileName)

	_, err = db.ExecContext(ctx, `UPDATE usuarios SET imagen_url = $1 WHERE id = $2`, imagenURL, userClaims.UserID)
	if err != nil {
		return "", fmt.Errorf("error actualizando base de datos: %w", err)
	}

	return imagenURL, nil

}

func (s *AuthService) RetornarFoto(ctx context.Context, userID string) (string, error) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=123 dbname=blogic_db sslmode=disable")
	if err != nil {
		return "", err
	}
	defer db.Close()

	var imagePath string
	err = db.QueryRow("SELECT imagen_url FROM usuarios WHERE id = $1", userID).Scan(&imagePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("usuario no encontrado")
		}
		return "", err
	}

	return imagePath, nil
}
