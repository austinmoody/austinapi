// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/sleep/date/{date}": {
            "get": {
                "description": "Retrieves sleep information with specified date",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "sleep"
                ],
                "summary": "Get sleep information by date",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Date",
                        "name": "date",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/main.Sleep"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/main.GenericMessage"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/main.GenericMessage"
                        }
                    }
                }
            }
        },
        "/sleep/id/{id}": {
            "get": {
                "description": "Retrieves sleep information with specified ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "sleep"
                ],
                "summary": "Get sleep information by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Sleep ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/main.Sleep"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/main.GenericMessage"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/main.GenericMessage"
                        }
                    }
                }
            }
        },
        "/sleep/list": {
            "get": {
                "description": "Retrieves list of sleep information in descending order by date\nSpecifying no query parameters pulls list starting with latest\nCaller can then specify a next_token or previous_token returned from\ncalls to go forward and back in the list of items.  Only next_token OR\nprevious_token should be specified.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "sleep"
                ],
                "summary": "Get list of sleep information",
                "parameters": [
                    {
                        "type": "string",
                        "format": "string",
                        "description": "next list search by next_token",
                        "name": "next_token",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "format": "string",
                        "description": "previous list search by previous_token",
                        "name": "previous_token",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/main.Sleeps"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/main.GenericMessage"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "main.GenericMessage": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "main.Sleep": {
            "type": "object",
            "properties": {
                "createdTimestamp": {
                    "type": "string"
                },
                "date": {
                    "type": "string"
                },
                "deepSleep": {
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "lightSleep": {
                    "type": "integer"
                },
                "rating": {
                    "type": "integer"
                },
                "remSleep": {
                    "type": "integer"
                },
                "totalSleep": {
                    "type": "integer"
                },
                "updatedTimestamp": {
                    "type": "string"
                }
            }
        },
        "main.Sleeps": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/main.Sleep"
                    }
                },
                "next_token": {
                    "type": "string"
                },
                "previous_token": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
