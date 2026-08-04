package kit

// KitVersion is the version of this kit build, stamped into the pasteable
// summary. The dev value marks unreleased builds; releases set the real
// number here.
const KitVersion = "0.0.0-dev"

// SpecVersion is the spec version this kit's cases are authored against.
// The declaration and the summary output are the same constant, so the
// claim cannot drift from what the cases actually check.
const SpecVersion = "1.0"
