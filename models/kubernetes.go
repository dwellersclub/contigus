package models

// Kubernetes describes the Kubernetes configuration for a project.
type Kubernetes struct {
	Namespace         string  `json:"namespace"`
	BuildStorage      Storage `json:"buildStorage"`
	CacheStorage      Storage `json:"cacheStorage"`
	AllowSecretKeyRef bool    `json:"allowSecretKeyRef"`
	ServiceAccount    string  `json:"serviceAccount"`
}

//Storage Storage Type
type Storage struct {
	Size  string `json:"size"`
	Class string `json:"class"`
}
