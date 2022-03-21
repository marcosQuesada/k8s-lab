package http

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"net"
	"testing"
)

// @TODO: test it properly
func TestVersionProvider_ItRepliesDefaultAssignation(t *testing.T) {
	t.Skip()
	vp := NewVersionProvider("9090")

	v, err := vp.Assignation(context.Background(), net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error getting assignation, error %v", err)
	}

	spew.Dump(v)
}
