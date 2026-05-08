package infakt

import "time"

// The helpers in this file create pointer values for primitive types. They
// are intended for use with *Request types whose fields are pointers in
// order to distinguish "unset" (nil) from the type's zero value. For
// example, in [InvoiceRequest] or [ClientEntityRequest] the field
// `Notes *string` is nil when the caller does not wish to send the field
// at all, but `infakt.String("")` sends an explicit empty string and
// `infakt.Int(0)` sends an explicit zero. Without these helpers callers
// would need to introduce intermediate variables solely to take their
// addresses.

// String returns a pointer to the given string value.
func String(v string) *string { return &v }

// Int returns a pointer to the given int value.
func Int(v int) *int { return &v }

// Int64 returns a pointer to the given int64 value.
func Int64(v int64) *int64 { return &v }

// Bool returns a pointer to the given bool value.
func Bool(v bool) *bool { return &v }

// Float64 returns a pointer to the given float64 value.
func Float64(v float64) *float64 { return &v }

// Time returns a pointer to the given time.Time value.
func Time(v time.Time) *time.Time { return &v }
