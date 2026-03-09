package scanner

import "testing"

func TestNewNERScanner_NilSession(t *testing.T) {
	ns := NewNERScanner(nil, 0.5)
	if ns != nil {
		t.Error("NewNERScanner with nil session should return nil")
	}
}

func TestNERScanner_NilScan(t *testing.T) {
	// A nil NERScanner should safely return nil on Scan.
	var ns *NERScanner
	result := ns.Scan("hello world")
	if result != nil {
		t.Errorf("nil NERScanner.Scan should return nil, got %v", result)
	}
}
