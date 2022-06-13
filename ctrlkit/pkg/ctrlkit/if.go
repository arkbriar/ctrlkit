package ctrlkit

func If(predicate bool, act ReconcileAction) ReconcileAction {
	if predicate {
		return act
	} else {
		return Nop
	}
}
