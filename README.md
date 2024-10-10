# Cartesi Rollups GO high level framework

```
This works for Cartesi Rollups Node version 2.0.x (cartesi cli version ?.?.x)
```

Create cartesi rolllups applications with simple lines of code like:

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

Check the [examples](example-apps/apps-src) for more use cases. 

You will need [cartesi cli](https://github.com/cartesi/cli) to create and run the example, and [curl](https://curl.se/) to interact with the app.

To run an example 

```shell
cd example-apps
rm -rf app.go
ln -sr ln -sr apps-src/example11_wallet.go app.go
```

The you can build and run with

```shell
cartesi build
cartesi run
```

Optionally, you can compile directly in the host and run with [nonodo](https://github.com/Calindra/nonodo/) and the [cartesi machine](https://github.com/cartesi/machine-emulator) (you'll also need `riscv64-linux-gnu-gcc`). First start nonodo:


```shell
GOOS=linux GOARCH=riscv64 CGO_ENABLED=1 CC=riscv64-linux-gnu-gcc go build -o app app.go
nonodo -- cartesi-machine --env=ROLLUP_HTTP_SERVER_URL=http://10.0.2.2:5004 --network --flash-drive=label:root,filename:.cartesi/image.ext2 --volume=.:/mnt --workdir=/mnt -- ./app
```

You can send inputs with (account and private key of anvil test accounts)

```shell
cartesi send generic --input="test"
```

Send inspects with

```shell
curl http://localhost:8080/inspect/test
```

View notices with 

```shell
curl -H 'Content-Type: application/json' -X POST http://localhost:8080/graphql -d '{"query": "query { notices { edges { node { index payload }}}}"}'
```

```
DISCLAIMER: This is a prototype to showcase the Cartesi Rollups features and is not intended to be used as-is in the production environment
```
