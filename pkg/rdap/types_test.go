package rdap

import (
	"encoding/json"
	"testing"
)

func TestError__unmarshal(t *testing.T) {
	in := []byte(`{
	"errorCode": 418,
	"title": "Your Beverage Choice is Not Available",
	"description":
	[
		"I know coffee has more ummppphhh.",
		"Sorry, dude!"
	]
}`)
	var e Error
	if err := json.Unmarshal(in, &e); err != nil {
		t.Fatal(e)
	}
	if e.Code != 418 {
		t.Errorf("got %d", e.Code)
	}
	if e.Title != "Your Beverage Choice is Not Available" {
		t.Errorf("got %s", e.Title)
	}
	if len(e.Description) != 2 {
		t.Errorf("len(e.Description)=%d", len(e.Description))
	}
}
