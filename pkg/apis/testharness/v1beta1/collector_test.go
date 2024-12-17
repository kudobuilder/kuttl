package v1beta1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestCollector_String(t *testing.T) {
	type fields struct {
		Type      string
		Pod       string
		Namespace string
		Container string
		Selector  string
		Cmd       string
	}
	tests := []struct {
		name     string
		fields   fields
		contains string
	}{
		{
			name:     "default pod",
			fields:   fields{Pod: "foo"},
			contains: "type==pod",
		},
		{
			name:     "default command",
			fields:   fields{Cmd: "foo"},
			contains: "type==command",
		},
		{
			name:     "bad type",
			fields:   fields{Type: "foo"},
			contains: "collector invalid:",
		},
		{
			name:     "valid pod",
			fields:   fields{Type: "pod", Pod: "foo"},
			contains: "pod==foo",
		},
		{
			name:     "invalid pod no pod or selector",
			fields:   fields{Type: "pod"},
			contains: "collector invalid:",
		},
		{
			name:     "invalid pod with command",
			fields:   fields{Type: "pod", Cmd: "foo"},
			contains: "collector invalid:",
		},
		{
			name:     "invalid pod no name or selector",
			fields:   fields{Type: "pod"},
			contains: "collector invalid:",
		},
		{
			name:     "valid events",
			fields:   fields{Type: "events"},
			contains: "type==events",
		},
		{
			name:     "invalid events with container",
			fields:   fields{Type: "events", Container: "foo"},
			contains: "collector invalid:",
		},
		{
			name:     "invalid events with selector",
			fields:   fields{Type: "events", Selector: "foo=bar"},
			contains: "collector invalid:",
		},
		{
			name:     "invalid events with command",
			fields:   fields{Type: "events", Cmd: "foo"},
			contains: "collector invalid:",
		},
		{
			name:     "valid command",
			fields:   fields{Type: "command", Cmd: "foo"},
			contains: "command: foo",
		},
		{
			name:     "invalid command without command",
			fields:   fields{Type: "command"},
			contains: "collector invalid:",
		},
		{
			name:     "invalid command with ns",
			fields:   fields{Type: "command", Namespace: "foo"},
			contains: "collector invalid:",
		},
		{
			name:     "invalid command with container",
			fields:   fields{Type: "command", Container: "foo"},
			contains: "collector invalid:",
		},
		{
			name:     "invalid command with pod",
			fields:   fields{Type: "command", Pod: "foo"},
			contains: "collector invalid:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &TestCollector{
				Type:      tt.fields.Type,
				Pod:       tt.fields.Pod,
				Namespace: tt.fields.Namespace,
				Container: tt.fields.Container,
				Selector:  tt.fields.Selector,
				Cmd:       tt.fields.Cmd,
			}
			got := tc.String()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("String() = %v, does not contain %v", got, tt.contains)
			}
		})
	}
}

func TestPodCommand(t *testing.T) {
	tests := []struct {
		name string
		tc   TestCollector
		cmd  string
	}{
		{
			name: "selector with default tail",
			tc:   TestCollector{Type: pod, Selector: "x=y"},
			cmd:  "kubectl logs --prefix -l x=y -n $NAMESPACE --all-containers --tail=10",
		},
		{
			name: "pod name with default tail",
			tc:   TestCollector{Type: pod, Pod: "foo"},
			cmd:  "kubectl logs --prefix foo -n $NAMESPACE --all-containers --tail=-1",
		},
		{
			name: "selector with set tail",
			tc:   TestCollector{Type: pod, Selector: "x=y", Tail: 42},
			cmd:  "kubectl logs --prefix -l x=y -n $NAMESPACE --all-containers --tail=42",
		},
		{
			name: "pod name with set tail",
			tc:   TestCollector{Type: pod, Pod: "foo", Tail: 42},
			cmd:  "kubectl logs --prefix foo -n $NAMESPACE --all-containers --tail=42",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := podCommand(&tt.tc)
			assert.Equal(t, cmd.Command, tt.cmd)
		})
	}
}
