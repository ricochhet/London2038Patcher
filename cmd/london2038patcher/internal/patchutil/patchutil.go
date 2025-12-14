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

type Dat struct {
	*LocaleMap

	Locales []string
}

// Unpack unpacks the patch file with the given index.
func (d *Dat) Unpack(index, patch, output string) error {
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

	idx, err := Decode(f)
	if err != nil {
		return errutil.WithFrame(err)
	}

	locales, err := d.toInt(d.Locales)
	if err != nil {
		return errutil.WithFrame(err)
	}

	if err := idx.Unpack(patch, output, locales); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// Pack packs the path with the given index.
func (d *Dat) Pack(index, path, output string) error {
	if !fsutil.Exists(index) {
		return errutil.WithFramef("path does not exist: %s", index)
	}

	f, err := fsutil.Read(index)
	if err != nil {
		return errutil.WithFrame(err)
	}

	idx, err := Decode(f)
	if err != nil {
		return errutil.WithFrame(err)
	}

	locales, err := d.toInt(d.Locales)
	if err != nil {
		return errutil.WithFrame(err)
	}

	if err := idx.Pack(path, output, locales); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// Pack packs the path with the given index.
func (d *Dat) PackWithIndex(path, index, patch string) error {
	locales, err := d.toInt(d.Locales)
	if err != nil {
		return errutil.WithFrame(err)
	}

	if err := d.LocaleMap.PackWithIndex(path, index, patch, locales); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// UnpackFromFile unpacks the patches from the specified file to the given output.
func (d *Dat) UnpackFromFile(path, output string) error {
	if !fsutil.Exists(path) {
		return errutil.WithFramef("path does not exist: %s", path)
	}

	p, err := jsonutil.ReadAndUnmarshal[Patches](path)
	if err != nil {
		return err
	}

	return d.unpackFromFile(output, p)
}

// unpackFromFile unpacks the patch files specified to the given output.
func (d *Dat) unpackFromFile(output string, patches *Patches) error {
	for _, patch := range patches.Patches {
		if err := d.Unpack(patch.Idx, patch.Dat, output); err != nil {
			return err
		}
	}

	return nil
}
