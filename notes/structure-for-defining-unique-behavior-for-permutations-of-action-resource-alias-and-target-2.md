# structure for defining unique behavior for permutations of action, resource, alias, and target: Part II

## Problem statement no.1

The previous note on this was really good. For the requirements that I had, the handler interface approach worked perfectly. But now I have a new requirement for which the handler interface approach falls short: I need to be able to support calling a missing "host-alias". This doesn't work in the approach I have unless I do something hacky like hard-code support for an empty string passed into `handlers.New`. I would have to make an empty string be one of the keys of `handlers.handlerMap`. it's a bit strange. 

I'm wondering if the multiple dispatch approach would work better now because it solved all the requirements that `handlers.handlerMap` solved except for one, being able to query for host names. but maybe this requirement can be dropped if it can give me clean support for missing hosts. Or, maybe it's possible anyway.

## Problem statement no.2

Also, there's another wrinkle in the `handlers.handlerMap` solution. It has a manual fallback where Proxmox tries to match a given "target" with something, and if it doesn't match it calls a function called `fallbackTargetToStruct` which will then have a bunch of functions that are generic to hosts. I feel like in native multiple dispatch, you would basically just have a "base" implementation of `Execute('*': host, ssh: target)` then other implementations that are more specific like `Execute('proxmox: host, ansible: target`)`, and if those specific implementations don't work, then the "base" implementation will execute.

## Problem summary
- Support for wildcard host and wildcard target for `labctl set secret`
- Support for wildcard host for `labctl set file ... --ssh`

## Approach: Multiple dispatch

Goal:
- `execute(get: action, config: resource, proxmox: host, ansible: target) // gets config for proxmox+ansible`
- `execute(set: action, secret: resource, <*>: host, <*>: target) // set secret globally, this runs whenever "set" and "secret" are the first two arguments, regardless of the values passed in to "host" and "target"`
- `execute(set: action, file: resource, <*>: host, ssh: target) // sets ssh file for host, this runs whenever "set" and "file" are the first two arguments, and "ssh" is the last argument, regardless of the value passed in to "host"`

`get_config.go`:
```go
	getConfigCmd := &cobra.Command{
    ...
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
      configMux := (...)

      result, err := handlers.New(configMux, "get", "config", targets)
		},
```

`handlers/handler.go`:
```go
type Rule struct {
  Match  func(host, target string) bool
  Action func(host, target string)
}

func New(configMux, action, resource, targets []string, hostAlias: Option[string]) {
  
}
```

### Evaluation

- How could we have one function signature for "GetConfig"/"SetConfig"/"SetFile" etc when all of these functions have different signatures? (e.g. have different parameters and return values)?

## Multiple dispatch v2 Solution

`get_config.go`
```go
	getConfigCmd := &cobra.Command{
    ...
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
      configMux := //

      a, err := app.New(configMux)
      if err != nil {
        // handle error
      }

      configs, diagnostics, err := a.GetConfig(hostAlias, targets)
      if err != nil {
        // handle error
      }

      // remaining logic
		},
```

`set_secret.go`
```go
	getConfigCmd := &cobra.Command{
    ...
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
      configMux := //

      a, err := app.New(configMux)
      if err != nil {
        // handle error
      }

      if err := a.SetSecret(secret); err != nil {
        // handle error
      }
		},
```

`app/app.go`:
```go
func New(configMux) {
	return App{
		configMux: configMux,
	}
}

func (a App) GetConfig(hostAlias string, targets []string) (map[string]string, map[string]string, error) {
	return injectHandler(hostAlias, a.getConfig)(targets)
}

func (a App) CheckConfig(hostAlias string, targets []string) (map[string]string, error) {
	return injectHandler(hostAlias, a.checkConfig)(targets)
}

func (a App) getConfig(handler Handler, targets []string) (map[string]string, map[string]string, error) {
	return handler.GetConfig(targets)
}

func (a App) checkConfig(handler Handler, targets []string) (map[string]string, error) {
	return handler.CheckConfig(targets)
}

func injectHandler(hostAlias string, fn func(Handler, []string) ? ) func(targets []string) ? {
  handler, ok := handlerMap[hostAlias]
  if !ok {
    // handle error
  }

  return func(targets []string) ? {
    return fn(handler, targets)
  }
}
```

### Evaluation

- The `injectHandler` pattern seems nice to share middleware between `GetConfig` and `CheckConfig`, but unfortunately since these have different return types, i'm not sure how it would work, unless i just forced both of them to get the same return type.
- I guess I theoretically could do this. But what about SetFile?
