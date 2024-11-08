// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Cloud Engineering",
            "email": "cloud.engineering@dfds.com"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "Returns health status of the API",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Health endpoint",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/clusters/{clusterId}/schemas": {
            "get": {
                "description": "Get a list of schemas from the Confluent Cloud API.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "schemas"
                ],
                "summary": "List schemas from Confluent Cloud",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject prefix to filter schemas by",
                        "name": "subjectPrefix",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Schema"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "handlers.ErrorResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "models.Metadata": {
            "type": "object",
            "properties": {
                "properties": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "sensitive": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "tags": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "array",
                        "items": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "models.Params": {
            "type": "object",
            "properties": {
                "property1": {
                    "type": "string"
                },
                "property2": {
                    "type": "string"
                }
            }
        },
        "models.Reference": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "subject": {
                    "type": "string"
                },
                "version": {
                    "type": "integer"
                }
            }
        },
        "models.Rule": {
            "type": "object",
            "properties": {
                "disabled": {
                    "type": "boolean"
                },
                "doc": {
                    "type": "string"
                },
                "expr": {
                    "type": "string"
                },
                "kind": {
                    "type": "string"
                },
                "mode": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "onFailure": {
                    "type": "string"
                },
                "onSuccess": {
                    "type": "string"
                },
                "params": {
                    "$ref": "#/definitions/models.Params"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "type": {
                    "type": "string"
                }
            }
        },
        "models.RuleSet": {
            "type": "object",
            "properties": {
                "domainRules": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Rule"
                    }
                },
                "migrationRules": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Rule"
                    }
                }
            }
        },
        "models.Schema": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "metadata": {
                    "$ref": "#/definitions/models.Metadata"
                },
                "references": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Reference"
                    }
                },
                "ruleSet": {
                    "$ref": "#/definitions/models.RuleSet"
                },
                "schema": {
                    "type": "string"
                },
                "schemaType": {
                    "type": "string"
                },
                "subject": {
                    "type": "string"
                },
                "version": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.1",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Confluent Gateway API",
	Description:      "This is a DFDS REST API for interacting with the Confluent Cloud API",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
