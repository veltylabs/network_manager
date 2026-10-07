//go:build !wasm

package ui

import (
	"webtyp.com/svg"
	"webtyp.com/svg/sprite"
)

// Icons hosts IconID/IconSvg for discovery and use in config.
type Icons struct{}

func (m *Icons) IconID() string { return ID }

// A Wi-Fi glyph: the screens administer the network.
func (m *Icons) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(svg.Icon(ID), "0 0 24 24",
			sprite.Path("M12 18a2 2 0 1 0 0 4 2 2 0 0 0 0-4zm-5.66-3.66 1.42 1.42a6 6 0 0 1 8.48 0l1.42-1.42a8 8 0 0 0-11.32 0zM2.1 10.1l1.42 1.42a12 12 0 0 1 16.96 0l1.42-1.42a14 14 0 0 0-19.8 0z"),
		),
	)
}
