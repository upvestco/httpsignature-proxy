package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nsf/termbox-go"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"github.com/upvestco/httpsignature-proxy/service/ui/elements"
	"github.com/upvestco/httpsignature-proxy/service/ui/window"
	"golang.design/x/clipboard"
)

const (
	eventsListWidth              = 40
	defaultHeadersWindowHeight   = 10
	minimizedHeadersWindowHeight = 1

	copyEventButtonText        = " Copy event "
	copyHeadersButtonText      = " Copy headers "
	copyLogsText               = " Copy logs "
	exitButtonText             = " Exit "
	toggleEventsText           = " WH Events "
	toggleEventsTextWithNumber = " WH Events(%d) "
	toggleLogsText             = " Proxy log "
	toggleLogsTextWithNumber   = " Proxy log(%d) "
	hideHeadersTest            = " Hide headers "
	showHeadersTest            = " Show headers "

	eventsScreenID = "eventsScreen"
	logScreenID    = "logsScreen"
)

var (
	mainWindow          *window.Window
	events              *window.Cards
	headers             *window.Cards
	eventsList          *elements.SelectView
	logsList            *elements.SelectView
	selected            *Event
	headersWindowHeight = defaultHeadersWindowHeight
	created             bool
)

func IsCreated() bool {
	return created
}

func AddPayload(payload PullItem) {
	if !created {
		return
	}
	var in Payload
	if err := json.Unmarshal([]byte(payload.Payload), &in); err != nil {
		AddLogs(err.Error())
		return
	}
	for i, ev := range in.Payload {
		id := uuid.NewString()

		v := elements.NewJSONView(window.WholeArea())
		v.Set(ev)
		events.Insert(id, v)
		var s string = " "
		if len(in.Payload) > 1 {
			if i == 0 {
				s = "┌"
			} else if i == len(in.Payload)-1 {
				s = "└"
			} else {
				s = "│"
			}
		}
		title := fmt.Sprintf("%s %s %s", s, time.Now().Format(time.DateTime), ev.Type)
		event := NewEvent(id, title, []byte(payload.Payload), payload.Headers)
		eventsList.Append(event)

		hdr := elements.NewHeadersView(window.WholeArea())
		hdr.SetKeyColor(termbox.ColorLightBlue)
		hdr.SetHeaders(payload.Headers)
		headers.Insert(id, hdr)
	}
	mainWindow.Update(events, headers)
}

func AddLogs(m string) {
	for _, s := range strings.Split(m, "\n") {
		logsList.Append(NewEvent("", s, nil, nil))
	}
	mainWindow.Update(logsList)
}

func Close() {
	mainWindow.Close()
}

