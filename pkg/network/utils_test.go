package network

import (
	"os"
	"runtime"
	"testing"
)

func TestUnlinkUdsFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	tmp := "tmpFile"
	var err error

	err = UnlinkUdsFile("unix", tmp)

	if err == nil {
		t.Errorf("should have error when unlinking a nonexistent file")
	}

	os.Create(tmp)

	err = UnlinkUdsFile("unix", tmp)
	if err != nil {
		t.Errorf("unlink file failed: %s", err.Error())
	}

	isExist, _ := pathExists(tmp)

	if isExist {
		t.Errorf("unlink file failed, file still exist")
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
