package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "{{.TermsOfService}}",
        "contact": {
            "name": "{{.Contact.Name}}",
            "url": "{{.Contact.URL}}",
            "email": "{{.Contact.Email}}"
        },
        "license": {
            "name": "{{.License.Name}}",
            "url": "{{.License.URL}}"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "健康检查",
                "tags": ["健康检查"],
                "summary": "健康检查",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "code": { "type": "integer", "example": 0 },
                                "message": { "type": "string", "example": "success" },
                                "data": {
                                    "type": "object",
                                    "properties": {
                                        "status": { "type": "string", "example": "ok" }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    },
    "securityDefinitions": {
        "Bearer": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "请输入 JWT Token，格式：Bearer {token}"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"http", "https"},
	Title:            "团圆寻亲志愿者系统 API",
	Description:      "团圆寻亲志愿者系统后端 API 文档 - 包含认证、用户、组织、走失人员、任务、方言、文件、仪表盘、审计日志等模块",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
