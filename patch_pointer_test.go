package diff_test

import (
	"encoding/json"
	"testing"

	"github.com/r3labs/diff/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatchPointerConversion(t *testing.T) {
	type patchStruct struct {
		Variable *int
	}
	type namedInt int
	value := 4
	cases := []struct {
		name         string
		to           interface{}
		needsOption  bool
		incompatible bool
		want         *int
	}{
		{name: "int", to: 4, want: &value},
		{name: "float", to: float64(4), needsOption: true, want: &value},
		{name: "named-int", to: namedInt(4), needsOption: true, want: &value},
		{name: "pointer", to: &value, want: &value},
		{name: "nil"},
		{name: "typed-nil", to: (*int)(nil)},
		{name: "incompatible", to: "four", incompatible: true},
	}
	for _, initial := range []string{"nil", "non-nil"} {
		for _, convert := range []bool{false, true} {
			mode := "default"
			if convert {
				mode = "convert"
			}
			for _, tc := range cases {
				t.Run(initial+"/"+mode+"/"+tc.name, func(t *testing.T) {
					original := 1
					var target patchStruct
					if initial == "non-nil" {
						target.Variable = &original
					}
					before := target.Variable
					var options []func(*diff.Differ) error
					if convert {
						options = append(options, diff.ConvertCompatibleTypes())
					}
					d, err := diff.NewDiffer(options...)
					require.NoError(t, err)
					log := d.Patch(diff.Changelog{{Type: diff.UPDATE, Path: []string{"Variable"}, From: before, To: tc.to}}, &target)
					require.Len(t, log, 1)
					if tc.incompatible || (tc.needsOption && !convert) {
						require.True(t, log.HasErrors(), "%+v", log)
						assert.True(t, log[0].HasFlag(diff.FlagFailed))
						assert.False(t, log[0].HasFlag(diff.FlagApplied))
						assert.Same(t, before, target.Variable)
					} else {
						require.False(t, log.HasErrors(), "%+v", log)
						assert.True(t, log[0].HasFlag(diff.FlagApplied))
						assert.False(t, log[0].HasFlag(diff.FlagFailed))
						assert.Equal(t, tc.want, target.Variable)
					}
					assert.Equal(t, 1, original)
				})
			}
		}
	}
}

func TestPatchNilPointerJSONRoundTrip(t *testing.T) {
	type patchStruct struct {
		Variable *int
	}
	value := 4
	before, after := patchStruct{}, patchStruct{Variable: &value}
	d, err := diff.NewDiffer(diff.ConvertCompatibleTypes())
	require.NoError(t, err)
	changes, err := d.Diff(before, after)
	require.NoError(t, err)
	data, err := json.Marshal(changes)
	require.NoError(t, err)
	var decoded diff.Changelog
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Len(t, decoded, 1)
	require.IsType(t, float64(0), decoded[0].To)
	log := d.Patch(decoded, &before)
	require.Len(t, log, 1)
	require.False(t, log.HasErrors(), "%+v", log)
	assert.True(t, log[0].HasFlag(diff.FlagApplied))
	assert.Equal(t, after, before)
}
