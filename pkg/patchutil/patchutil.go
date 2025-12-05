package patchutil

import (
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
)

// Unpacker unpacks the patch file with the given index.
func Unpacker(index, patch, output string) error {
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
