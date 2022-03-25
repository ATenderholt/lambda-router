package domain

// DevFunction contains the settings to support development without deploying directly
// to lambda-router with AWS CLI / Terraform
type DevFunction struct {
	Handler     string
	Runtime     string
	BasePath    string `yaml:"basePath"`
	Environment []string
}
