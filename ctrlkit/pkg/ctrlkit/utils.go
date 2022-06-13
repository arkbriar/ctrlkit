package ctrlkit

import "bytes"

// actions must be a slice with at least 1 elements.
func describeGroup(head string, actions ...ReconcileAction) string {
	buf := &bytes.Buffer{}

	buf.WriteString(head)
	buf.WriteString("(")
	for _, act := range actions[:len(actions)-1] {
		buf.WriteString(act.Description())
		buf.WriteString(", ")
	}
	buf.WriteString(actions[len(actions)-1].Description())
	buf.WriteString(")")

	return buf.String()
}
