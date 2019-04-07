package api

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Reserved bool   `json:"reserved"`
	Metadata string `json:"metadata,omitempty"`
}

type CredentialsStore struct {
	Credentials []Credential
}

type Worker struct {
	IP       string `json:"ip"`
	SSHKey   string `json:"sshKey"`
	Reserved bool   `json:"reserved"`
	Name     string `json:"name"`
}

type WorkersStore struct {
	Workers []Worker
}
