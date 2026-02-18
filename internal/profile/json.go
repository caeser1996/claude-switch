package profile

import "encoding/json"

// jsonUnmarshal is a package-level wrapper to avoid importing encoding/json
// in multiple places and to make testing easier.
func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
