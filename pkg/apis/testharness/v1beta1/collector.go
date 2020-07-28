package v1beta1

import (
	"fmt"
	"strings"
)

const (
	pod     = "pod"
	events  = "events"
	command = "command"
)

// Validate checks user input and updates type if not provided
// It is expected to be called prior to any other call
func (tc *TestCollector) Validate() {
	cleanType(tc)
	switch tc.Type {
	case command:
		validateCmd(tc)
	case pod:
		validPod(tc)
	case events:
		validEvents(tc)
	default:
		tc.InvalidReason = fmt.Sprintf("collector type %q unknown", tc.Type)
	}
}

func validEvents(tc *TestCollector) {
	if tc.Cmd != "" || tc.Selector != "" || tc.Container != "" {
		tc.InvalidReason = "event collector can not have a selector, container or command"
		return
	}
	tc.Valid = true
}

func validPod(tc *TestCollector) {
	if tc.Cmd != "" {
		tc.InvalidReason = "pod collector can NOT have a command"
		return
	}
	if tc.Pod == "" && tc.Selector == "" {
		tc.InvalidReason = "pod collector requires a pod or selector"
		return
	}
	tc.Valid = true
}

func validateCmd(tc *TestCollector) {
	if tc.Cmd == "" {
		tc.InvalidReason = "command collector requires a command"
		return
	}
	if tc.Pod != "" || tc.Namespace != "" || tc.Container != "" || tc.Selector != "" {
		tc.InvalidReason = "command collectors can NOT have pod, namespace, container or selectors"
		return
	}
	tc.Valid = true
}

// determines and cleans collector type
func cleanType(tc *TestCollector) {
	// intuit pod or command or invalid
	if tc.Type == "" {
		// assume command if cmd provided
		if tc.Cmd != "" {
			tc.Type = command
		} else {
			tc.Type = pod
		}
	}
	tc.Type = strings.ToLower(tc.Type)
}

// Command provides the command to exec to perform the collection
func (tc *TestCollector) Command() *Command {
	switch tc.Type {
	case pod:
		return podCommand(tc)
	case command:
		return &Command{
			Command:       tc.Cmd,
			IgnoreFailure: true,
		}
	case events:
		return eventCommand(tc)
	}
	return nil
}

func eventCommand(tc *TestCollector) *Command {
	var b strings.Builder
	b.WriteString("kubectl get events")
	if len(tc.Pod) > 0 {
		fmt.Fprintf(&b, " %s", tc.Pod)
	}
	ns := tc.Namespace
	if len(tc.Namespace) == 0 {
		ns = "$NAMESPACE"
	}
	fmt.Fprintf(&b, " -n %s", ns)
	return &Command{
		Command:       b.String(),
		IgnoreFailure: true,
	}
}

func podCommand(tc *TestCollector) *Command {
	var b strings.Builder
	b.WriteString("kubectl logs --prefix")
	if len(tc.Pod) > 0 {
		fmt.Fprintf(&b, " %s", tc.Pod)
	}
	if len(tc.Selector) > 0 {
		fmt.Fprintf(&b, " -l %s", tc.Selector)
	}
	ns := tc.Namespace
	if len(tc.Namespace) == 0 {
		ns = "$NAMESPACE"
	}
	fmt.Fprintf(&b, " -n %s", ns)
	if len(tc.Container) > 0 {
		fmt.Fprintf(&b, " -c %s", tc.Container)
	} else {
		b.WriteString(" --all-containers")
	}
	return &Command{
		Command:       b.String(),
		IgnoreFailure: true,
	}
}

// String provides defaults of the type of collector
func (tc *TestCollector) String() string {
	if !tc.Valid {
		return fmt.Sprintf("[collector invalid: %s]", tc.InvalidReason)
	}
	if !(tc.Type == pod || tc.Type == events || tc.Type == command) {
		return fmt.Sprintf("unexpected collector type: %q", tc.Type)
	}
	var b strings.Builder
	b.WriteString("[")
	details := []string{}
	details = append(details, fmt.Sprintf("type==%s", tc.Type))
	if len(tc.Pod) > 0 {
		details = append(details, fmt.Sprintf("pod==%s", tc.Pod))
	}
	if len(tc.Selector) > 0 {
		details = append(details, fmt.Sprintf("label: %s", tc.Selector))
	}
	if len(tc.Namespace) > 0 {
		details = append(details, fmt.Sprintf("namespace: %s", tc.Namespace))
	}
	if len(tc.Container) > 0 {
		details = append(details, fmt.Sprintf("container: %s", tc.Container))
	}
	if len(tc.Cmd) > 0 {
		details = append(details, fmt.Sprintf("command: %s", tc.Cmd))
	}
	b.WriteString(strings.Join(details, ","))
	b.WriteString("]")
	return b.String()
}
