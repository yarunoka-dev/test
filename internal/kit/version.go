package kit

// KitVersion is the version of this kit build, stamped into the pasteable
// summary. Its x.y is the spec version the embedded cases target (the kit
// is the one binary that carries them, so the case set decides the number);
// the last digit counts the kit's own fixes within that target.
const KitVersion = "1.0.0"

// SpecVersion is the spec version this kit's cases are authored against.
// The declaration and the summary output are the same constant, so the
// claim cannot drift from what the cases actually check.
const SpecVersion = "1.0"
