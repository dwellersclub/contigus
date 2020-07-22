package models

//Provider Event Provider Interface
type Provider interface {
	GetWebHookURL() string
}

//BaseProvider Project
type BaseProvider struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
