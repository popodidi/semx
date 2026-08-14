package version

import "testing"

func TestGetReturnsDefaultBuildMetadata(t *testing.T) {
	got := Get()
	want := Info{
		Version:   "dev",
		Commit:    "unknown",
		BuildTime: "unknown",
	}
	if got != want {
		t.Fatalf("Get() = %#v, want %#v", got, want)
	}
}
