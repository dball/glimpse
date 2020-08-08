# Glimpse

Glimpse is a Clojure-inspired LISP interpreter written in Golang.

## Usage

1. `make`
2. `./glimpse`

## Philosophy

It was written using the MAL methodology, but does not intend MAL compatibility.
In particular:

* Map keys may be any mal values
* Strings will seq as lists of characters, not strings
* Metadata must be maps
* The core data abstraction is a Seq, allowing laziness

Intended features include:

* Concurrency using CSP with a core.async compatible syntax
* Some form of general Golang module interop

## Limitations

It's not intended for any use other than pedagogical.
