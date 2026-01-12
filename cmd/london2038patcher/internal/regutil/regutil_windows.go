package regutil

import (
	"path/filepath"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
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
		return errutil.New("registry.CreateKey", err)
	}
	defer k.Close()

	rk := RegistryKey{Key: &k}

	if err := rk.cuKey(cuKey); err != nil {
		return errutil.New("rk.cuKey", err)
	}

	if err := rk.key(key); err != nil {
		return errutil.New("rk.key", err)
	}

	return nil
}

// cuKey sets "HellgateCUKey" in registry to the specified path.
func (k *RegistryKey) cuKey(path string) error {
	key, err := filepath.Abs(path)
	if err != nil {
		return errutil.New("filepath.Abs", err)
	}

	if !fsutil.Exists(key) || key == "" {
		return errutil.WithFramef("path does not exist: %s", key)
	}

	logutil.Infof(logutil.Get(), "Setting \"%s\\%s\" to \"%s\"\n", hglKeyPath, hglKey, path)

	return k.SetStringValue(hglCUKey, key)
}

// key sets "HellgateKey" in registry to the specified path.
func (k *RegistryKey) key(path string) error {
	key, err := filepath.Abs(path)
	if err != nil {
		return errutil.New("filepath.Abs", err)
	}

	if !fsutil.Exists(key) || key == "" {
		return errutil.WithFramef("path does not exist: %s", key)
	}

	logutil.Infof(logutil.Get(), "Setting '%s\\%s' to '%s'\n", hglKeyPath, hglKey, path)

	return k.SetStringValue(hglKey, key)
}
