package handler

import (
	"testing"
)

func TestImageHandler(t *testing.T) {
	r := testInit(t)
	token := testGetToken(t, r)

	t.Run("UploadImage", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("GetImage", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
}
