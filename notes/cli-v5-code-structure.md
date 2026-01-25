## Problem

The command structure changed to be this:

| Sub-command       | Action | Host-Alias   | Description                                              |
|-------------------|--------|--------------|----------------------------------------------------------|
| ansible-inventory | add    | &lt;host&gt; | default logic to add a host to an ansible inventory file |
| ansible-playbook  | run    | &lt;host&gt; | run default ansible playbook for host &lt;host&gt;       |
| *ssh               | add    | &lt;host&gt; | default logic to add a host to home ssh config file      |
| terraform         | apply  | proxmox      | apply specialized terraform project for proxmox          |
| *ansible-playbook  | run    | proxmox      | run specialized ansible playbook for proxmox             |
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

func (h DefaultHandler) AddAnsibleInventory(dryRun bool) error {}
func (h DefaultHandler) RunAnsiblePlaybook(dryRun bool) error  {}
func (h DefaultHandler) AddHostToSSHConfig(dryRun bool) error  {}

type ProxmoxHandler struct{}

func (h ProxmoxHandler) ApplyTerraform(dryRun bool) error     {}
func (h ProxmoxHandler) RunAnsiblePlaybook(dryRun bool) error {}

type PlexHandler struct{}

func (h PlexHandler) ApplyTerraform(dryRun bool) error     {}
func (h PlexHandler) RunAnsiblePlaybook(dryRun bool) error {}

type WireguardVMHandler struct{}

func (h WireguardVMHandler) ApplyTerraform(dryRun bool) error     {}
func (h WireguardVMHandler) RunAnsiblePlaybook(dryRun bool) error {}

var handlerMap = map[string]any{
	"proxmox":      ProxmoxHandler{},
	"plex":         PlexHandler{},
	"wireguard-vm": WireguardVMHandler{},
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

      fmt.Println(format(diagnostics))
		},
```

### Evaluation

Pros:
- allows me to query for supported subcommands easily: `slices.Collect(maps.Keys(handlerMap))`

Cons:
- obviously it's not possible to dynamically create a method name and call it. we could improve this by making each handler have an internal map where the keys are `subcommand:action` like `terraform:apply` and the values are an actual function like `func() error`. This might work, but it's a bit risky because it would break if any future function ends up having a different signature.
- this assumes we can just pass a bool argument called "dryRun" to any handler method and if its true, we just get diagnostics back, and if its false, we get an entirely different return value. but go doesn't quite work like that. I suppose you could create some sort of enum as a return value but that's not ideal because then we have to do a type assertion on the return value.

## Approach: Handlers V2

- Have a "handler" for each host-alias. Each method in the handler has one method for each "action" (e.g. AddAnsibleInventory method corresponds to "add ansible" action).
- Have a map at the top level called `handlerMap` that maps to those handlers
- Each method in each function has a unique "get*Diagnostics" function
- Each handler will have a "ActionToDiagnostics" method that will take a string like "ansible-inventory:add" and return the function which is used to get diagnostics for `AddAnsibleInventory`.
```go
type DefaultHandler struct{
  actionToDiagnosticsMap: map[string]func() (map[string]string, error)
  /*
  {
    "ansible-inventory:add": h.getAnsibleAddDiagnostics,
    "ansible-playbook:run": h.getAnsiblePlaybookRunDiagnostics,
    "ssh:add": h.getSSHAddDiagnostics,
  }
  */
}
func (h DefaultHandler) AddAnsibleInventory() error {}
func (h DefaultHandler) RunAnsiblePlaybook() error {}
func (h DefaultHandler) AddHostToSSHConfig() error {}
func (h DefaultHandler) ActionToDiagnostics(action string) (func() (map[string]string, error), bool) {
  return h.actionToDiagnosticsMap[action]
}

type ProxmoxHandler struct{}
func (h ProxmoxHandler) ApplyTerraform() error {}
func (h ProxmoxHandler) RunAnsiblePlaybook() error {}
func (h ProxmoxHandler) ActionToDiagnostics(action string) (func() (map[string]string, error), bool) {
  return h.actionToDiagnosticsMap[action]
}

