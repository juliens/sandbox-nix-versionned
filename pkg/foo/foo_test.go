package foo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPackageFlakeURL(t *testing.T) {
	n, err := New("./fixtures/all.json")
	require.NoError(t, err)
	url, err := n.GetPackageVersionnedFlakeURL("go", "1.21.1")
	require.NoError(t, err)

	require.Equal(t, "https://github.com/NixOS/nixpkgs/archive/commit_1.21.1.zip", url)
}

func TestGetBinaryVersionnedFlakeURL(t *testing.T) {
	n, err := New("./fixtures/all.json")
	require.NoError(t, err)

	url, err := n.GetBinaryVersionnedFlakeURL("go", "1.25.2")
	require.NoError(t, err)

	require.Equal(t, "https://github.com/NixOS/nixpkgs/archive/commit_1.25.2.zip", url)
}

func TestVersion(t *testing.T) {
	n, err := New("./fixtures/all.json")
	require.NoError(t, err)

	testCases := []struct {
		desc       string
		constraint string
		expected   string
	}{
		{
			desc:       "exact",
			constraint: "1.21.1",
			expected:   "1.21.1",
		},
		{
			desc:       "patch_wildcard",
			constraint: "1.21.*",
			expected:   "1.21.5",
		},
		{
			desc:       "patch_wildcard_not_last",
			constraint: "1.19.*",
			expected:   "1.19.7",
		},
		{
			desc:       "minor_wildcard",
			constraint: "1.*.*",
			expected:   "1.21.5",
		},
		{
			desc:       "only_wildcard",
			constraint: "1.*.*",
			expected:   "1.21.5",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, actual, err := n.GetBinaryVersionned("go", test.constraint)
			require.NoError(t, err)

			assert.Equal(t, "commit_"+test.expected, actual.Commit)
		})
	}

}
