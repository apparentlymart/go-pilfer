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
//
// This program only brings the type definitions themselves, and not any
// methods associated with them. In particular, this means that types that
// implement interfaces like json.Marshaler, gob.Decoder, etc will not have
// these custom behaviors preserved, which will probably cause marshalling or
// unmarshalling to fail. The user must manually copy or re-implement such
// methods.
//
// This program will import interface types along with all other named types,
// but note that this may not actually prove useful because any named types
// used by those interfaces will also be imported, creating a new interface
// type that is incompatible with the original. This will cause problems for
// gob encoding and decoding because any stored interface types can never
// match.
//
// At this time the program does not import constants of a type that are found
// within the same package, which is problematic for named types used as
// enumeration types since the constants in the source package will have
// an incompatible type. It is necessary, therefore, to manually copy the
// relevant constants.
package pilfer
