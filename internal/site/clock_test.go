package site

import (
	"testing"
	"time"
)

func TestTimeFromUnixMilliseconds(t *testing.T) {
	got := timeFromUnixMilliseconds(1785136979000)
	want := time.Date(2026, time.July, 27, 7, 22, 59, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("timeFromUnixMilliseconds() = %s, want %s", got, want)
	}
}
