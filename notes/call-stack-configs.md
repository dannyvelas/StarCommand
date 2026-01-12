

* fullConfigReader.DryRun()
  * UnmarshalInto(fullConfigReader, hostConfig)
    * fullConfigReader.ReadUnvalidated()
      * bitwardenSecretReader.ReadUnvalidated()
        * validateConfig(bitwardenConfig)
        * RETURNS map, newErrInvalidFields
      * RETURNS nil, newErrInvalidFields

* suppose that there are bitwarden creds missing.
* when this happens we want to print a table to the CLI showing:
  * the bitwarden creds missing + the hostConfig fields missing
* there are two approaches to do this:
  * approach 1: making `validateConfig` have a `map, error` return type.
    * suppose that there are bitwarden creds missing and the `--dry-run` flag is used.
      * `fullConfigReader.DryRun()` will call:
      * `UnmarshalInto(fullConfigReader, hostConfig)` which calls
      * `fullConfigReader.ReadUnvalidated()` which calls
      * `bitwardenSecretReader.ReadUnvalidated()` which calls
      * `validateConfig(bitwardenConfig)`
      * when bitwardenSecretReader.ReadUnvalidated() calls `validateConfig`, `validateConfig` will return a `map, error`.
        * the keys of the `map` will be the bitwardenConfig keys. the values of the `map` will be a message indicating whether the corresponding key was found, not found, or invalid. let's call this map the `keyDiagnosticMap`
        * the `error` will be `ErrInvalidFields`
      * next, `bitwardenSecretReader.ReadUnvalidated()` will return `nil, err` it WONT forward the `keyDiagnosticMap`. it can't. why not? because `bitwardenSecretReader.ReadUnvalidated()` is expected to return a `map, error` where the `map` where the keys are the keys of the bitwarden secrets found and the values are the values of the bitwarden secrets found. it can't return the `keyDiagnosticMap` in its place because the `keyDiagnosticMap` is not expected to be returned from that function.
      * so that means that `fullConfigReader.ReadUnvalidated` will also return `nil, err`
      * which means that `UnmarshalInto(fullConfigReader, hostConfig)`, which has a return type of `error` will also just return the error.
      * which means that `fullConfigReader.DryRun()` gets the `ErrInvalidFields` value but not the diagnostic map so it can't merge the diagnostic map of the missing credentials to the `diagnosticMap` that gets created in a subsequent call to `validateConfig`
  * approach 2: making `validateConfig` have an `error` return type and making that error embed the map.
    * suppose that there are bitwarden creds missing and the `--dry-run` flag is used. also suppose there are no unexpected internal errors.
      * `fullConfigReader.DryRun()` will call:
      * `UnmarshalInto(fullConfigReader, hostConfig)` which calls
      * `fullConfigReader.ReadUnvalidated()` which calls
      * `bitwardenSecretReader.ReadUnvalidated()` which calls
      * `validateConfig(bitwardenConfig)`
      * when bitwardenSecretReader.ReadUnvalidated() calls `validateConfig`, `validateConfig` will return an `ErrInvalidFields` error. this error will have the `keyDiagnosticMap` embedded into it
      * next, `bitwardenSecretReader.ReadUnvalidated()` will return `nil, err`.
      * so that means that `fullConfigReader.ReadUnvalidated` will also return `nil, err`
      * which means that `UnmarshalInto(fullConfigReader, hostConfig)`, which has a return type of `error` will also just return the error.
      * which means that `fullConfigReader.DryRun()` gets the `ErrInvalidFields` value with the diagnostic map so it can merge the diagnostic map of the missing credentials to the `diagnosticMap` that gets created in a subsequent call to `validateConfig`
    * suppose that there are NOT bitwarden creds missing and the `--dry-run` flag is used. also suppose there are no unexpected internal errors.
      * `fullConfigReader.DryRun()` will call:
      * `UnmarshalInto(fullConfigReader, hostConfig)` which calls
      * `fullConfigReader.ReadUnvalidated()` which calls
      * `bitwardenSecretReader.ReadUnvalidated()` which calls
      * `validateConfig(bitwardenConfig)`
      * when bitwardenSecretReader.ReadUnvalidated() calls `validateConfig`, `validateConfig` will `nil`. This is because there are no invalid fields or internal errors.
      * This means that the straight-up map of Bitwarden secrets to values will be returned from `bitwardenSecretReader.ReadUnvalidated()`. 
      * which means that `fullConfigReader.ReadUnvalidated()` returns a straight-up map of configs to values
      * which means that `UnmarshalInto` just fills `hostConfig` with the values it could find and returns a `nil` error.
      * which means that `UnmarshalInto` didn't return any `keyDiagnosticMap` that can be used to merge with the subsequent call to `validateConfig`


* we need something that:
  1. in the case that bitwarden creds are missing, can persist through `ReadUnvalidated` and `UnmarshalInto`.
  2. in the case that bitwarden creds are there, will still be returned from `ReadUnvalidated` and `UnmarshalInto`

