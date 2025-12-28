package internal

type Event struct {
	Provider string                 `json:"provider"`
	Name     string                 `json:"name"`
	Data     map[string]interface{} `json:"data"`
}
