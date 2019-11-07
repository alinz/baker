package rule

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Director func(r *http.Request)

type RequestUpdater interface {
	Director(director Director) Director
}

type RequestUpdaters []RequestUpdater

var _ json.Unmarshaler = (*RequestUpdaters)(nil)

func (r *RequestUpdaters) UnmarshalJSON(p []byte) error {
	var rawMessages []json.RawMessage

	err := json.Unmarshal(p, &rawMessages)
	if err != nil {
		return err
	}

	check := struct {
		Name string `json:"name"`
	}{}

	for _, rawMessage := range rawMessages {
		json.Unmarshal(rawMessage, &check)

		switch check.Name {
		case "replace_path":
			replacePath := &ReplacePath{}
			json.Unmarshal(rawMessage, replacePath)
			*r = append(*r, replacePath)
		default:
			return fmt.Errorf("failed to process %s RequestUpdater", check.Name)
		}
	}

	return nil
}
