package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Chat struct {
	id      int
	tab     *container.TabItem
	content *fyne.Container
}

type Gui struct {
	app    fyne.App
	window fyne.Window

	globalBox *fyne.Container

	idLabel *widget.Label

	// global container
	chatTabs *container.AppTabs

	// list of single chats
	chats []*Chat

	input          *widget.Entry
	messageChannel chan *Message
}

func (g *Gui) findCurrentChatId() (int, error) {
	currentTab := g.chatTabs.Selected()

	for i := 0; i < len(g.chats); i++ {
		if g.chats[i].tab == currentTab {
			return g.chats[i].id, nil
		}
	}

	return -1, fmt.Errorf("no tab was found")
}

func (g *Gui) SendMessage(message string) {
	m := new(Message)

	currId, err := g.findCurrentChatId()

	if err != nil {
		panic(err)
	}

	m.Init(currId, []byte(message), true)
	g.messageChannel <- m

	// resent input field
	g.input.SetText("")
	g.AddMessage(m)
}

func (g *Gui) Init(windowName string, messageChannel chan *Message) {
	g.app = app.New()
	g.window = g.app.NewWindow(windowName)
	g.messageChannel = messageChannel

	g.idLabel = widget.NewLabel("Waiting...")

	header := container.New(layout.NewHBoxLayout(), g.idLabel)
	g.chatTabs = container.NewAppTabs()

	g.input = widget.NewEntry()
	g.input.SetPlaceHolder("message")
	g.input.OnSubmitted = g.SendMessage

	g.globalBox = container.New(layout.NewVBoxLayout(), header, g.chatTabs, layout.NewSpacer(), g.input)

	g.window.SetContent(g.globalBox)

}

func (g *Gui) AddMessage(message *Message) {
	var chat *Chat
	found := false

	for i := 0; i < len(g.chats) && !found; i++ {
		if g.chats[i].id == message.ClientId {
			found = true
			chat = g.chats[i]
		}
	}

	if !found {
		g.addNewChat(message)
	} else {
		chat.content.Add(message.GetLabel())
	}
}

func (g *Gui) addNewChat(message *Message) {
	// append a new tab with a new message inside
	chat := new(Chat)
	chat.id = message.ClientId
	chat.content = container.New(layout.NewVBoxLayout(), message.GetLabel())
	chat.tab = container.NewTabItem(fmt.Sprintf("%d", message.ClientId), chat.content)

	g.chats = append(g.chats, chat)

	g.chatTabs.Append(chat.tab)

}

func (g *Gui) UpdateId(id int) {
	g.idLabel.SetText(fmt.Sprintf("ID: %d", id))
}

// blocking
func (g *Gui) Run() {
	g.window.ShowAndRun()
}
