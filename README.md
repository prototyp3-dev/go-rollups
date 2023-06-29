# Cartesi Rollups GO high level framework

Create cartesi rolllups DApp with codes like:

```go
package main

import (
  "fmt"
  "github.com/prototyp3-dev/go-rollups/handler"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

func Handle(payload string) error {
  fmt.Println("Handle: Received payload:",payload)
  return nil
}

func main() {
  handler.HandleDefault(Handle)

  err := handler.RunDebug()
  if err != nil {
    log.Panicln(err)
  }
}
```

Check the [examples](examples) for more use cases. 

You will need [sunodo cli](https://github.com/sunodo/sunodo/tree/main/apps/cli) to create and the example. and [curl](https://curl.se/) to interact with the dapp.

To run an example 

```shell
cd examples
rm -rf example.go
ln -sr example1.go example.go
sunodo build
sunodo run
```

You can send inputs with (account and private key of anvil test accounts)

```shell
sunodo send generic --input="test"
```

Send inspects with

```shell
curl http://localhost:8080/inspect/test
```

Send notices with 

```shell
curl -H 'Content-Type: application/json' -X POST http://localhost:8080/graphql -d '{"query": "query { notices { edges { node { index payload }}}}"}'
```
