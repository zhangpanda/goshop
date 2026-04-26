package service

import "testing"

func TestMd5Sign(t *testing.T) {
	got := md5Sign("hello")
	if len(got) != 32 {
		t.Fatalf("md5Sign length = %d; want 32", len(got))
	}
	// MD5("hello") = 5D41402ABC4B2A76B9719D911017C592
	if got != "5D41402ABC4B2A76B9719D911017C592" {
		t.Errorf("md5Sign(hello) = %s; want 5D41402ABC4B2A76B9719D911017C592", got)
	}
}
