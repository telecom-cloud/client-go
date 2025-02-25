package openapi

type CtapiResponse struct {
	StatusCode string `json:"statusCode,omitempty"`
	// Error code, which is a three part code for product.module.code
	Error string `json:"error,omitempty"`
	// Error description during failure, usually in English
	Message string `json:"message,omitempty"`
	// Error description during failure, usually in Chinese
	Description string `json:"description,omitempty"`
	// Data returned upon success
	ReturnObj interface{} `json:"returnObj"`
}

type OpenapiResponse struct {
	StatusCode int `json:"statusCode,omitempty"`
	// Error code, which is a three part code for product.module.code
	Error string `json:"error,omitempty"`
	// Error description during failure, usually in English
	Message string `json:"message,omitempty"`
	// Error description during failure, usually in Chinese
	Description string `json:"description,omitempty"`
	// Data returned upon success
	ReturnObj interface{} `json:"returnObj"`
}
