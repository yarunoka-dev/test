# yarunoka-test

The conformance test kit for [Yrnk](https://github.com/yarunoka-dev/spec),
the Yarunoka schedule DSL: the authored conformance cases and the runner
that judges them, in one binary.

The kit checks that an implementation reads Yrnk documents with the
meaning the spec defines. The implementer writes a single small
**adapter** — an executable that receives a document and a query on
stdin and answers on stdout — and the kit does the rest: it runs every
embedded case against the adapter, judges the answers, and reports what
failed and why. See [docs/protocol.md](docs/protocol.md) for the
adapter contract.

> **Status**: pre-release. The runner works, but the embedded cases are
> a small smoke set; the authored coverage that a release can claim
> conformance against is still being written.

## Usage

```console
$ yarunoka-test eval php adapter.php
$ yarunoka-test emit php adapter.php
$ yarunoka-test all  php adapter.php
```

Everything after the mode is the adapter command, passed through
verbatim and started once per case. The runner exits non-zero unless
every case passes, so one line integrates it into any CI.

The mode is required:

- **eval** — the three queries of the evaluation model (the judgment at
  a point, the judgment over a period, the enumeration)
- **emit** — the round-trip spelling check: parse, re-emit, compare
  against the canonical spelling. For implementations with a serializer
- **all** — both

Passing eval and passing emit are independent claims — an
evaluation-only implementation ships with eval alone, and "kit vX.Y
eval passed" is a complete statement. The summary printed after a run
(kit version, targeted spec version, case count, per-mode tally) is
made to be pasted wherever that claim is stated.

Passing the kit does not certify an implementation — a finite case set
cannot. The claim it supports is exactly "this kit version, passed".

## Install

Prebuilt binaries per OS and architecture come with each
[release](https://github.com/yarunoka-dev/test/releases). Building from
source works the usual way:

```console
$ go install github.com/yarunoka-dev/test/cmd/yarunoka-test@latest
```

## License

[MIT](LICENSE)
