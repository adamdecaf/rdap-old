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

func TestError__verisignlabs(t *testing.T) {
	in := []byte(`{"notices":[{"description":["Service subject to Terms of Use."],"links":[{"href":"http:\/\/rdap-pilot.verisignlabs.com\/terms_of_use","type":"text\/html"}],"title":"Terms of Use"}],"rdapConformance":["rdap_level_0","rdap_objectTag_level_0","rdap_openidc_level_0"],"errorCode":400,"description":["bad path"],"lang":"en-US","title":"Error in processing the request."}`)

	var e Error
	if err := json.Unmarshal(in, &e); err != nil {
		t.Fatal(e)
	}
	if e.Code != 400 {
		t.Errorf("got %d", e.Code)
	}
	if e.Title != "Error in processing the request." {
		t.Errorf("got %s", e.Title)
	}
	if len(e.Description) != 1 {
		if e.Description[0] != "bad path" {
			t.Error(e.Description)
		}
		t.Errorf("len(e.Description)=%d", len(e.Description))
	}
}
