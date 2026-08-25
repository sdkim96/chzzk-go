package main

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

func isEOF(msg tea.KeyPressMsg) bool {
	return slices.Contains(
		[]string{"ctrl+c", "q"},
		msg.String(),
	)
}
