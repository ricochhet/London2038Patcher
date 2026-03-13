package hostsutil

import (
	"path"
	"strings"

	"github.com/goodhosts/hostsfile"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
	"github.com/ricochhet/london2038patcher/pkg/timeutil"
)

// NewHosts returns a new hostsfile.Hosts.
func NewHosts() (*hostsfile.Hosts, error) {
	hosts, err := hostsfile.NewHosts()
	if err != nil {
		return nil, errutil.New("hostsfile.NewHosts", err)
	}

	if err := backupHosts(hosts); err != nil {
		return nil, errutil.New("backupHosts", err)
	}

	return hosts, nil
}

// Add adds an entry to the hosts file.
func Add(hf *hostsfile.Hosts, hosts map[string]string) error {
	for key, value := range hosts {
		if err := add(hf, key, value); err != nil {
			return err
		}
	}

	return hf.Flush()
}

// Remove removes an entry from the hosts file.
func Remove(hf *hostsfile.Hosts, hosts map[string]string) error {
	for key, value := range hosts {
		if err := remove(hf, key, value); err != nil {
			return err
		}
	}

	return hf.Flush()
}

// add adds an entry to the hosts file.
func add(hf *hostsfile.Hosts, ip string, hosts ...string) error {
	logutil.Infof(logutil.Get(), "Adding hostsfile entry: %s %s\n", ip, strings.Join(hosts, " "))
	return hf.Add(ip, hosts...)
}

// remove removes an entry from the hosts file.
func remove(hf *hostsfile.Hosts, ip string, hosts ...string) error {
	logutil.Infof(logutil.Get(), "Removing hostsfile entry: %s %s\n", ip, strings.Join(hosts, " "))
	return hf.Remove(ip, hosts...)
}

// backupHosts writes the hostsfile to the specified directory with the current timestamp.
func backupHosts(hf *hostsfile.Hosts) error {
	return fsutil.Write(path.Join("hosts", "hosts_"+timeutil.TimeStamp()), []byte(hf.String()))
}
