package exec

import (
	"context"
	"testing"
)

func TestCommandContext(t *testing.T) {
	_ = CommandContext(context.Background(), "printf", `"bar\nbaz"`)
}
