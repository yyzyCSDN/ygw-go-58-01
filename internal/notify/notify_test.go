package notify

import (
	"context"
	"testing"

	"reconcilesvc/internal/model"
)

func TestCountingSubscriber(t *testing.T) {
	counting := NewCountingSubscriber()
	notifier := NewNotifier(counting)
	msg := Message{WindowID: "w0", Missing: 1, Mismatch: 2, Extra: 3}
	if err := notifier.Notify(context.Background(), model.Window{ID: "w0"}, &model.Result{Missing: 1, Mismatch: 2, Extra: 3}, "k"); err != nil {
		t.Fatal(err)
	}
	summary := counting.Snapshot()
	if summary.Notifications != 1 || summary.Missing != 1 || summary.Mismatch != 2 || summary.Extra != 3 {
		t.Fatalf("summary = %+v", summary)
	}
	_ = msg
}

func TestChannelSubscriber(t *testing.T) {
	sub := NewChannelSubscriber(1)
	notifier := NewNotifier(sub)
	if err := notifier.Notify(context.Background(), model.Window{ID: "w"}, &model.Result{}, "k"); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-sub.Chan():
		if got.WindowID != "w" {
			t.Fatalf("message = %+v", got)
		}
	default:
		t.Fatalf("no message received")
	}
}
