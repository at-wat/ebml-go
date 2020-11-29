# ebml-go

[![GoDoc](https://godoc.org/github.com/at-wat/ebml-go?status.svg)](http://godoc.org/github.com/at-wat/ebml-go) ![ci](https://github.com/at-wat/ebml-go/workflows/ci/badge.svg) [![codecov](https://codecov.io/gh/at-wat/ebml-go/branch/master/graph/badge.svg)](https://codecov.io/gh/at-wat/ebml-go) [![Go Report Card](https://goreportcard.com/badge/github.com/at-wat/ebml-go)](https://goreportcard.com/report/github.com/at-wat/ebml-go) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## A pure Go implementation of bi-directional EBML encoder/decoder

EBML (Extensible Binary Meta Language) is a binary and byte-aligned format that was originally developed for the Matroska audio-visual container.
See https://matroska.org/ for details.

This package implements EBML Marshaler and Unmarshaler for Go.
Currently, commonly used elements of WebM subset is supported.


## Usage

Check out the examples placed under [./examples](./examples/) directory.

API is documented using [GoDoc](http://godoc.org/github.com/at-wat/ebml-go).
EBML can be `Marshal`-ed and `Unmarshal`-ed between tagged struct and binary stream through `io.Reader` and `io.Writer`.


## References

- [Matroska Container Specifications](https://matroska.org/technical/specs/index.html)
- [WebM Container Guidelines](https://www.webmproject.org/docs/container/)


## License

This package is licensed under [Apache License Version 2.0](./LICENSE).
