package models

type User struct {
	NombreCompleto         string `json:"nombre_completo"`
	Edad                   int    `json:"edad"`
	Rol                    string `json:"rol"`
	Identificacion         string `json:"identificacion"`
	Correo                 string `json:"correo"`
	Contrasena             string `json:"contrasena"`
	AceptaTratamientoDatos bool   `json:"acepta_tratamiento_datos"`
	ImagenURL              string `json:"imagen_url"`
}

type UserInfoResponse struct {
	NombreCompleto string `json:"nombre_completo"`
	Identificacion string `json:"identificacion"`
	Correo         string `json:"correo"`
}
