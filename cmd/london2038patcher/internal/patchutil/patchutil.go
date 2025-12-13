package patchutil

import (
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/jsonutil"
)

type Patches struct {
	Patches []Patch `json:"patches"`
}

type Patch struct {
	Idx string `json:"idx"`
	Dat string `json:"dat"`
}

// Unpack unpacks the patch file with the given index.
func Unpack(index, patch, output string) error {
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

// UnpackFromFile unpacks the patches from the specified file to the given output.
func UnpackFromFile(path, output string) error {
	if !fsutil.Exists(path) {
		return errutil.WithFramef("path does not exist: %s", path)
	}

	p, err := jsonutil.Unmarshal[Patches](path)
	if err != nil {
		return err
	}

	return p.unpackFromFile(output)
}

// unpackFromFile unpacks the patch files specified to the given output.
func (p *Patches) unpackFromFile(output string) error {
	for _, patch := range p.Patches {
		if err := Unpack(patch.Idx, patch.Dat, output); err != nil {
			return err
		}
	}

	return nil
}
