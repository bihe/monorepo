# Monorepo

The monorepo is a collecton of services used for login, documents and bookmarks

[![codecov](https://codecov.io/gh/bihe/monorepo/branch/main/graph/badge.svg)](https://codecov.io/gh/bihe/monorepo)

<hr/>

## Golang cross-compile with CGO
Golang has the neat funcitionality to cross-compile to different target architectures. This is very useful and very easy by just set env-vars `GOOS` and `GOARCH`. There are cases when the process of cross-compilation is getting complicated. Specifically when [CGO](https://pkg.go.dev/cmd/cgo) is in the mix. On Linux this is typically quite easy, but when on different architectures this can be a challenge.

When using sqlite it is necessary to use `CGO_ENABLED`. To cross-compile from a Mac source-system to e.g. Linux there is a nice support which is described in this blog-post: https://www.yellowduck.be/posts/cross-compile-a-go-package-which-uses-sqlite3

1. Install `musl-cross` -> "One-click static-friendly musl-based GCC macOS-to-Linux cross-compilers" (https://github.com/FiloSottile/homebrew-musl-cross)
   ```bash
   brew install FiloSottile/musl-cross/musl-cross
   ```
2. Compile by specifying `CC` and `CXX`
   ```bash
   CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" ./...
   ```

The result is the correct platform-specific build: ```ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked```
