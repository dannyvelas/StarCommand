# structure for defining unique behavior for permutations of action, resource, alias, and target

## Problem

There will be many permutations of: action, resource, alias, and target that will be passed to my CLI

I want to define functions for each supported permutation in an extensible way.

## Current solution

Cobra takes care of the permutations of action and resource. So, all I need to do is define one function for each supported action+resource combo:
```
GetConfig(host, targets)
CheckConfig(host, targets)
SetFile(host, targets)
```

Those functions take `host` and `targets` as an argument.

## Why I'm not happy with the current solution

- I'd like to be able to extract supported hosts from my solution like, it'd be cool if I could extract my supported hosts (e.g. `proxmox`, `plex`, `vm`, etc) from a map, or something like this (e.g. `map.Keys(someMap)`.)
- Each host and target combo has a unique function implementation which I just resolve via a `if/else` statement, which maybe isn't the best way to do it. The switch statement can start to get big or crazy:
```
if host == "proxmox" && target == "ansible" {
  return proxmoxAnsible{}
} else if host == "plex" && target == "ansible" {
  return plexAnsible{}
} else if host == "vm" && target == "ansible" {
  return vmAnsible{}
} else if target == "ssh" {
  return sshHost{}
}
```

## Alternative solution: Double dispatch

I'm not sure how this would work with Cobra. Ideally, I'd take advantage of the dispatching powers that Cobra is already giving me for free, but I'm not sure how. I'll show how I think it would work without Cobra and a spec of how it might work with Cobra.

### Without using cobra to resolve action+resource for me

Define multiple "execute" functions, each for a unique permutation:
```
execute(action: Get, resource: Config, target: Ansible, hostAlias: Proxmox)
execute(action: Get, resource: Config, target: Ansible, hostAlias: Plex)
execute(action: Get, resource: Config, target: Ansible, hostAlias: VM)
execute(action: Check, resource: Config, target: Ansible, hostAlias: Proxmox)
execute(action: Check, resource: Config, target: Ansible, hostAlias: Plex)
execute(action: Check, resource: Config, target: Ansible, hostAlias: VM)
execute(action: Get, resource: Config, target: SSH, hostAlias: Proxmox)
execute(action: Get, resource: Config, target: SSH, hostAlias: Plex)
execute(action: Get, resource: Config, target: SSH, hostAlias: VM)
execute(action: Check, resource: Config, target: SSH, hostAlias: Proxmox)
execute(action: Check, resource: Config, target: SSH, hostAlias: Plex)
execute(action: Check, resource: Config, target: SSH, hostAlias: VM)
execute(action: Set, resource: File, target: SSH, hostAlias: Proxmox)
execute(action: Set, resource: File, target: SSH, hostAlias: Plex)
execute(action: Set, resource: File, target: SSH, hostAlias: VM)
```

The right "execute" function will execute based on the action, resource, target, and host-alias.

This would roughly look like this:

`get_config.go`
```
	getConfigCmd := &cobra.Command{
    ...
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]

      var combinedConfigs
      for target in targets {
        temp := execute(get, config, hostAlias, target)
        combinedConfigs.merge(temp)
      }

      print(json.Marshal(combinedConfigs))
		},

```

### With Cobra

Same as above but only dispatch on target and host alias
```
execute(target: Ansible, hostAlias: Proxmox)
execute(target: Ansible, hostAlias: Plex)
execute(target: Ansible, hostAlias: VM)
execute(target: SSH, hostAlias: Proxmox)
execute(target: SSH, hostAlias: Plex)
execute(target: SSH, hostAlias: VM)
```

- Not sure how this would work though. Consider the implementation of the first `execute` function that should only dispatch for `target==Ansible` and `hostAlias==Proxmox`. When you're implementing this function, how would you know what the action and resource are? In other words, how would you know what this function needs to do? Would it be getting the config, checking the config, setting a file?
- I suppose you could have the be methods that have a `action` and `resource` field embedded in the receiver object. But then that just creates the same problem all over again. Now you need to do nested multiple dispatch on `action`+`resource`. At that point just go for the full multiple dispatch solution without Cobra, I would think.

### Evaluation

Pros
- Extensible

Cons
- I don't like having to run multiple `execute` functions in a for-loop when multiple targets are passed in
- Doesn't give us a way to get the list of supported hosts

## Alternative solution: Host handlers

- Have a "handler" for each host. Each method in the handler is an action+resource.
- Have a map that maps to those handlers
```
class ProxmoxHandler {
  GetConfig(targets: []string)
  CheckConfig(targets: []string)
  SetFile(targets: []string)
}
class PlexHandler {
  GetConfig(targets: []string)
  CheckConfig(targets: []string)
  SetFile(targets: []string)
}
class VMHandler {
  GetConfig(targets: []string)
  CheckConfig(targets: []string)
  SetFile(targets: []string)
}

var handlerMap = map[string]any{
  "proxmox": ProxmoxHandler{},
  "plex": PlexHandler{},
  "vm": VMHandler{},
}
```

Implementation:

`get_config.go`
```
	getConfigCmd := &cobra.Command{
    ...
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]

      handler := handlerMap[hostAlias]
      if !handler: return ErrInvalidHost

      if handler.GetConfig == null: return ErrNotSupportedForHost

      return handler.GetConfig(targets)
		},
```

### Evaluation

Pros:
- Extensible
- Can get hosts from thing
- `GetConfig()` abstracts the whole merging configs for multiple targets part

## todo
- in mutliple dispatch, what about making "targets" not be one of the dispatching arguments and just a regular input argument, an array of strings?
