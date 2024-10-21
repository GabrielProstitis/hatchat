package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Message struct {
	ClientId int
	Contents []byte
	isSender bool
}

func (m *Message) Init(clientId int, contents []byte, isSender bool) {
	m.ClientId = clientId
	m.Contents = contents
	m.isSender = isSender
}

func (m *Message) GetLabel() *fyne.Container {
	if m.isSender {
		return container.New(layout.NewHBoxLayout(), widget.NewLabel(string(m.Contents)), layout.NewSpacer())
	} else {
		return container.New(layout.NewHBoxLayout(), layout.NewSpacer(), widget.NewLabel(string(m.Contents)))
	}
}
