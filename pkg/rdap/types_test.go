package rdap

import (
	"encoding/json"
)

func TestError__unmarshal(t *testing.T) {
	in := `{
	"errorCode": 418,
	"title": "Your Beverage Choice is Not Available",
	"description":
	[
		"I know coffee has more ummppphhh.",
		"Sorry, dude!"
	]
}`

}
