package logger

import (
	"strings"

	"github.com/blendlabs/go-util/env"
)

var (
	// DefaultFlags is the default verbosity for a diagnostics agent inited from the environment.
	DefaultFlags = NewFlagSet(Fatal, Error, Info)
)

// NewFlagSet returns a new FlagSet with the given flags enabled.
func NewFlagSet(flags ...Flag) *FlagSet {
	efs := &FlagSet{
		flags: make(map[Flag]bool),
	}
	for _, flag := range flags {
		efs.Enable(flag)
	}
	return efs
}

// NewFlagSetAll returns a new FlagSet with all flags enabled.
func NewFlagSetAll() *FlagSet {
	return &FlagSet{
		flags: make(map[Flag]bool),
		all:   true,
	}
}

// NewFlagSetNone returns a new FlagSet with no flags enabled.
func NewFlagSetNone() *FlagSet {
	return &FlagSet{
		flags: make(map[Flag]bool),
		none:  true,
	}
}

// NewFlagSetFromEnv returns a new FlagSet from the environment.
func NewFlagSetFromEnv() *FlagSet {
	envEventsFlag := env.Env().String(EnvVarLogEvents)
	if len(envEventsFlag) > 0 {
		return NewFlagSetFromCSV(envEventsFlag)
	}
	return NewFlagSet()
}

// NewFlagSetFromCSV returns a new event flag set from a csv of event flags.
// These flags are case insensitive.
func NewFlagSetFromCSV(flagCSV string) *FlagSet {
	flagSet := &FlagSet{
		flags: map[Flag]bool{},
	}

	flags := strings.Split(flagCSV, ",")

	for _, flag := range flags {
		parsedFlag := Flag(strings.Trim(strings.ToLower(flag), " \t\n"))
		if string(parsedFlag) == string(FlagAll) {
			flagSet.all = true
		}

		if string(parsedFlag) == string(FlagNone) {
			flagSet.none = true
		}

		if strings.HasPrefix(string(parsedFlag), "-") {
			flag := Flag(strings.TrimPrefix(string(parsedFlag), "-"))
			flagSet.flags[flag] = false
		} else {
			flagSet.flags[parsedFlag] = true
		}
	}

	return flagSet
}

// FlagSet is a set of event flags.
type FlagSet struct {
	flags map[Flag]bool
	all   bool
	none  bool
}

// Enable enables an event flag.
func (efs *FlagSet) Enable(flag Flag) {
	efs.none = false
	efs.flags[flag] = true
}

// Disable disabled an event flag.
func (efs *FlagSet) Disable(flag Flag) {
	efs.flags[flag] = false
}

// SetAll flips the `all` bit on the flag set.
func (efs *FlagSet) SetAll() {
	efs.all = true
	efs.none = false
}

// All returns if the all bit is flipped on.
func (efs *FlagSet) All() bool {
	return efs.all
}

// SetNone flips the `none` bit on the flag set.
func (efs *FlagSet) SetNone() {
	efs.all = false
	efs.flags = map[Flag]bool{}
	efs.none = true
}

// None returns if the none bit is flipped on.
func (efs *FlagSet) None() bool {
	return efs.none
}

// IsEnabled checks to see if an event is enabled.
func (efs FlagSet) IsEnabled(flag Flag) bool {
	if efs.all {
		// figure out if we explicitly disabled the flag.
		if enabled, hasEvent := efs.flags[flag]; hasEvent && !enabled {
			return false
		}
		return true
	}
	if efs.none {
		return false
	}
	if efs.flags != nil {
		if enabled, hasFlag := efs.flags[flag]; hasFlag {
			return enabled
		}
	}
	return false
}

func (efs FlagSet) String() string {
	if efs.none {
		return string(FlagNone)
	}

	var flags []string
	if efs.all {
		flags = []string{string(FlagAll)}
	}
	for key, enabled := range efs.flags {
		if key != FlagAll {
			if enabled {
				if !efs.all {
					flags = append(flags, string(key))
				}
			} else {
				flags = append(flags, "-"+string(key))
			}
		}
	}
	return strings.Join(flags, ", ")
}

// CoalesceWith sets the set from another, with the other taking precedence.
func (efs *FlagSet) CoalesceWith(other *FlagSet) {
	if other.all {
		efs.all = true
	}
	if other.none {
		efs.none = true
	}
	for key, value := range other.flags {
		efs.flags[key] = value
	}
}
