# The adapter protocol

The kit talks to an implementation through an **adapter**: a small
executable the implementer writes. The runner starts the adapter once
per case, writes one request as JSON to its stdin, and reads one
response as JSON from its stdout. That is the whole contract — how the
cases are stored inside the kit is not part of it.

An adapter is thin wiring by principle: it hands what it receives to the
implementation **unvalidated and unmodified**. Some cases deliberately
carry broken input to check that the implementation itself rejects it; an
adapter that "helpfully" validates or normalizes on the way in would
absorb exactly what those cases exist to detect.

The expectations and the case names never reach the adapter. Judging is
the runner's job, and an adapter that cannot see which case it is running
cannot special-case one.

## Request

```json
{
  "action": "eval",
  "document": { "version": "1.0", "timezone": "Asia/Tokyo", "schedules": [ ... ] },
  "query": { "type": "point", "at": "2026-07-27T10:00:00+09:00" },
  "bindings": { "company-closures": ["2026-08-05"] }
}
```

- `action` — `"eval"` or `"emit"`. An emit request carries no `query`.
- `document` — a Yrnk document, embedded as the JSON value it is.
- `query` — one of the three queries of the spec's evaluation model.
  The field names follow the spec's own wording:

  | `type` | Fields | Asks |
  |---|---|---|
  | `"point"` | `at` | is the given instant an occurrence? |
  | `"period"` | `after`, `through` | is there an occurrence after `after`, through `through`? |
  | `"enumeration"` | `from`, `through` | which occurrences lie from `from` through `through`? |

- `bindings` — present when the document declares resolvers: a map of
  resolver name to a list of date literals (`YYYY-MM-DD`). The adapter
  registers each as a resolver that returns the list as-is (the
  pass-through above). Emit requests carry bindings too — parsing a
  document that declares resolvers needs them, and emit starts with a
  parse.

Query instants are **ISO 8601 with a UTC offset** (RFC 3339), never a
zone name: the wire carries moments, and the tz-database name's job is
done by the document's own `timezone` declaration.

## Response

Answer with exactly one of the three shapes:

```json
{ "result": true }
{ "result": ["2026-07-28", "2026-07-28T00:00:00+09:00"] }
{ "document": { ...the re-emitted document... } }
{ "invalid": true }
```

- **A judgment** (`point`, `period`) answers `result` with a boolean.
- **An enumeration** answers `result` with the occurrence list in
  ascending order. An all-day occurrence is a date (`YYYY-MM-DD`); a
  timed occurrence is an instant with an offset. The spelling tells the
  two kinds apart, and they never merge. Instants are compared as
  moments, so the offset your language's formatter picks is fine.
- **An emit request** answers `document` with the document parsed and
  re-emitted by your serializer. The comparison is structural — JSON key
  order and whitespace do not matter — against the authored spelling
  (round-tripping a valid document is the identity in this language).
- **An invalid document** answers `invalid` — for any action, and
  nothing more. No error codes, no messages: reporting *why* a case
  expected rejection is the runner's job, from the case's authoring
  metadata.

Answering invalid is a **normal answer**, delivered with exit status 0.
A crash, a non-zero exit, or non-JSON output is adapter breakage: the
runner reports it apart from test results, as infrastructure trouble
rather than a FAIL.

## What the runner does with it

```text
for each case:
    start the adapter (argv from the command line, verbatim)
    write the request JSON to its stdin, close it
    read its stdout, wait for exit
    judge the answer against the authored expectation
report failures, then the summary
```

The runner exits non-zero unless every case passed, so the CI line

```console
$ yarunoka-test eval php adapter.php
```

is the whole integration.
