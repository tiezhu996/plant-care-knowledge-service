package util

import "testing"

func TestReminderStatusTextSkipped(t *testing.T) {
	if got := ReminderStatusText("skipped"); got != "已跳过" {
		t.Fatalf("ReminderStatusText(skipped)=%q want 已跳过", got)
	}
}
