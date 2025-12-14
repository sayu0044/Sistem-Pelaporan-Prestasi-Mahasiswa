package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sayu0044/Sistem-Pelaporan-Prestasi-Mahasiswa/docs"
)

func RegisterSwaggerRoutes(app *fiber.App) {
	swaggerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Documentation - Sistem Pelaporan Prestasi Mahasiswa</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .swagger-ui section.models {
            display: none !important;
        }
        .swagger-ui .model-box {
            display: none !important;
        }
        .swagger-ui .model-container {
            display: none !important;
        }
        .swagger-ui .model-toggle {
            display: none !important;
        }
        .swagger-ui .models-control {
            display: none !important;
        }
        .swagger-ui .models {
            display: none !important;
        }
        .swagger-ui .model-jump-to-path {
            display: none !important;
        }
        .swagger-ui .model-title {
            display: none !important;
        }
        .swagger-ui .model-box-control {
            display: none !important;
        }
        .swagger-ui .model-hint {
            display: none !important;
        }
        .swagger-ui .model-jump-to-path {
            display: none !important;
        }
        .swagger-ui .opblock-tag-section[data-tag="Schemas"] {
            display: none !important;
        }
        .swagger-ui .opblock-tag[data-tag="Schemas"] {
            display: none !important;
        }
        .swagger-ui .opblock-tag-section[data-tag="schemas"] {
            display: none !important;
        }
        .swagger-ui .opblock-tag[data-tag="schemas"] {
            display: none !important;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/api/docs/swagger.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                tryItOutEnabled: true
            });
            window.ui = ui;
            
            setTimeout(function() {
                const schemasSection = document.querySelector('.swagger-ui .opblock-tag-section[data-tag="Schemas"]');
                if (schemasSection) {
                    schemasSection.style.display = 'none';
                }
                const schemasTag = document.querySelector('.swagger-ui .opblock-tag[data-tag="Schemas"]');
                if (schemasTag) {
                    schemasTag.style.display = 'none';
                }
                const modelsSection = document.querySelector('.swagger-ui section.models');
                if (modelsSection) {
                    modelsSection.style.display = 'none';
                }
                const modelsControl = document.querySelector('.swagger-ui .models-control');
                if (modelsControl) {
                    modelsControl.style.display = 'none';
                }
                const allModels = document.querySelectorAll('.swagger-ui .model-box, .swagger-ui .model-container, .swagger-ui .model-toggle');
                allModels.forEach(function(el) {
                    el.style.display = 'none';
                });
            }, 500);
        }
    </script>
</body>
</html>`

	app.Get("/swagger", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerHTML)
	})

	app.Get("/swagger/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerHTML)
	})

	app.Get("/swagger/index.html", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerHTML)
	})

	app.Get("/api/docs", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerHTML)
	})

	app.Get("/api/docs/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerHTML)
	})

	app.Get("/api/docs/swagger.yaml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/x-yaml")
		return c.SendString(docs.SwaggerYAML)
	})

	app.Get("/api/docs/swagger.json", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendString(docs.SwaggerJSON)
	})
}
