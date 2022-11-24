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

var fileBrowser = tview.NewList()
var currentDir = ""

func updateFileBrowser(dir string, relative bool) (*tview.List, error) {
	//browser := tview.NewList()
	if relative {
		currentDir = filepath.Join(currentDir, dir)
	} else {
		currentDir = dir
	}

	dirs, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, err
	}

	fileBrowser.Clear()
	for _, dirItem := range dirs {
		if itemName := dirItem.Name(); dirItem.IsDir() || filepath.Ext(itemName) == "proto" {
			fileBrowser.AddItem(itemName, "", 0, nil)
		}
	}
}

func init() {
	fileBrowser.SetBorder(true)
	fileBrowser.SetSelectedFunc(func(_ int, itemName string, _ string, _ rune) {

	})
}

func main() {
	flag.Parse()
	app := tview.NewApplication()
	cwd, _ := os.Getwd()
	layout := tview.NewGrid()
	layout.SetColumns(-1, -1).SetRows(-1, -1)
	root := tview.NewFrame(layout).
		AddText("(Ctrl+arrow) navigate", true, tview.AlignLeft, tcell.ColorLightGray).
		AddText("(Ctrl-S) send message", true, tview.AlignLeft, tcell.ColorLightGray)

	serviceList := tview.NewList()
	serviceList.SetBorder(true).SetTitle("Services")

	//editor := tview.NewTextArea().SetTitle("Message content").SetBorder(true)

	editorPlaceholder := tview.NewBox().SetTitle("Message content").SetBorder(true)
	layout.AddItem(serviceList, 0, 0, 2, 1, 0, 0, true)
	layout.AddItem(editorPlaceholder, 0, 1, 2, 1, 0, 0, true)

	if err := refreshServices(serviceList, cwd); err != nil {
		panic(err)
	}

	messageSelector := tview.NewList()
	messageSelector.SetBorder(true)
	messageSelector.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			layout.RemoveItem(messageSelector)
			layout.AddItem(serviceList, 0, 0, 2, 1, 0, 0, true)
			app.SetFocus(serviceList)
		}

		return event
	})

	messageEditor := tview.NewForm()
	messageEditor.SetBorder(true)

	descriptors := map[string]*descriptorpb.DescriptorProto{}

	serviceList.SetSelectedFunc(func(index int, service string, _ string, _ rune) {
		// find the filename
		items, err := os.ReadDir(fmt.Sprintf("%s/%s/proto", cwd, service))
		if err != nil {
			panic(err)
		}

		file := ""
		for _, item := range items {
			if name := item.Name(); !item.IsDir() && strings.HasSuffix(name, ".proto") {
				file = name
				break
			}
		}
		if file == "" {
			return
		}
		// read the file
		p := protoparse.Parser{}

		d, err := p.ParseFilesButDoNotLink(fmt.Sprintf("%s/%s/proto/%s", cwd, service, file))
		if err != nil {
			return
		}

		messageSelector.Clear()
		messageSelector.SetTitle(fmt.Sprintf("Protobuf message: %s", service))
		descriptors = map[string]*descriptorpb.DescriptorProto{}
		for _, descriptor := range d {
			for _, mdesc := range descriptor.MessageType {
				// fully qualified message name
				fqmn := fmt.Sprintf("%s.%s", descriptor.GetPackage(), mdesc.GetName())
				descriptors[fqmn] = mdesc
				messageSelector.AddItem(fqmn, "", 0, nil)
			}
		}

		layout.RemoveItem(serviceList)
		layout.AddItem(messageSelector, 0, 0, 2, 1, 0, 0, true)
		app.SetFocus(messageSelector)
	})

	messageSelector.SetSelectedFunc(func(index int, fqmn string, _ string, _ rune) {
		messageEditor.Clear(true)
		messageEditor.SetTitle(fmt.Sprintf("Message: %s", fqmn))
		for _, field := range descriptors[fqmn].Field {
			fieldName := field.GetName()
			fieldType := strings.ToLower(strings.TrimPrefix(field.GetType().String(), "TYPE_"))
			fieldModifier := strings.ToLower(strings.TrimPrefix(field.GetLabel().String(), "LABEL_"))
			messageEditor.AddInputField(fmt.Sprintf("%s (%s %s)", fieldName, fieldModifier, fieldType), field.GetDefaultValue(), 0, nil, nil)
		}
		messageEditor.AddButton("Send", func() {
			root.AddText("Sending message!", true, tview.AlignLeft, tcell.ColorGreen)
		})
		layout.RemoveItem(editorPlaceholder)
		layout.AddItem(messageEditor, 0, 1, 2, 1, 0, 0, true)
		app.SetFocus(messageEditor)
	})

	app.SetRoot(root, true).Run()
}

func refreshServices(targetList *tview.List, sourceDir string) error {
	targetList.Clear()
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			targetList.AddItem(entry.Name(), "", 0, nil)
		}
	}

	return nil
}
