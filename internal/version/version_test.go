package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr error
	}{
		{
			name:  "Valid version",
			input: "1.2.3",
			want:  &Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:    "Invalid format",
			input:   "1.2",
			wantErr: ErrInvalidVersion,
		},
		{
			name:    "Invalid major",
			input:   "a.2.3",
			wantErr: ErrInvalidMajor,
		},
		{
			name:    "Invalid minor",
			input:   "1.b.3",
			wantErr: ErrInvalidMinor,
		},
		{
			name:    "Invalid patch",
			input:   "1.2.c",
			wantErr: ErrInvalidPatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersion_IsCompatible(t *testing.T) {
	tests := []struct {
		name    string
		version *Version
		other   *Version
		want    bool
	}{
		{
			name:    "Same major version",
			version: &Version{Major: 1, Minor: 0, Patch: 0},
			other:   &Version{Major: 1, Minor: 1, Patch: 0},
			want:    true,
		},
		{
			name:    "Different major version",
			version: &Version{Major: 1, Minor: 0, Patch: 0},
			other:   &Version{Major: 2, Minor: 0, Patch: 0},
			want:    false,
		},
		{
			name:    "Same version",
			version: &Version{Major: 1, Minor: 2, Patch: 3},
			other:   &Version{Major: 1, Minor: 2, Patch: 3},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.IsCompatible(tt.other)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersion_IsNewer(t *testing.T) {
	tests := []struct {
		name    string
		version *Version
		other   *Version
		want    bool
	}{
		{
			name:    "Newer major version",
			version: &Version{Major: 1, Minor: 0, Patch: 0},
			other:   &Version{Major: 2, Minor: 0, Patch: 0},
			want:    true,
		},
		{
			name:    "Older major version",
			version: &Version{Major: 2, Minor: 0, Patch: 0},
			other:   &Version{Major: 1, Minor: 0, Patch: 0},
			want:    false,
		},
		{
			name:    "Newer minor version",
			version: &Version{Major: 1, Minor: 0, Patch: 0},
			other:   &Version{Major: 1, Minor: 1, Patch: 0},
			want:    true,
		},
		{
			name:    "Newer patch version",
			version: &Version{Major: 1, Minor: 0, Patch: 0},
			other:   &Version{Major: 1, Minor: 0, Patch: 1},
			want:    true,
		},
		{
			name:    "Same version",
			version: &Version{Major: 1, Minor: 0, Patch: 0},
			other:   &Version{Major: 1, Minor: 0, Patch: 0},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.IsNewer(tt.other)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersion_String(t *testing.T) {
	v := &Version{Major: 1, Minor: 2, Patch: 3}
	assert.Equal(t, "1.2.3", v.String())
}

func TestExtractFromNotes(t *testing.T) {
	tests := []struct {
		name  string
		notes string
		want  string
	}{
		{
			name: "Valid notes",
			notes: `# Release Notes
## Engine Version
1.2.3
## Other Section`,
			want: "1.2.3",
		},
		{
			name: "No engine version",
			notes: `# Release Notes
## Other Section`,
			want: "",
		},
		{
			name:  "Empty notes",
			notes: "",
			want:  "",
		},
		{
			name: "Engine version at end",
			notes: `# Release Notes
## Engine Version
1.2.3`,
			want: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractFromNotes(tt.notes)
			assert.Equal(t, tt.want, got)
		})
	}
}
