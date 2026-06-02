package disk_test

import (
	"testing"

	"github.com/gammazero/fsutil/disk"
)

func TestUsage(t *testing.T) {
	tempDir := t.TempDir()
	us, err := disk.Usage(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Path:", us.Path)
	if us.Path != tempDir {
		t.Fatal("incorrect path:", us.Path)
	}

	t.Log("Total:", us.Total)
	if us.Total == 0 {
		t.Fatal("Total should not be 0")
	}

	t.Log("Free:", us.Free)
	if us.Free == 0 {
		t.Fatal("Free should not be 0")
	}

	t.Log("Used:", us.Used)
	if us.Used == 0 {
		t.Fatal("Used should not be 0")
	}

	t.Log("Percent:", us.Percent)
	if us.Percent <= 0.0 {
		t.Fatal("Percent must be greater than 0.0")
	}
}
