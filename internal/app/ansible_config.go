package app

type ansibleConfig interface {
	GetNodeIP() string
	GetSSHUser() string
	GetSSHPort() string
	GetSSHPrivateKeyPath() string
}
