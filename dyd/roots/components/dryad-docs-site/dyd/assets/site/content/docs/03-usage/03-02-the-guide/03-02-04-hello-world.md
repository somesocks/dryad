---
title: "03.02.04 - Hello, World!"
description: "Building and testing your garden."
type: docs
layout: single
---

# Hello, World!

Now that we have a working go compiler and a source package, we can implement our hello, world webserver.

In `dyd/roots/server/dyd/assets/main.go`, we can add a simple webserver:

```go
package main

import (
  "fmt"
  "net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
  fmt.Println("request received")
  fmt.Fprintf(w, "Hello, World!")
}

func main() {
  http.HandleFunc("/", hello)
  http.ListenAndServe(":8080", nil)
}

```

We also want to add a skeleton `go.mod` file for the compiler.  In `dyd/roots/server/dyd/assets/go.mod`, add:

```
module hello-world

go 1.19

require (
)

```

Now that those are in place, we want to add our go root as a dependency for the server root.  We can do so by navigating to our server root and using `dryad root link`: `cd ./dyd/roots/server && dryad root link .tools/go`.

Finally, we want to update our build script for the root to actually build our server.  In `dyd/roots/server/dyd/commands/default`:

```sh
#!/usr/bin/env sh

SRC_DIR=$DYD_STEM
DEST_DIR=$1
DEST_MAIN=$DEST_DIR/dyd/commands/default

BUILD_VERSION="0.0.1"
BUILD_FINGERPRINT="$(cat $SRC_DIR/dyd/fingerprint)"


# add package traits
mkdir -p $DEST_DIR/dyd/traits
echo -n "hello-world-server" > $DEST_DIR/dyd/traits/name
echo -n "$BUILD_VERSION" > $DEST_DIR/dyd/traits/version
echo -n "$DYD_OS" > $DEST_DIR/dyd/traits/os
echo -n "$DYD_ARCH" > $DEST_DIR/dyd/traits/arch


# run go build for specified arch
# diable workspace mode to avoid "missing module" errors
cd $SRC_DIR/dyd/assets/ && \
GOARCH="$DYD_ARCH" \
GOOS="$DYD_OS" \
GOWORK=off \
go build \
  -ldflags "-X=main.Version=$BUILD_VERSION -X=main.Fingerprint=$BUILD_FINGERPRINT" \
  -o $DEST_MAIN

```

With these in place, `dryad garden build` should build successfully, and `dryad sprouts run --include=server` should start our server.  We can visit `localhost:8080` in the browser to confirm.

