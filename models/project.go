package models

//Project Project
type Project struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	ImagePullSecrets    string     `json:"imagePullSecrets"`
	AllowPrivilegedJobs bool       `json:"allowPrivilegedJobs"`
	AllowHostMounts     bool       `json:"allowHostMounts"`
	Providers           []Provider `json:"providers"`
	Kubernetes          Kubernetes `json:"kubernetes"`
}
