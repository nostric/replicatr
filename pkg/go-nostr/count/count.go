package count

import (
	"encoding/json"
	"fmt"

	"github.com/Hubmakerlabs/replicatr/pkg/go-nostr/envelopes"
	"github.com/Hubmakerlabs/replicatr/pkg/go-nostr/filter"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jwriter"
	"github.com/tidwall/gjson"
)

var _ envelopes.Envelope = (*CountEnvelope)(nil)

type CountEnvelope struct {
	SubscriptionID string
	filter.Filters
	Count *int64
}

func (_ CountEnvelope) Label() string { return "COUNT" }
func (c CountEnvelope) String() string {
	v, _ := json.Marshal(c)
	return string(v)
}

func (v *CountEnvelope) UnmarshalJSON(data []byte) error {
	r := gjson.ParseBytes(data)
	arr := r.Array()
	if len(arr) < 3 {
		return fmt.Errorf("failed to decode COUNT envelope: missing filters")
	}
	v.SubscriptionID = arr[1].Str

	if len(arr) < 3 {
		return fmt.Errorf("COUNT array must have at least 3 items")
	}

	var countResult struct {
		Count *int64 `json:"count"`
	}
	if err := json.Unmarshal([]byte(arr[2].Raw), &countResult); err == nil && countResult.Count != nil {
		v.Count = countResult.Count
		return nil
	}

	v.Filters = make(filter.Filters, len(arr)-2)
	f := 0
	for i := 2; i < len(arr); i++ {
		item := []byte(arr[i].Raw)

		if err := easyjson.Unmarshal(item, &v.Filters[f]); err != nil {
			return fmt.Errorf("%w -- on filter %d", err, f)
		}

		f++
	}

	return nil
}

func (v CountEnvelope) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	w.RawString(`["COUNT",`)
	w.RawString(`"` + v.SubscriptionID + `"`)
	if v.Count != nil {
		w.RawString(fmt.Sprintf(`{"count":%d}`, *v.Count))
	} else {
		for _, f := range v.Filters {
			w.RawString(`,`)
			f.MarshalEasyJSON(&w)
		}
	}
	w.RawString(`]`)
	return w.BuildBytes()
}
