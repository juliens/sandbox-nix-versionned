package foo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPackageFlakeURL(t *testing.T) {
	n, err := New("./fixtures/all.json")
	require.NoError(t, err)
	url, err := n.GetPackageVersionnedFlakeURL("go", "1.21.1")
	require.NoError(t, err)

	require.Equal(t, "https://github.com/NixOS/nixpkgs/archive/2dde8b588897.zip", url)
}

func TestGetBinaryVersionnedFlakeURL(t *testing.T) {
	n, err := New("./fixtures/all.json")
	require.NoError(t, err)

	url, err := n.GetBinaryVersionnedFlakeURL("go", "1.25.2")
	require.NoError(t, err)

	require.Equal(t, "https://github.com/NixOS/nixpkgs/archive/commit_1_25_2.zip", url)
}
