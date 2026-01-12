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
func (o *Options) Unpack(index, patch, output string) error {
	if !fsutil.Exists(index) {
		return errutil.Newf("fsutil.Exists(index)", "path does not exist: %s", index)
	}

	if !fsutil.Exists(patch) {
		return errutil.Newf("fsutil.Exists(patch)", "path does not exist: %s", patch)
	}

	f, err := fsutil.Read(index)
	if err != nil {
		return errutil.New("fsutil.Read", err)
	}

	idx, err := Decode(f)
	if err != nil {
		return errutil.New("Decode", err)
	}

	if err := idx.Unpack(patch, output, o.Filter, o.Archs, o.IdxOptions); err != nil {
		return errutil.New("idx.Unpack", err)
	}

	return nil
}

// Pack packs the path with the given index.
func (o *Options) Pack(index, path, output string) error {
	if !fsutil.Exists(index) {
		return errutil.WithFramef("path does not exist: %s", index)
	}

	f, err := fsutil.Read(index)
	if err != nil {
		return errutil.New("fsutil.Read", err)
	}

	idx, err := Decode(f)
	if err != nil {
		return errutil.New("Decode", err)
	}

	if err := idx.Pack(path, output, o.Filter, o.Archs, o.IdxOptions); err != nil {
		return errutil.New("idx.Pack", err)
	}

	return nil
}

// Pack packs the path with the given index.
func (o *Options) PackWithIndex(path, index, patch string) error {
	if err := o.Registry.PackWithIndex(
		path,
		index,
		patch,
		o.Filter,
		o.Archs,
		o.IdxOptions,
	); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// UnpackFromFile unpacks the patches from the specified file to the given output.
func (o *Options) UnpackFromFile(path, output string) error {
	if !fsutil.Exists(path) {
		return errutil.WithFramef("path does not exist: %s", path)
	}

	p, err := jsonutil.ReadAndUnmarshal[Patches](path)
	if err != nil {
		return errutil.New("jsonutil.ReadAndUnmarshal", err)
	}

	return o.unpackFromFile(output, p)
}

// unpackFromFile unpacks the patch files specified to the given output.
func (o *Options) unpackFromFile(output string, patches *Patches) error {
	for _, patch := range patches.Patches {
		if err := o.Unpack(patch.Idx, patch.Dat, output); err != nil {
			return errutil.WithFrame(err)
		}
	}

	return nil
}
