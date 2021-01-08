package handler

import (
	"testing"
)

func TestSearchHandler(t *testing.T) {
	r := testInit(t)
	token := testGetToken(t, r)

	t.Run("SearchUser", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("SearchTeam", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
}
