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