type PlexHandler struct{}
func (h PlexHandler) ApplyTerraform() error {}
func (h PlexHandler) RunAnsiblePlaybook() error {}
func (h PlexHandler) ActionToDiagnostics(action string) (func() (map[string]string, error), bool) {
  return h.actionToDiagnosticsMap[action]
}

type WireguardVMHandler struct{}
func (h WireguardVMHandler) ApplyTerraform() error {}
func (h WireguardVMHandler) RunAnsiblePlaybook() error {}
func (h WireguardVMHandler) ActionToDiagnostics(action string) (func() (map[string]string, error), bool) {
  return h.actionToDiagnosticsMap[action]
}

var handlerMap = map[string]any{
  "proxmox": ProxmoxHandler{},
  "plex": PlexHandler{},
  "wireguard-vm": WireguardVMHandler{},
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

Implementation:

`check.go`
```go
	ansiblePlaybookRunCmd := &cobra.Command{
    ...
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
      handler := handlerMap[hostAlias] or defaultHandler

      var allDiagnostics map[string]string
      var errors []string
      for _, action := range actions {
        diagnosticFn := handler.ActionToDiagnostics(action)
        if diagnosticFn == null {
          errors.append(`${action} not supported`)
          continue
        }

        diagnostics, err := diagnosticFn()
        if err != nil {
          // handle error
        }

        allDiagnostics.merge(diagn)
      }

      if len(invalidArgs) > 0 {
        return ErrInvalidArgs
      }

      fmt.Println(format(allDiagnostics))
		},
```

### Evaluation

Pros:
- allows me to query for supported subcommands easily: `slices.Collect(maps.Keys(handlerMap))`
- allows us to dynamically query for the correct diagnostic function

## Approach: Handlers V3

- Have a "handler" for each host-alias. Each method in the handler has one method for each "action" (e.g. AddAnsibleInventory method corresponds to "add ansible" action).
- Have a map at the top level called `handlerMap` that maps to those handlers
- Each method returns a struct that corresponds to its action (e.g. AddAnsibleInventory will return a `AnsibleInventoryAdd` struct, RunAnsiblePlaybook will return a `AnsiblePlaybookRun` struct). These structs that are returned from methods are called "Action" structs.
```go
type DefaultHandler struct{}

func (h DefaultHandler) AddAnsibleInventory() AnsibleInventoryAdd {}
func (h DefaultHandler) RunAnsiblePlaybook() AnsiblePlaybookRun   {}
func (h DefaultHandler) AddHostToSSHConfig() SSHAdd               {}

type ProxmoxHandler struct{}

func (h ProxmoxHandler) ApplyTerraform() TerraformApply         {}
func (h ProxmoxHandler) RunAnsiblePlaybook() AnsiblePlaybookRun {}

type PlexHandler struct{}

func (h PlexHandler) ApplyTerraform() TerraformApply         {}
func (h PlexHandler) RunAnsiblePlaybook() AnsiblePlaybookRun {}

type WireguardVMHandler struct{}

func (h WireguardVMHandler) ApplyTerraform() TerraformApply         {}
func (h WireguardVMHandler) RunAnsiblePlaybook() AnsiblePlaybookRun {}

var handlerMap = map[string]any{
	"proxmox":      ProxmoxHandler{},
	"plex":         PlexHandler{},
	"wireguard-vm": WireguardVMHandler{},
}
```

The action structs will look something like this:
```go
type AnsibleInventoryAdd struct {
	Run            func() error
	GetDiagnostics func() (map[string]string, error)
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

      return handler.RunAnsiblePlaybook().Run()
		},
```

Another thing i need to support is commands like:
```sh
./bin/labctl check some-host ansible-inventory:add ansible-playbook:run ssh:add
```

Implementation:

- Ahh wait... this won't work because i would need to somehow call `handler."ansible:run".Check()`

## Approach: simulating multiple dispatch with registry pattern

Implementation

```go
type Action string

const (
	ApplyAction Action = "action"
	AddAction   Action = "add"
  ...
)

type Resource string

const (
  AnsiblePlaybook  Resource = "ansiblePlaybook"
  AnsibleInventory Resource = "ansibleInventory"
  terraform        Resource = "terraform"
  ...
)

var Registry = []Rule{
	// SPECIALIZED: Proxmox Terraform
	{
		Name: "Proxmox Terraform",
		Match: func(action Action, resource Resource, host string) bool {
			return action == ApplyAction && resource == terraform && host == "proxmox"
		},
		Execute: func(host string, dryRun bool) Result {
			if dryRun {
				return Result{Output: map[string]string{"terraform": "Plan: Create VM"}}
			}
			// ... Actual Logic
			return Result{}
		},
	},
	// DEFAULT: Ansible Playbook (matches any host if action is ansible)
	{
		Name: "Default Ansible",
		Match: func(action Action, resource Resource, host string) bool {
			return action == runAction && resource == AnsiblePlaybook
		},
		Execute: func(host string, dryRun bool) Result {
			// ... Logic
			return Result{}
		},
	},
}

func Execute(action Action, resource Resource, host string, dryRun bool) {
	for _, rule := range Registry {
		if rule.Match(action, resource, host) {
			rule.Action(host, dryRun)
			return // Run the first one that matches
		}
	}
} 
```

I would call it like this:

`ansibleplaybook_run.go`
```go
    ansiblePlaybookRunCmd := &cobra.Command{
        Run: func(cmd *cobra.Command, args []string) {
            hostAlias := args[0]
            return Execute(RunAction, AnsiblePlaybook, hostAlias, false)
        },
```

And for the check command, I'd do something like:

`check.go`
```go
    ansiblePlaybookRunCmd := &cobra.Command{
        Run: func(cmd *cobra.Command, args []string) {
            hostAlias := args[0]
            handler := handlerMap[hostAlias] or defaultHandler

            var allDiagnostics map[string]string
            var errors []string
            for _, target := range targets {
              split := target.Split(':')
              resource := split[0]
              action := split[1]
              diagnostics, err := Execute(action, resource, hostAlias, true)
              if err != nil {
                // handle error
              }

              allDiagnostics.merge(diagn)
            }

            if len(invalidArgs) > 0 {
              return ErrInvalidArgs
            }

            fmt.Println(format(allDiagnostics))
        },
    }
```

## Approach C: simulating multiple dispatch with radix tree

```go
type Action string

const (
	ApplyAction Action = "action"
	AddAction   Action = "add"
  ...
)

type Resource string

const (
  AnsiblePlaybook  Resource = "ansiblePlaybook"
  AnsibleInventory Resource = "ansibleInventory"
  Terraform        Resource = "terraform"
  ...
)

// tree.go
type Node struct {
  edges map[string]*Node
  impl func(host string, dryRun bool) // if not nil, then this is the end node
}

func (t *Tree) findNode(action Action, resource Resource, host string) {
  currNode := t.Root

  // first see if we have an edge from root to action node, if so go to action node.
  // next, see if we have an edge from action node to resource node. if so, go to resource node.
  // finally, see if we have an edge from resource ndoe to host node. if so, go to host node.

  // if any of these steps fail, return error: invalid arguments

  // once we're at host node, just run it
}

func (a App) Execute(action Action, resource Resource, host string, dryRun bool) {
  node, err := a.tree.findNode(action, resource, host)
  if err != nil {
    // handle error
  }

  node.impl(host, dryRun)
} 
```



I would call it like this:

`ansibleplaybook_run.go`
```go
    ansiblePlaybookRunCmd := &cobra.Command{
        Run: func(cmd *cobra.Command, args []string) {
            hostAlias := args[0]
            app := app.New()
            app.Execute(app.RunAction, app.AnsiblePlaybook, hostAlias, false)
        },
```

And for the check command, I'd do something like:

`check.go`
```go
    ansiblePlaybookRunCmd := &cobra.Command{
        Run: func(cmd *cobra.Command, args []string) {
            hostAlias := args[0]
            handler := handlerMap[hostAlias] or defaultHandler

            var allDiagnostics map[string]string
            var errors []string
            for _, target := range targets {
              split := target.Split(':')
              resource := split[0]
              action := split[1]
              diagnostics, err := app.Execute(app.Action(action), app.Resource(resource), hostAlias, true)
              if err != nil {
                // handle error
              }

              allDiagnostics.merge(diagn)
            }

            if len(invalidArgs) > 0 {
              return ErrInvalidArgs
            }

            fmt.Println(format(allDiagnostics))
        },
    }
```

