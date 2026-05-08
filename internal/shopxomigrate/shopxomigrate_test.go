package shopxomigrate

import (
	"strings"
	"testing"
)

func TestRenderSQL_ReplacesDBAndPrefix(t *testing.T) {
	s := RenderSQL("srcdb", "dstdb", "pre_")
	if !strings.Contains(s, "`srcdb`.pre_user") {
		t.Fatalf("expected srcdb + pre_user, sample:\n%s", s[:min(400, len(s))])
	}
	if strings.Contains(s, "sxo_user") {
		t.Fatal("prefix replace should remove sxo_user")
	}
	if !strings.Contains(s, "`dstdb`.`users`") && !strings.Contains(s, "`dstdb`.users") {
		t.Fatal("expected dst db on insert")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
