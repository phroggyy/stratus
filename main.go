package main

import (
	"flag"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/rivo/tview"
	"google.golang.org/protobuf/types/descriptorpb"
	"os"
	"path/filepath"
	"strings"
)

var (
	kafkaHost = flag.String("kafka-host", "localhost:9090", "Set the host to your kafka instance")
)

var app = tview.NewApplication()
var layout = tview.NewGrid()
var fileBrowser = tview.NewList()
var messageEditor = tview.NewForm()
var currentDir = ""
var descriptors = map[string]*descriptorpb.DescriptorProto{}

func init() {
	configureGridLayout()
	configureFileBrowser()
	configureMessageEditor()
}

func configureGridLayout() {
	layout.SetColumns(-1, -1).SetRows(-1, -1)
}

func configureFileBrowser() {
	fileBrowser.SetBorder(true)
	fileBrowser.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			updateFileBrowser("..", true)
		}

		return event
	})

	fileBrowser.SetSelectedFunc(func(_ int, itemName string, _ string, _ rune) {
		if filepath.Ext(currentDir) == ".proto" {
			showProtobufForm(itemName)
			return
		}

		updateFileBrowser(itemName, true)
	})
}

func configureMessageEditor() {
	messageEditor.SetBorder(true)
	messageEditor.SetTitle("Protobuf message editor")
	messageEditor.SetFieldBackgroundColor(tcell.ColorDimGrey)
	messageEditor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.SetFocus(fileBrowser)
		}

		return event
	})
}

func main() {
	flag.Parse()
	cwd, _ := os.Getwd()
	root := tview.NewFrame(layout).
		AddText("Esc to navigate back", true, tview.AlignLeft, tcell.ColorLightGray)

	layout.AddItem(fileBrowser, 0, 0, 2, 1, 0, 0, true)
	layout.AddItem(messageEditor, 0, 1, 2, 1, 0, 0, true)
	updateFileBrowser(cwd, false)

	app.SetRoot(root, true).Run()
}

func updateFileBrowser(dir string, relative bool) error {
	if relative {
		currentDir = filepath.Join(currentDir, dir)
	} else {
		currentDir = dir
	}

	fileBrowser.Clear()
	fileBrowser.SetTitle(currentDir)
	descriptors = map[string]*descriptorpb.DescriptorProto{}
	// if we're "inside" a proto file, we should render messages instead of directories
	if filepath.Ext(currentDir) == ".proto" {
		// read the file
		p := protoparse.Parser{}

		d, err := p.ParseFilesButDoNotLink(currentDir)
		if err != nil {
			return err
		}
		fileBrowser.AddItem("..", "", 0, nil)

		for _, descriptor := range d {
			for _, mdesc := range descriptor.MessageType {
				// fully qualified message name
				fqmn := fmt.Sprintf("%s.%s", descriptor.GetPackage(), mdesc.GetName())
				descriptors[fqmn] = mdesc
				fileBrowser.AddItem(fqmn, "", 0, nil)
			}
		}
		return nil
	}

	dirs, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}

	if currentDir != "/" {
		fileBrowser.AddItem("..", "", 0, nil)
	}
	for _, dirItem := range dirs {
		if itemName := dirItem.Name(); dirItem.IsDir() || filepath.Ext(itemName) == ".proto" {
			fileBrowser.AddItem(itemName, "", 0, nil)
		}
	}

	return nil
}

func showProtobufForm(fqmn string) {
	// we assume that the descriptors have already been populated
	messageEditor.Clear(true)
	messageEditor.SetTitle(fmt.Sprintf("Message: %s", fqmn))
	messageEditor.AddInputField("Event", "", 0, nil, nil)

	for _, field := range descriptors[fqmn].Field {
		fieldName := field.GetName()
		fieldType := strings.ToLower(strings.TrimPrefix(field.GetType().String(), "TYPE_"))
		fieldModifier := strings.ToLower(strings.TrimPrefix(field.GetLabel().String(), "LABEL_"))
		messageEditor.AddInputField(fmt.Sprintf("%s (%s %s)", fieldName, fieldModifier, fieldType), field.GetDefaultValue(), 0, nil, nil)
	}
	messageEditor.AddButton("Send", func() {
		// send message over kafka
		//root.AddText("Sending message!", true, tview.AlignLeft, tcell.ColorGreen)
	})
	//layout.RemoveItem(editorPlaceholder)
	//layout.AddItem(messageEditor, 0, 1, 2, 1, 0, 0, true)
	app.SetFocus(messageEditor)
}
