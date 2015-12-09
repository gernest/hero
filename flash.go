package hero

import (
	"encoding/gob"
)

func init() {
	gob.Register(&FlashMessage{})
	gob.Register(FlashMessages{})
}

//FlashMessage is session bassed flash message
type FlashMessage struct {
	// Kind is the type of the flash message
	// e.g error, info etc
	Kind string

	// text is the flash message body
	Text string
}

// FlashMessages is a slice of flash messages
type FlashMessages []*FlashMessage
