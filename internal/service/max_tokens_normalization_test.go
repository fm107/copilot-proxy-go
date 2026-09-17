package service

import (
	"strings"
	"testing"

	"github.com/tonghaoch/copilot-proxy-go/internal/state"
)

type fakeModels struct{ m map[string]*state.Model }

func (f fakeModels) FindModel(id string) *state.Model { return f.m[id] }

func newFakeFinder(id string, maxOut int) fakeModels {
	return fakeModels{m: map[string]*state.Model{
		id: {
			ID: id,
			Capabilities: state.ModelCapabilities{
				Limits: state.ModelLimits{MaxOutputTokens: maxOut, MaxContextWindowTokens: 1000000},
			},
		},
	}}
}

// TestRepro_ClientSendsOnlyMaxCompletionTokens reproduces the reported bug:
// a client (ZCode / modern OpenAI SDKs) sends only max_completion_tokens.
func TestRepro_ClientSendsOnlyMaxCompletionTokens(t *testing.T) {
	body := `{"model":"claude-opus-5","stream":true,"max_completion_tokens":32768,
	          "messages":[{"role":"user","content":"hi"}]}`

	out, _, _, err := ParseAndPatchChatCompletionWithModels(strings.NewReader(body), newFakeFinder("claude-opus-5", 64000))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	s := string(out)
	t.Logf("patched payload: %s", s)

	hasMax := strings.Contains(s, `"max_tokens"`)
	hasMaxCT := strings.Contains(s, `"max_completion_tokens"`)
	if hasMax && hasMaxCT {
		t.Errorf("BUG REPRODUCED: payload contains BOTH max_tokens and max_completion_tokens -> upstream 400")
	}
}

// TestRepro_ClientSendsBothFields reproduces the variant where the client
// itself already sends both fields (passthrough forwards both untouched).
func TestRepro_ClientSendsBothFields(t *testing.T) {
	body := `{"model":"claude-opus-5","max_tokens":32000,"max_completion_tokens":32768,
	          "messages":[{"role":"user","content":"hi"}]}`

	out, _, _, err := ParseAndPatchChatCompletionWithModels(strings.NewReader(body), newFakeFinder("claude-opus-5", 64000))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	s := string(out)
	t.Logf("patched payload: %s", s)

	hasMax := strings.Contains(s, `"max_tokens"`)
	hasMaxCT := strings.Contains(s, `"max_completion_tokens"`)
	if hasMax && hasMaxCT {
		t.Errorf("BUG REPRODUCED: both fields forwarded -> upstream 400")
	}
}

// TestRepro_ClientSendsMaxTokensOnly is the classic case that keeps working.
func TestRepro_ClientSendsMaxTokensOnly(t *testing.T) {
	body := `{"model":"claude-opus-5","max_tokens":32000,
	          "messages":[{"role":"user","content":"hi"}]}`

	out, _, _, err := ParseAndPatchChatCompletionWithModels(strings.NewReader(body), newFakeFinder("claude-opus-5", 64000))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	s := string(out)
	t.Logf("patched payload: %s", s)
	if !strings.Contains(s, `"max_tokens":32000`) {
		t.Errorf("expected max_tokens preserved")
	}
}
