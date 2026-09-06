package incident

import "testing"

func TestNewPaymentIncidentRequiresCaseOwnership(t *testing.T) {
	if _, err := NewPaymentIncident("", 1, 1, "payment_timeout"); err == nil {
		t.Fatal("missing id should be rejected")
	}
	if _, err := NewPaymentIncident("pi-1", 0, 1, "payment_timeout"); err == nil {
		t.Fatal("missing tenant should be rejected")
	}
	if _, err := NewPaymentIncident("pi-1", 1, 0, "payment_timeout"); err == nil {
		t.Fatal("missing creator should be rejected")
	}

	caseRecord, err := NewPaymentIncident("pi-1", 1, 2, "payment_timeout")
	if err != nil {
		t.Fatal(err)
	}
	if caseRecord.Status != StatusPendingVerification || caseRecord.Priority != "P2" {
		t.Fatalf("unexpected defaults: %+v", caseRecord)
	}
}

func TestPaymentIncidentStateMachine(t *testing.T) {
	caseRecord, err := NewPaymentIncident("pi-1", 1, 2, "callback_failure")
	if err != nil {
		t.Fatal(err)
	}
	if err := caseRecord.TransitionTo(StatusAwaitingConfirmation); err == nil {
		t.Fatal("pending verification must not skip investigating")
	}
	for _, target := range []Status{StatusInvestigating, StatusAwaitingConfirmation, StatusResolved} {
		if err := caseRecord.TransitionTo(target); err != nil {
			t.Fatalf("transition to %s failed: %v", target, err)
		}
	}
	if !caseRecord.IsTerminal() {
		t.Fatal("resolved case should be terminal")
	}
	if err := caseRecord.TransitionTo(StatusInvestigating); err == nil {
		t.Fatal("terminal case must not reopen silently")
	}
}
