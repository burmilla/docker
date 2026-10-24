package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/registry"
	"github.com/imdario/mergo"
	"github.com/spf13/pflag"
)

const (
	// DefaultMaxConcurrentDownloads is the default value for
	// maximum number of downloads that
	// may take place at a time for each pull.
	DefaultMaxConcurrentDownloads = 3
	// DefaultShmSize is the default value for container's shm size
	DefaultShmSize = int64(67108864)
	// DefaultNetworkMtu is the default value for network MTU
	DefaultNetworkMtu = 1500
	// DisableNetworkBridge is the default value of the option to disable network bridge
	DisableNetworkBridge = "none"
)

// flatOptions contains configuration keys
// that MUST NOT be parsed as deep structures.
// Use this to differentiate these options
var flatOptions = map[string]bool{
	"log-opts":        true,
	"default-ulimits": true,
}

// ignoredOptions are config file keys from earlier releases that no longer
// do anything. They are dropped with a warning instead of failing startup.
var ignoredOptions = map[string]bool{
	"live-restore":        true,
	"userland-proxy-path": true,
}

// LogConfig represents the default log configuration.
// It includes json tags to deserialize configuration from a file
// using the same names that the flags in the command line use.
type LogConfig struct {
	Type   string            `json:"log-driver,omitempty"`
	Config map[string]string `json:"log-opts,omitempty"`
}

// commonBridgeConfig stores all the platform-common bridge driver specific
// configuration.
type commonBridgeConfig struct {
	Iface     string `json:"bridge,omitempty"`
	FixedCIDR string `json:"fixed-cidr,omitempty"`
}

// CommonConfig defines the configuration of a docker daemon which is
// common across platforms.
// It includes json tags to deserialize configuration from a file
// using the same names that the flags in the command line use.
type CommonConfig struct {
	AutoRestart   bool                `json:"-"`
	Context       map[string][]string `json:"-"`
	DisableBridge bool                `json:"-"`
	DNS           []string            `json:"dns,omitempty"`
	DNSOptions    []string            `json:"dns-opts,omitempty"`
	DNSSearch     []string            `json:"dns-search,omitempty"`
	GraphDriver   string              `json:"storage-driver,omitempty"`
	GraphOptions  []string            `json:"storage-opts,omitempty"`
	Labels        []string            `json:"labels,omitempty"`
	Mtu           int                 `json:"mtu,omitempty"`
	Pidfile       string              `json:"pidfile,omitempty"`
	RawLogs       bool                `json:"raw-logs,omitempty"`
	Root          string              `json:"data-root,omitempty"`
	SocketGroup   string              `json:"group,omitempty"`

	// MaxConcurrentDownloads is the maximum number of downloads that
	// may take place at a time for each pull.
	MaxConcurrentDownloads *int `json:"max-concurrent-downloads,omitempty"`

	// ShutdownTimeout is the timeout value (in seconds) the daemon will wait for the container
	// to stop when daemon is being shutdown
	ShutdownTimeout int `json:"shutdown-timeout,omitempty"`

	Debug    bool     `json:"debug,omitempty"`
	Hosts    []string `json:"hosts,omitempty"`
	LogLevel string   `json:"log-level,omitempty"`

	// Embedded structs that allow config
	// deserialization without the full struct.

	LogConfig
	BridgeConfig // bridgeConfig holds bridge network specific configuration.
	registry.ServiceOptions

	sync.Mutex
}

// New returns a new fully initialized Config struct
func New() *Config {
	config := Config{}
	config.LogConfig.Config = make(map[string]string)
	return &config
}

