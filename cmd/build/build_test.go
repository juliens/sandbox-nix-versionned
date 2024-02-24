package main

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
)

func TestErr(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), "nix", "--help")

	output, err := cmd.CombinedOutput()

	fmt.Println(string(output), err)

}
