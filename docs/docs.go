package docs

import (
	_ "embed"
)

//go:embed swagger.yaml
var SwaggerYAML string

//go:embed swagger.json
var SwaggerJSON string

const (
	SwaggerInfoTitle       = "Sistem Pelaporan Prestasi Mahasiswa API"
	SwaggerInfoVersion     = "1.0.0"
	SwaggerInfoDescription = "REST API untuk sistem pelaporan prestasi mahasiswa dengan dukungan Role-Based Access Control (RBAC), JWT Authentication, Achievement management dengan field dinamis, Verifikasi prestasi oleh dosen wali, dan Reporting dan analytics"
)