// GetConflictFreeLabels validates Labels for conflict
// In swarm the duplicates for labels are removed
// so we only take same values here, no conflict values
// If the key-value is the same we will only take the last label
func GetConflictFreeLabels(labels []string) ([]string, error) {
	labelMap := map[string]string{}
	for _, label := range labels {
		stringSlice := strings.SplitN(label, "=", 2)
		if len(stringSlice) > 1 {
			// If there is a conflict we will return an error
			if v, ok := labelMap[stringSlice[0]]; ok && v != stringSlice[1] {
				return nil, fmt.Errorf("conflict labels for %s=%s and %s=%s", stringSlice[0], stringSlice[1], stringSlice[0], v)
			}
			labelMap[stringSlice[0]] = stringSlice[1]
		}
	}

	newLabels := []string{}
	for k, v := range labelMap {
		newLabels = append(newLabels, fmt.Sprintf("%s=%s", k, v))
	}
	return newLabels, nil
}

// boolValue is an interface that boolean value flags implement
// to tell the command line how to make -name equivalent to -name=true.
type boolValue interface {
	IsBoolFlag() bool
}

// MergeDaemonConfigurations reads a configuration file,
// loads the file configuration in an isolated structure,
// and merges the configuration provided from flags on top
// if there are no conflicts.
func MergeDaemonConfigurations(flagsConfig *Config, flags *pflag.FlagSet, configFile string) (*Config, error) {
	fileConfig, err := getConflictFreeConfiguration(configFile, flags)
	if err != nil {
		return nil, err
	}

	if err := Validate(fileConfig); err != nil {
		return nil, fmt.Errorf("configuration validation from file failed (%v)", err)
	}

	// merge flags configuration on top of the file configuration
	if err := mergo.Merge(fileConfig, flagsConfig); err != nil {
		return nil, err
	}

	// We need to validate again once both fileConfig and flagsConfig
	// have been merged
	if err := Validate(fileConfig); err != nil {
		return nil, fmt.Errorf("merged configuration validation from file and command line flags failed (%v)", err)
	}

	return fileConfig, nil
}

// getConflictFreeConfiguration loads the configuration from a JSON file.
// It compares that configuration with the one provided by the flags,
// and returns an error if there are conflicts.
func getConflictFreeConfiguration(configFile string, flags *pflag.FlagSet) (*Config, error) {
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var jsonConfig map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&jsonConfig); err != nil {
		return nil, err
	}
	for key := range jsonConfig {
		if ignoredOptions[key] {
			logrus.Warnf("The %q config file option is no longer supported and is ignored", key)
			delete(jsonConfig, key)
		}
	}
	if v, ok := jsonConfig["graph"]; ok {
		logrus.Warn(`The "graph" config file option is deprecated. Please use "data-root" instead.`)
		if _, both := jsonConfig["data-root"]; both {
			return nil, fmt.Errorf(`cannot specify both "graph" and "data-root" config file options`)
		}
		jsonConfig["data-root"] = v
		delete(jsonConfig, "graph")
	}
	b, err = json.Marshal(jsonConfig)
	if err != nil {
		return nil, err
	}

	var config Config
	if flags != nil {
		configSet := configValuesSet(jsonConfig)

		if err := findConfigurationConflicts(configSet, flags); err != nil {
			return nil, err
		}

		// Override flag values to make sure the values set in the config file with nullable values, like `false`,
		// are not overridden by default truthy values from the flags that were not explicitly set.
		// See https://github.com/docker/docker/issues/20289 for an example.
		//
		// TODO: Rewrite configuration logic to avoid same issue with other nullable values, like numbers.
		namedOptions := make(map[string]interface{})
		for key, value := range configSet {
			f := flags.Lookup(key)
			if f == nil { // ignore named flags that don't match
				namedOptions[key] = value
				continue
			}

			if _, ok := f.Value.(boolValue); ok {
				f.Value.Set(fmt.Sprintf("%v", value))
			}
		}
		if len(namedOptions) > 0 {
			// set also default for mergeVal flags that are boolValue at the same time.
			flags.VisitAll(func(f *pflag.Flag) {
				if opt, named := f.Value.(opts.NamedOption); named {
					v, set := namedOptions[opt.Name()]
					_, boolean := f.Value.(boolValue)
					if set && boolean {
						f.Value.Set(fmt.Sprintf("%v", v))
					}
				}
			})
		}

	}

	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// configValuesSet returns the configuration values explicitly set in the file.
