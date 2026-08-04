package kit

import "encoding/json"

// Request is what the runner writes to an adapter's stdin: the document,
// and for eval the query (with bindings when the case uses resolver-backed
// names). The document rides as raw JSON — the runner never interprets it,
// it only carries it.
type Request struct {
	Action   string              `json:"action"`
	Document json.RawMessage     `json:"document"`
	Query    *Query              `json:"query,omitempty"`
	Bindings map[string][]string `json:"bindings,omitempty"`
}

// Query names one of the three queries of the spec's evaluation model.
// The field names follow the spec's own wording: a point is asked "at",
// a period runs "after … through", an enumeration "from … through".
type Query struct {
	Type    string `json:"type"`
	At      string `json:"at,omitempty"`
	After   string `json:"after,omitempty"`
	Through string `json:"through,omitempty"`
	From    string `json:"from,omitempty"`
}

// Response is what an adapter answers on stdout. Exactly one of the three
// shapes is present: a result (a judgment boolean or an enumeration list),
// an emitted document, or the invalid flag.
type Response struct {
	Result   json.RawMessage `json:"result,omitempty"`
	Document json.RawMessage `json:"document,omitempty"`
	Invalid  bool            `json:"invalid,omitempty"`
}
