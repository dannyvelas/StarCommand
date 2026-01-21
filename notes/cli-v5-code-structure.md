## Problem

The command structure changed to be this:

| Sub-command       | Action | Host-Alias   | Description                                              |
|-------------------|--------|--------------|----------------------------------------------------------|
| ansible-inventory | add    | &lt;host&gt; | default logic to add a host to an ansible inventory file |
| ansible-playbook  | run    | &lt;host&gt; | run default ansible playbook for host &lt;host&gt;       |
| ssh               | add    | &lt;host&gt; | default logic to add a host to home ssh config file      |
| terraform         | apply  | proxmox      | apply specialized terraform project for proxmox          |
| ansible-playbook  | run    | proxmox      | run specialized ansible playbook for proxmox             |
| terraform         | apply  | plex         | apply specialized terraform project for plex             |
| ansible-playbook  | run    | plex         | run specialized ansible playbook for plex                |
| terraform         | apply  | wireguard-vm | apply specialized terraform project for wireguard-vm     |
| ansible-playbook  | run    | wireguard-vm | run specialized ansible playbook for wireguard-vm        |


| Sub-command | Host         | Arguments (minimum one of the following)                           |
|-------------|--------------|--------------------------------------------------------------------|
| check       | proxmox      | terraform:apply ansible-inventory:add ansible-playbook:run ssh:add |
| check       | plex         | terraform:apply ansible-inventory:add ansible-playbook:run ssh:add |
| check       | wireguard-vm | terraform ansible-inventory:add ansible-playbook:run ssh:add       |
| check       | &lt;host&gt; | ansible-inventory:add ansible-playbook:run ssh:add                 |

What would be the best way to structure this?

## Approach: Handlers

All commands are guaranteed to have a host-alias

- Have a "handler" for each host-alias. Each method in the handler is an action which takes an dryRun argument.
- Have a map that maps to those handlers
```go
type DefaultHandler struct{}
func (h defaultHandler) AddAnsibleInventory(dryRun bool) error {}
func (h defaultHandler) RunAnsiblePlaybook(dryRun bool) error {}
func (h DefaultHandler) AddHostToSSHConfig(dryRun bool) error {}

type ProxmoxHandler struct{}
func (h ProxmoxHandler) ApplyTerraform(dryRun bool) error {}
func (h ProxmoxHandler) RunAnsiblePlaybook(dryRun bool) error {}

type PlexHandler struct{}
func (h PlexHandler) ApplyTerraform(dryRun bool) error {}
func (h PlexHandler) RunAnsiblePlaybook(dryRun bool) error {}

type WireguardVM struct{}
func (h WireguardVM) ApplyTerraform(dryRun bool) error {}
func (h WireguardVM) RunAnsiblePlaybook(dryRun bool) error {}

var handlerMap = map[string]any{
  "proxmox": ProxmoxHandler{},
  "plex": PlexHandler{},
  "wireguard-vm": WireguardVM{},
}
```

Implementation:

`ansibleplaybook_run.go`
```go
	ansiblePlaybookRunCmd := &cobra.Command{
    ...
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
      handler := handlerMap[hostAlias] or defaultHandler

      if handler.RunAnsiblePlaybook == null: return ErrInvalidArgs

      return handler.RunAnsiblePlaybook()
		},
```

Another thing i need to support is commands like:
```sh
./bin/labctl check some-host ansible-inventory:add ansible-playbook:run ssh:add
```

How would this work?

`check.go`
```go
	ansiblePlaybookRunCmd := &cobra.Command{
    ...
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
      handler := handlerMap[hostAlias] or defaultHandler

      var diagnostics map[string]string
      var errors []string
      for _, target := range targets {
        subCommand := target.Split(":")[0]
        action := target.Split(":")[1]
        fn := handler.`${subcommand}${action}`
        if fn == null {
          errors.append(`${target} not supported`)
          continue
        }

        diagnostics.merge(fn(true))
      }

      if len(invalidArgs) > 0 {
        return ErrInvalidArgs
      }

      handler := handlerMap[hostAlias] or defaultHandler

      if handler.RunAnsiblePlaybook == null: return ErrInvalidArgs

      return handler.RunAnsiblePlaybook()
		},
```

### Evaluation

Pros:
- allows me to query for supported subcommands easily: `slices.Collect(maps.Keys(handlerMap))`

Cons:
- obviously it's not possible to dynamically create a method name and call it. we could improve this by making each handler have an internal map where the keys are `subcommand:action` like `terraform:apply` and the values are an actual function like `func() error`. This might work, but it's a bit risky because it would break if any future function ends up having a different signature.
- this assumes we can just pass a bool argument called "dryRun" to any handler method and if its true, we just get diagnostics back, and if its false, we get an entirely different return value. but go doesn't quite work like that. I suppose you could create some sort of enum as a return value but that's not ideal because then we have to do a type assertion on the return value.

