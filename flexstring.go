package infakt

import (
	"bytes"
	"encoding/json"
)

// FlexString is a string that decodes from a JSON string, number, boolean, or
// null. Several inFakt fields are typed inconsistently by the API — most
// notably client days_to_payment, which is returned as an empty string ("")
// when unset but as a bare number (e.g. 14) once a value is assigned.
// FlexString normalizes every shape to its string form so responses round-trip
// cleanly. With no custom MarshalJSON it marshals back as a JSON string, which
// the API accepts on input, and an empty value is omitted under `omitempty`.
//
// Because the underlying type is string, callers can assign string literals
// (e.g. FlexString-typed fields accept `"14"`) and compare against string
// constants directly; convert with string(v) when an explicit string is
// required.
type FlexString string

// UnmarshalJSON implements [json.Unmarshaler]. It accepts a JSON string
// (used as-is), a JSON number or boolean (kept as its literal text), or null
// (decoded to the empty string).
func (f *FlexString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*f = ""
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*f = FlexString(s)
		return nil
	}
	// Number, boolean, or other scalar: preserve the raw literal text.
	*f = FlexString(data)
	return nil
}
