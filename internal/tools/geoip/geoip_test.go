package geoip

import (
	"testing"
)

func TestOpen(t *testing.T) {
	record, err := City("5.182.16.30")
	if err != nil {
		t.Fatal(err)
	}

	if !record.HasData() {
		t.Fatal("expected record to have data")
		return
	}

}
