package host

type SSHHost struct {
	Alias         string
	HostName      string
	User          string `json:"ssh_user" required:"true"`
	PublicKeyPath string `json:"ssh_public_key_path" required:"true"`
	Port          string `json:"ssh_port" required:"true"`
}