func configValuesSet(config map[string]interface{}) map[string]interface{} {
	flatten := make(map[string]interface{})
	for k, v := range config {
		if m, isMap := v.(map[string]interface{}); isMap && !flatOptions[k] {
			for km, vm := range m {
				flatten[km] = vm
			}
			continue
		}

		flatten[k] = v
	}
	return flatten
}

// findConfigurationConflicts iterates over the provided flags searching for
// duplicated configurations and unknown keys. It returns an error with all the conflicts if
// it finds any.
func findConfigurationConflicts(config map[string]interface{}, flags *pflag.FlagSet) error {
	// 1. Search keys from the file that we don't recognize as flags.
	unknownKeys := make(map[string]interface{})
	for key, value := range config {
		if flag := flags.Lookup(key); flag == nil {
			unknownKeys[key] = value
		}
	}

	// 2. Discard values that implement NamedOption.
	// Their configuration name differs from their flag name, like `labels` and `label`.
	if len(unknownKeys) > 0 {
		unknownNamedConflicts := func(f *pflag.Flag) {
			if namedOption, ok := f.Value.(opts.NamedOption); ok {
				if _, valid := unknownKeys[namedOption.Name()]; valid {
					delete(unknownKeys, namedOption.Name())
				}
			}
		}
		flags.VisitAll(unknownNamedConflicts)
	}

	if len(unknownKeys) > 0 {
		var unknown []string
		for key := range unknownKeys {
			unknown = append(unknown, key)
		}
		return fmt.Errorf("the following directives don't match any configuration option: %s", strings.Join(unknown, ", "))
	}

	var conflicts []string
	printConflict := func(name string, flagValue, fileValue interface{}) string {
		return fmt.Sprintf("%s: (from flag: %v, from file: %v)", name, flagValue, fileValue)
	}

	// 3. Search keys that are present as a flag and as a file option.
	duplicatedConflicts := func(f *pflag.Flag) {
		// search option name in the json configuration payload if the value is a named option
		if namedOption, ok := f.Value.(opts.NamedOption); ok {
			if optsValue, ok := config[namedOption.Name()]; ok {
				conflicts = append(conflicts, printConflict(namedOption.Name(), f.Value.String(), optsValue))
			}
		} else {
			// search flag name in the json configuration payload
			for _, name := range []string{f.Name, f.Shorthand} {
				if value, ok := config[name]; ok {
					conflicts = append(conflicts, printConflict(name, f.Value.String(), value))
					break
				}
			}
		}
	}

	flags.Visit(duplicatedConflicts)

	if len(conflicts) > 0 {
		return fmt.Errorf("the following directives are specified both as a flag and in the configuration file: %s", strings.Join(conflicts, ", "))
	}
	return nil
}

// Validate validates some specific configs.
// such as config.DNS, config.Labels, config.DNSSearch,
// as well as config.MaxConcurrentDownloads
func Validate(config *Config) error {
	// validate DNS
	for _, dns := range config.DNS {
		if _, err := opts.ValidateIPAddress(dns); err != nil {
			return err
		}
	}

	// validate DNSSearch
	for _, dnsSearch := range config.DNSSearch {
		if _, err := opts.ValidateDNSSearch(dnsSearch); err != nil {
			return err
		}
	}

	// validate Labels
	for _, label := range config.Labels {
		if _, err := opts.ValidateLabel(label); err != nil {
			return err
		}
	}
	// validate MaxConcurrentDownloads
	if config.MaxConcurrentDownloads != nil && *config.MaxConcurrentDownloads < 0 {
		return fmt.Errorf("invalid max concurrent downloads: %d", *config.MaxConcurrentDownloads)
	}

	return nil
}