* the problem with errors is that they are good for persisting through both `ReadUnvalidated` and `UnmarshalInto` (no.1). but since they are used to signal whether or not something went wrong, they don't allow the map to persist in the case that bitwarden creds are there

* realistically, we might have to change the signature of `ReadUnvalidated` and `UnmarshalInto` so that maybe the become more "persisting-friendly". so that:
  * when bitwarden creds are missing, they return a `keyDiagnosticMap` (like an error with an embedded map)
  * when bitwarden creds are there, they still return a `keyDiagnosticMap` (like a non-error)

* of course, `ReadUnvalidated` still needs to return a `map[string]string` (for its configs) and an `error` in case an internal error happens. Naïvely, for it to become more persisting-friendly, it would have to have a return value that looks something like this: `(map[string]string, map[string]string, error)`, where the first value are its configs, the second is the `keyDiagnosticMap` and the third is just an error. but the problem with that is that `ReadUnvalidated` is a function that exists to satisfy an interface, and 99% of structs that implement this interface shouldn't have a `ReadUnvalidated` function that returns a `keyDiagnosticMap`. this is because 99% of structs that implement this interface just read or return an internal error. they don't have a concept of needing to return an "invalid error."

* So two approaches i'm thinking:
  * making `s` and `t` actually not implement the `readUnvalidated` interface because it's kind of different to the structs that implement that interface
    * if we do this though, how would `UnmarshalInto` be able to work for `s` and `t`?
    * we could make `s` and `t` implement a slightly different interface i guess and `UnmarshalInto` will just dynamically check if it implements one or the other. and if it does it will call the one it implements.
    * But `UnmarshalInto` will have to always have to return the additional `map[string]string` even if its calling a struct from the 99% which doesn't need to return the additional map, which is kind-of ugly, so i don't really like this approach
  * making the `ReadUnvalidated` function in the interface actually accept like a "context" variable that gives strange `readUnvalidated` implementers some more freedom to diverge
    * if we do this though, can a `context` variable actually allow a `keyDiagnosticMap` to persist both in cases where an error happened and where an error didn't happen?
  * making the `ReadUnvalidated` function in the interface actually return like an `(enum, error)`, or more accurately in go,  `(<some-interface>, error)` where `<some-interface>` is guaranteed to have a `config()` function which returns the first map and could optionally also have a `getSecondMap()` function inside of it
    * and then `UnmarshalInto()` would basically return a `(any, error)` and in the two places where `UnmarshalInto()` is called and the second map is expected, the caller, we would check if `any` implements `getSecondMap()`, and if it does, it would use it for the diagnostics.


## gemini question
I have a question:

I have a bunch of structs that implement this interface:
```
type unvalidatedReader interface {
	ReadUnvalidated() (map[string]string, error)
}
```

There is a function called `UnmarshalInto` that takes a struct that implements `unvalidatedReader` as the first argument and calls the `ReadUnvalidated()` function of that struct. `UnmarshalInto` is called in multiple places for multiple structs. `UnmarshalInto` does some stuff but eventually forwards the return values (`(map[string]string, error)`) that were received from `unvalidatedReader.ReadUnvalidated()`.

For two structs only, lets call them `s` and `t`, I'm realizing that the `ReadUnvalidated` function of these structs needs to actually return something that looks more like `(map[string]string, map[string]string, error)`. In other words, there are two places where `UnmarshalInto` is called and the caller actually expects a `map[string]string`, in addition to the one that is already returned.

I could theoretically just change my interface to look like this:
```
type unvalidatedReader interface {
  ReadUnvalidated() (map[string]string, map[string]string, error)
}
```

but i don't really want to do this because 99% of structs will have a `ReadUnvalidated` function that will just be returning `nil` for the middle value. also, that middle value is really not relevant or makes sense for those structs. so, i don't want to change all those structs because of this 1% case.

i'm thinking that maybe i could make make my interface look like this:
```
type unvalidatedReader interface {
	ReadUnvalidated(ctx context.Context) (map[string]string, error)
}
```

and make all the structs that implement this interface just take a `ctx` argument that they ignore. The `ReadUnvalidated` implementation of `s` and `t` will actually use this `ctx` argument to store the `map[string]string`. I'll also make the `UnmarshalInto` function receive a `ctx` argument. That way, in those two places where `UnmarshalInto` is called and the caller expects the addition `map[string]string`, it can just reach into the `ctx` value that it passed and query for it. 

do you think this is a good solution to this problem?

another approach i was considering:
  * making the `ReadUnvalidated` function in the interface actually return like an `(enum, error)`, or more accurately in go,  `(<some-interface>, error)` where `<some-interface>` is guaranteed to have a `config()` function which returns the first map and could optionally also have a `getSecondMap()` function inside of it
    * and then `UnmarshalInto()` would basically return a `(any, error)` and in the two places where `UnmarshalInto()` is called and the second map is expected, the caller, we would check if `any` implements `getSecondMap()`, and if it does, it would use it for the diagnostics.
