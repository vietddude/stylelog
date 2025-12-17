# slogx

Lightweight `log/slog` helper that:

- routes **Info/Debug/Warn** to a compact handler (no source info)
- routes **Error+** to a handler with source locations
- highlights error attributes using [`tint`](https://github.com/lmittmann/tint)
- optionally sets the logger as the global default, so other packages can just use `slog.*`

## Install

```bash
go get github.com/vietddude/slogx
```

## Usage

```go
package main

import (
	"errors"

	"github.com/vietddude/slogx/logger"
)

func main() {
	// Create and register the default logger.
	log := logger.InitDefault()

	log.Debug("Debug message", "detail", "something")
	log.Info("User login", "user", "alice", "ip", "127.0.0.1")
	log.Warn("High memory usage", "usage", "85%")

	err := errors.New("connection refused")
	log.Error("Failed to load", "err", err)
}
```

Any other package can now just use the global `slog`:

```go
import (
	"errors"
	"log/slog"
)

func doSomething() {
	testErr := errors.New("test error")
	slog.Error("test error", "error", testErr)
}
```


