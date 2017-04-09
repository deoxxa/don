// Package qstring provides an an easy way to marshal and unmarshal query string
// data to and from structs representations.
//
// This library was designed with consistency in mind, to this end the provided
// Marshal and Unmarshal interfaces should seem familiar to those who have used
// packages such as "encoding/json".
//
// Additionally this library includes support for converting query parameters
// to and from time.Time types as well as nested structs and pointer values.
package qstring
