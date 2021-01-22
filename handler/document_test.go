package handler

import (
	"testing"
)

func TestDocumentHandler(t *testing.T) {
	r := testInit(t)
	token := testGetToken(t, r)

	t.Run("CreateDoc", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("GetDocInfo", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("UpdateDocInfo", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("MoveDoc", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("RemoveDoc", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("DocWebSocket", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
}
