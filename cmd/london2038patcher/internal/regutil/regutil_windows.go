package regutil

import (
	"fmt"
	"os"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"golang.org/x/sys/windows/registry"
)

const (
	hglKeyPath = `SOFTWARE\WOW6432Node\Flagship Studios\Hellgate London`
	hglCUKey   = "HellgateCUKey"
	hglKey     = "HellgateKey"
)

type RegistryKey struct {
	*registry.Key
}

// Regedit edits the registry and adds keys created by Hellgate: London setup.
func Regedit(cuKey, key string) error {
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, hglKeyPath, registry.ALL_ACCESS)
	if err != nil {
		return errutil.WithFrame(err)
	}
	defer k.Close()

	rk := RegistryKey{Key: &k}

	if err := rk.cuKey(cuKey); err != nil {
		return errutil.WithFrame(err)
	}

	if err := rk.key(key); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// cuKey sets "HellgateCUKey" in registry to the specified path.
func (k *RegistryKey) cuKey(path string) error {
	key, err := fsutil.Normalize(path)
	if err != nil {
		return errutil.WithFrame(err)
	}

	if !fsutil.Exists(key) || key == "" {
		return errutil.WithFramef("path does not exist: %s", key)
	}

	fmt.Fprintf(os.Stdout, "Setting \"%s\\%s\" to \"%s\"\n", hglKeyPath, hglKey, path)

	return k.SetStringValue(hglCUKey, key)
}

// key sets "HellgateKey" in registry to the specified path.
func (k *RegistryKey) key(path string) error {
	key, err := fsutil.Normalize(path)
	if err != nil {
		return errutil.WithFrame(err)
	}

	if !fsutil.Exists(key) || key == "" {
		return errutil.WithFramef("path does not exist: %s", key)
	}

	fmt.Fprintf(os.Stdout, "Setting '%s\\%s' to '%s'\n", hglKeyPath, hglKey, path)

	return k.SetStringValue(hglKey, key)
}
