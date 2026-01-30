package handlers

var _ Handler = TerraformProxmoxHandler{}

type TerraformProxmoxHandler struct{}

func NewTerraformProxmoxHandler() TerraformProxmoxHandler {
	return TerraformProxmoxHandler{}
}

func (h TerraformProxmoxHandler) GetConfig(_ string) any {
	return nil
}

func (h TerraformProxmoxHandler) Execute(config any, hostAlias string) (map[string]string, error) {
	return nil, nil
}
