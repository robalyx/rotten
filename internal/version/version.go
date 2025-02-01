package version

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidVersion = errors.New("invalid version")
	ErrInvalidMajor   = errors.New("invalid major version")
	ErrInvalidMinor   = errors.New("invalid minor version")
	ErrInvalidPatch   = errors.New("invalid patch version")
)

// EngineVersion is the version of the engine, set from export_config.json at build time.
// Please use the -ldflags option to set this value at build time.
var EngineVersion = "0.0.0" //nolint:gochecknoglobals

// Version represents a semantic version.
type Version struct {
	Major int
	Minor int
	Patch int
}

// Parse converts a version string to a Version struct.
func Parse(v string) (*Version, error) {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidVersion
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidMajor, err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidMinor, err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPatch, err)
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// IsCompatible checks if the given version is compatible with this version.
func (v *Version) IsCompatible(other *Version) bool {
	// Major version must match exactly for compatibility
	return v.Major == other.Major
}

// IsNewer checks if the other version is newer than this version.
func (v *Version) IsNewer(other *Version) bool {
	// Compare major version first
	if other.Major > v.Major {
		return true
	}
	if other.Major < v.Major {
		return false
	}
	// If major versions are equal, compare minor version
	if other.Minor > v.Minor {
		return true
	}
	if other.Minor < v.Minor {
		return false
	}
	// If minor versions are equal, compare patch version
	return other.Patch > v.Patch
}

// String returns the string representation of the version.
func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// ExtractFromNotes extracts the engine version from release notes.
func ExtractFromNotes(notes string) string {
	lines := strings.Split(notes, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Engine Version" && i+1 < len(lines) {
			return strings.TrimSpace(lines[i+1])
		}
	}
	return ""
}
