package bootstrap

import (
	"time"
)

type Response struct {
	Version     string       `json:"version"`
	Publication time.Time    `json:"publication,omitempty"`
	Description string       `json:"description,omitempty"`
	Services    [][][]string `json:"services"`
}
