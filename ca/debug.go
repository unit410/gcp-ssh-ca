package ca

import "log"

// DebugEnabled toggles debug logging
var DebugEnabled = false

func debugf(format string, v ...interface{}) {
	if DebugEnabled {
		log.Printf(format, v...)
	}
}
func debugln(v ...interface{}) {
	if DebugEnabled {
		log.Println(v...)
	}
}
