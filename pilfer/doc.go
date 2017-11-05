// Package pilfer is the internals of the go-pilfer tool, which extracts
// a named type and all of its dependent named types from a given package
// and produces a new file, with a different package name, containing
// those types.
//
// Its intended purpose is to extract types that are used with
// Marshal/Unmarshal functions for encodings such as JSON, gob, etc so that
// an external program can produce or consume data files compatible with
// some other program without depending on that program.
//
// Why not just depend on the other program? Where possible that is
// recommended, but there are two situations where that isn't easy: first,
// it's not possible to import types from a "main" package. Secondly, the
// shape of a given type may evolve over multiple versions of the defining
// package -- all under the same import path -- but your program needs to
// support multiple versions at once. In that latter case, this tool
// effectively creates a "snapshot" of the needed types; it is a funny sort
// of "vendoring" that works on individual types rather than whole packages.
//
// While processing the given type it may be necessary to copy a type from
// another source package entirely. Since this new type comes from an entirely
// different namespace, its name may collide with others. Pilfer does not
// currently attempt to deal with this in any special way, so in some cases
// it may be necessary to do some manual renaming work after pilfer has
// finished in order to resolve conflicting type names.
package pilfer
