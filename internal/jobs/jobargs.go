package jobs

// JobArgs is an interface that represents the arguments for a job of type T.
// These arguments are serialized into JSON and stored in the database.
//
// The struct is serialized using `encoding/json`. All exported fields are
// serialized, unless skipped with a struct field tag.
