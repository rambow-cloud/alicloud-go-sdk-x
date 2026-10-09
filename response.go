package alicloud

import "encoding/json/v2"

type serviceEnvelope struct {
	Code, Message, RequestID string
}

// decodeServiceEnvelope accepts only the reviewed native member spellings.
// Presence, including an empty uppercase value, takes precedence over lowercase.
func decodeServiceEnvelope(data []byte, success bool) (serviceEnvelope, error) {
	if success {
		// Business output fields named code/message need not be error strings.
		var metadata struct {
			RequestID  *string `json:"RequestId"`
			LowerReqID *string `json:"requestId"`
		}
		if err := json.Unmarshal(data, &metadata); err != nil {
			return serviceEnvelope{}, err
		}
		if metadata.RequestID != nil {
			return serviceEnvelope{RequestID: *metadata.RequestID}, nil
		}
		if metadata.LowerReqID != nil {
			return serviceEnvelope{RequestID: *metadata.LowerReqID}, nil
		}
		return serviceEnvelope{}, nil
	}
	var wire struct {
		Code       *string `json:"Code"`
		LowerCode  *string `json:"code"`
		Message    *string `json:"Message"`
		LowerMsg   *string `json:"message"`
		RequestID  *string `json:"RequestId"`
		LowerReqID *string `json:"requestId"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return serviceEnvelope{}, err
	}
	selectValue := func(primary, fallback *string) string {
		if primary != nil {
			return *primary
		}
		if fallback != nil {
			return *fallback
		}
		return ""
	}
	return serviceEnvelope{selectValue(wire.Code, wire.LowerCode), selectValue(wire.Message, wire.LowerMsg), selectValue(wire.RequestID, wire.LowerReqID)}, nil
}