func Create(onClose func()) {
	mainWindow = window.New(termbox.ColorGreen, termbox.ColorDefault)
	mainFrame := elements.NewFrame(window.WholeArea(), window.NormalFrameStyle)
	mainWindow.Add(mainFrame)

	screens := window.CreateCards(window.ShrinkAreaTransformer(1, 1))
	mainFrame.Add(screens)

	eventsScreen := elements.NewContainer(window.WholeArea())
	logsScreen := elements.NewContainer(window.WholeArea())

	exitButton := elements.NewButton(mainWindow, exitButtonText, func(parent window.Area) window.Area {
		return window.HLine(parent.Right(2+len(exitButtonText)),
			parent.Right(2), parent.OnTop())
	}).OnRelease(func() {
		mainWindow.Close()
		os.Exit(0)
	})
	mainFrame.Add(exitButton)

	getToggleEventsText := func(all int) string {
		if all == 0 {
			return toggleEventsText
		}
		return fmt.Sprintf(toggleEventsTextWithNumber, all)
	}
	toggleEvents := elements.NewToggle(mainWindow, getToggleEventsText(0), nil)
	toggleEvents.SetTransformer(func(parent window.Area) window.Area {
		return window.HLine(parent.Right(len(toggleEvents.GetText())+2),
			parent.Right(2), parent.OnBottom())
	})

	mainFrame.Add(toggleEvents)

	getToggleLogsText := func(all int) string {
		if all == 0 {
			return toggleLogsText
		}
		return fmt.Sprintf(toggleLogsTextWithNumber, all)
	}

	toggleLogs := elements.NewToggle(mainWindow, getToggleLogsText(0), nil)
	toggleLogs.SetTransformer(func(parent window.Area) window.Area {
		return window.HLine(parent.Right(2+len(toggleEvents.GetText())+2+len(toggleLogs.GetText())),
			parent.Right(2+len(toggleEvents.GetText())), parent.OnBottom())
	})

	mainFrame.Add(toggleLogs)

	vLine := elements.NewVLine(func(parent window.Area) window.Area {
		return window.VLine(parent.Top(-1), parent.Bottom(-1), parent.Left(eventsListWidth+1))
	})

	eventsScreen.Add(vLine)

	hLine := elements.NewHLine(func(parent window.Area) window.Area {
		return window.HLine(parent.Left(eventsListWidth+1), parent.Right(-1), parent.Bottom(headersWindowHeight-1))
	})
	eventsScreen.Add(hLine)

	events = window.CreateCards(func(parent window.Area) window.Area {
		return window.Rectangle(window.NewPoint(parent.Left(eventsListWidth+2), parent.OnTop()),
			window.NewPoint(parent.OnRight(), parent.Bottom(headersWindowHeight)))
	})
	eventsScreen.Add(events)

	// events screen
	headers = window.CreateCards(func(parent window.Area) window.Area {
		return window.Rectangle(window.NewPoint(parent.Left(eventsListWidth+2), parent.Bottom(headersWindowHeight-2)), parent.BottomRight())
	})
	eventsScreen.Add(headers)

	eventsList = elements.NewSelectView(func(parent window.Area) window.Area {
		return window.Rectangle(parent.TopLeft(), window.NewPoint(parent.Left(eventsListWidth), parent.OnBottom()))
	})
	eventsList.SetVisitedColor(termbox.ColorDarkGray)
	eventsList.OnSelect(func(l interface{}) {
		if p, ok := l.(*Event); ok {
			selected = p
			events.BringUp(p.id)
			headers.BringUp(p.id)
			toggleEvents.SetText(getToggleEventsText(eventsList.GetNotVisited()))
			mainWindow.Update(events, headers, mainFrame)
		}
	})
	eventsList.OnChange(func() {
		toggleEvents.SetText(getToggleEventsText(eventsList.GetNotVisited()))
		if !toggleEvents.IsPressed() {
			toggleEvents.SetCustomStyle(termbox.AttrBlink)
		} else {
			toggleLogs.SetCustomStyle(0)
		}
		mainWindow.Update(mainFrame)
	})
	eventsScreen.Add(eventsList)

	copyHeadersButton := elements.NewButton(mainWindow, copyHeadersButtonText, func(parent window.Area) window.Area {
		return window.HLine(parent.Left(2+len(copyEventButtonText)+2), parent.Left(2+len(copyEventButtonText)+2+len(copyHeadersButtonText)), parent.OnTop())
	}).OnPress(func() {
		if selected == nil {
			return
		}
		buf := bytes.Buffer{}
		for _, line := range elements.FormatHeaders(selected.headers) {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		clipboard.Write(clipboard.FmtText, buf.Bytes())
	})
	hLine.Add(copyHeadersButton)

	showHideHeadersToggle := elements.NewToggle(mainWindow, hideHeadersTest, func(parent window.Area) window.Area {
		return window.HLine(parent.Right(len(hideHeadersTest)+2), parent.Right(1), parent.OnBottom())
	})
	showHideHeadersToggle.SetReleasedStyle(mainWindow.DefaultButtonStyle(false))
	showHideHeadersToggle.SetPressedStyle(mainWindow.DefaultButtonStyle(false))
	showHideHeadersToggle.OnPress(func() {
		showHideHeadersToggle.SetText(showHeadersTest)
		headersWindowHeight = minimizedHeadersWindowHeight
		copyHeadersButton.SetHidden()
		mainWindow.Update(mainFrame)
	})

	showHideHeadersToggle.OnRelease(func() {
		showHideHeadersToggle.SetText(hideHeadersTest)
		headersWindowHeight = defaultHeadersWindowHeight
		copyHeadersButton.SetVisible()
		mainWindow.Update(mainFrame)
	})
	hLine.Add(showHideHeadersToggle)

	copyEventButton := elements.NewButton(mainWindow, copyEventButtonText, func(parent window.Area) window.Area {
		return window.HLine(parent.Left(2), parent.Left(2+len(copyEventButtonText)), parent.OnTop())
	}).OnPress(func() {
		if selected == nil {
			return
		}
		clipboard.Write(clipboard.FmtText, selected.source)
	})
	hLine.Add(copyEventButton)

	logsList = elements.NewSelectView(window.WholeArea())
	logsList.SetColor(window.Color{
		FG: termbox.ColorDefault,
		BG: termbox.ColorDefault,
	})
	logsList.OnChange(func() {
		if toggleLogs.IsPressed() {
			logsList.MarkAllVisited()
		}
		toggleLogs.SetText(getToggleLogsText(logsList.GetNotVisited()))
		if !toggleLogs.IsPressed() {
			toggleLogs.SetCustomStyle(termbox.AttrBlink)
		} else {
			toggleLogs.SetCustomStyle(0)
		}
		mainWindow.Update(mainFrame)
	})
	logsScreen.Add(logsList)

	copyLogsButton := elements.NewButton(mainWindow, copyLogsText, func(parent window.Area) window.Area {
		return window.HLine(parent.Left(2),
			parent.Left(2+len(copyLogsText)), parent.Top(0))
	}).OnPress(func() {
		b := bytes.Buffer{}
		for _, item := range logsList.GetItems() {
			b.WriteString(item.String() + "\n")
		}
		clipboard.Write(clipboard.FmtText, b.Bytes())
	})

	mainFrame.Add(copyLogsButton)

	toggleEvents.OnPress(func() {
		screens.BringUp(eventsScreenID)
		toggleLogs.Release()
		toggleLogs.Enabled()
		toggleEvents.SetCustomStyle(0)
		toggleEvents.Disabled()
		copyLogsButton.SetHidden()
		mainWindow.Update(mainFrame)
	})

	toggleLogs.OnPress(func() {
		screens.BringUp(logScreenID)
		toggleEvents.Release()
		toggleEvents.Enabled()
		logsList.MarkAllVisited()
		toggleLogs.SetText(getToggleLogsText(logsList.GetNotVisited()))
		toggleLogs.SetCustomStyle(0)
		toggleLogs.Disabled()
		copyLogsButton.SetVisible()
		mainWindow.Update(mainFrame)
	})

	screens.Insert(eventsScreenID, eventsScreen)
	screens.Insert(logScreenID, logsScreen)
	toggleLogs.Press()

	go func() {
		mainWindow.Run(termbox.KeyCtrlC)
		onClose()
	}()
	created = true
}
func NewEvent(id, str string, source []byte, headers http.Header) *Event {
	return &Event{
		lines:   str,
		id:      id,
		source:  source,
		headers: headers,
	}
}

type Event struct {
	id      string
	lines   string
	source  []byte
	headers http.Header
}

func (l Event) String() string {
	return l.lines
}

type PullItem struct {
	Headers   http.Header `json:"headers"`
	Payload   string      `json:"payload"`
	CreatedAt time.Time   `json:"created_at"`
}

type Payload struct {
	Payload []struct {
		CreatedAt time.Time              `json:"created_at"`
		Id        string                 `json:"id"`
		Object    map[string]interface{} `json:"object"`
		Type      string                 `json:"type"`
		WebhookId string                 `json:"webhook_id"`
	} `json:"payload"`
}

func CreateLogger(verboseMode bool) logger.Logger {
	return &uiLogger{
		verboseMode: verboseMode,
	}
}

type uiLogger struct {
	verboseMode bool
}

func (e *uiLogger) Log(message string) {
	if e.verboseMode {
		AddLogs(message)
	}
}
func (e *uiLogger) LogF(format string, a ...interface{}) {
	if e.verboseMode {
		AddLogs(fmt.Sprintf(format, a...))
	}
}

func (e *uiLogger) PrintF(format string, a ...interface{}) {
	AddLogs(fmt.Sprintf(format, a...))
}
func (e *uiLogger) PrintLn(message string) {
	AddLogs(message)
}
