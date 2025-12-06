package patchutil

import (
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
)

// Unpacker unpacks the patch file with the given index.
func Unpacker(index, patch, output string) error {
	if !fsutil.Exists(index) {
		return errutil.WithFramef("path does not exist: %s", index)
	}

	if !fsutil.Exists(patch) {
		return errutil.WithFramef("path does not exist: %s", patch)
	}

	f, err := fsutil.Read(index)
	if err != nil {
		return errutil.WithFrame(err)
	}

	idx, err := Parse(f)
	if err != nil {
		return errutil.WithFrame(err)
	}

	if err := idx.Unpack(patch, output); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}
