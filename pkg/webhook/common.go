package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"

	"githooks/internal"
)

// rawObjectAndFlatten unmarshals a raw JSON byte slice into both an interface{}
// and a flattened map[string]interface{}. This is useful for both preserving the
// original structure and for easy access to nested fields.
func rawObjectAndFlatten(raw []byte) (interface{}, map[string]interface{}) {
	var out interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, map[string]interface{}{}
	}
	objectMap, ok := out.(map[string]interface{})
	if !ok {
		return out, map[string]interface{}{}
	}
	return out, internal.Flatten(objectMap)
}

func requestID(r *http.Request) string {
	if r == nil {
		return watermill.NewUUID()
	}
	if id := r.Header.Get("X-Request-Id"); id != "" {
		return id
	}
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	if id := r.Header.Get("X-Correlation-Id"); id != "" {
		return id
	}
	return watermill.NewUUID()
}
