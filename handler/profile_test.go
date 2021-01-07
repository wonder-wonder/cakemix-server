package handler

import (
	"testing"
)

func TestProfileHandler(t *testing.T) {
	r, _ := testInit(t)
	token := testGetToken(t, r)

	t.Run("GetProfile", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("UpdateProfile", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
}
