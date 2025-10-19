package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/unobeswarch/businesslogic/internal/handlers"
)

const defaultPort = "8081"

func main() {
	port := defaultPort
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	// URL del servicio de prediagnóstico (configurable por variable de entorno)
	prediagnosticURL := os.Getenv("PREDIAGNOSTIC_SERVICE_URL")
	if prediagnosticURL == "" {
		prediagnosticURL = "http://localhost:8081" // URL por defecto
	}

	// Middleware para extraer Authorization header y agregarlo al contexto
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			// Extraer Authorization header
			authHeader := r.Header.Get("Authorization")

			// Agregar al contexto si existe
			ctx := r.Context()
			if authHeader != "" {
				ctx = context.WithValue(ctx, "Authorization", authHeader)
			}

			// Continuar con el siguiente handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	http.Handle("/register", authMiddleware(http.HandlerFunc(handlers.HandlerRegistrarUsuario)))
	http.Handle("/auth", authMiddleware(http.HandlerFunc(handlers.HandlerIniciarSesion)))
	http.Handle("/upload", authMiddleware(http.HandlerFunc(handlers.HandlerGuardarFotoPerfil)))
	http.Handle("/validation", authMiddleware(http.HandlerFunc(handlers.HandlerValidarTokenYRol)))
	http.Handle("/userExists", authMiddleware(http.HandlerFunc(handlers.HandlerUserExists)))
	http.Handle("/userImage", authMiddleware(http.HandlerFunc(handlers.HandlerObtenerImagenUsuario)))
	http.Handle("/userInfo", authMiddleware(http.HandlerFunc(handlers.HandlerObtenerUsuario)))
	http.Handle("/getUserInfo", authMiddleware(http.HandlerFunc(handlers.HandlerGetUserInfo)))
	log.Printf("business logic service URL: %s", prediagnosticURL)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
