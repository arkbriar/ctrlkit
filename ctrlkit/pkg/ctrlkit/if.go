package ctrlkit

// If returns the given action if predicate is true, or an Nop otherwise.
func If(predicate bool, act ReconcileAction) ReconcileAction {
	if predicate {
		return act
	} else {
		return Nop
	}
}
