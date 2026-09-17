package handler

// normalizeTokenLimitFields collapses the mutually exclusive
// max_tokens / max_completion_tokens pair on a decoded Anthropic Messages
// payload. The native Messages handler forwards the client's JSON nearly
// untouched, so without this an extra max_completion_tokens sent by modern
// SDKs rides along with max_tokens and Copilot rejects the request with
// 400 "max_tokens and max_completion_tokens cannot both be set".
// max_tokens wins when both are present; max_completion_tokens is carried
// over only when max_tokens is missing or null.
func normalizeTokenLimitFields(payload map[string]any) {
	mct, ok := payload["max_completion_tokens"]
	if !ok {
		return
	}
	if existing, hasMax := payload["max_tokens"]; !hasMax || existing == nil {
		if v, isNum := mct.(float64); isNum && v > 0 {
			payload["max_tokens"] = int(v)
		}
	}
	delete(payload, "max_completion_tokens")
}

// normalizeResponsesTokenLimit maps max_completion_tokens onto the Responses
// API's max_output_tokens (dropping the duplicate when max_output_tokens is
// already set) and removes the alien field before the payload is forwarded
// to Copilot's Responses endpoint.
func normalizeResponsesTokenLimit(payload map[string]any) {
	mct, ok := payload["max_completion_tokens"]
	if !ok {
		return
	}
	if existing, hasMax := payload["max_output_tokens"]; !hasMax || existing == nil {
		if v, isNum := mct.(float64); isNum && v > 0 {
			payload["max_output_tokens"] = int(v)
		}
	}
	delete(payload, "max_completion_tokens")
}
