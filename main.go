package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	format "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gdamore/tcell/v2"
	proto2 "github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/rivo/tview"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	kafkaHost = flag.String("kafka-host", "localhost:9092", "Set the host to your kafka instance")
)

const serviceName = "stratus"

var logger *zap.SugaredLogger

var app = tview.NewApplication()
var layout = tview.NewGrid()
var root = tview.NewFrame(layout).
	AddText("Esc to navigate back", true, tview.AlignLeft, tcell.ColorLightGray)
var fileBrowser = tview.NewList()
var messageEditor = tview.NewForm()
var currentDir = ""
var descriptors = map[string]*desc.MessageDescriptor{}
var publish func(event string, message proto.Message) error

func init() {
	configureLogger()
	configureGridLayout()
	configureFileBrowser()
	configureMessageEditor()
	connectToKafka()
}

func configureLogger() {
	c := zap.NewDevelopmentConfig()
	//const logDir = "/var/log/stratus.log"
	f, err := os.CreateTemp("", "stratus-*.log")
	//_, err := os.Create(logDir)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	c.OutputPaths = []string{f.Name()}
	l, err := c.Build()
	if err != nil {
		panic(err)
	}
	logger = l.Sugar()
	root.AddText(fmt.Sprintf("Debug logs at %s", f.Name()), false, tview.AlignCenter, tcell.ColorLightGrey)
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

func connectToKafka() {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0

	logger.Debugw(
		"connecting to kafka",
		"broker", kafkaAddr(),
		"topic", kafkaTopic(),
	)
	sender, err := kafka_sarama.NewSender([]string{kafkaAddr()}, saramaConfig, kafkaTopic())
	if err != nil {
		logger.Fatalw("failed to create protocol",
			"error", err,
			"protocol", "kafka",
		)

		panic(err)
	}
	//sender.Close()

	c, err := cloudevents.NewClient(
		sender,
		cloudevents.WithTimeNow(),
		cloudevents.WithUUIDs(),
		cloudevents.WithEventDefaulter(func(ctx context.Context, event event.Event) event.Event {
			event.SetSource(serviceName)
			return event
		}),
	)

	if err != nil {
		logger.Fatalw("failed to create cloudevent client",
			"error", err,
		)

		panic(err)
	}

	root.AddText("✔ Connected to kafka", true, tview.AlignRight, tcell.ColorGreen)

	publish = func(event string, message proto.Message) error {
		data, err := anypb.New(message)

		if err != nil {
			logger.Errorw("failed to encode event", "error", err)

			return err
		}
		e, err := format.FromProto(&pb.CloudEvent{
			SpecVersion: cloudevents.VersionV1,
			Type:        event,
			Data: &pb.CloudEvent_ProtoData{
				ProtoData: data,
			},
		})

		if err != nil {
			logger.Errorw("failed to format event", "error", err)
			return err
		}

		return c.Send(context.Background(), *e)
	}
}

func kafkaAddr() string {
	if addr := os.Getenv("KAFKA_BROKER"); addr != "" {
		return addr
	}

	return "127.0.0.1:9092"
}

func kafkaTopic() string {
	if topic := os.Getenv("KAFKA_TOPIC"); topic != "" {
		return topic
	}

	return "events"
}

func main() {
	flag.Parse()
	cwd, _ := os.Getwd()

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
	descriptors = map[string]*desc.MessageDescriptor{}
	// if we're "inside" a proto file, we should render messages instead of directories
	if filepath.Ext(currentDir) == ".proto" {
		// read the file
		p := protoparse.Parser{}

		d, err := p.ParseFiles(currentDir)
		if err != nil {
			return err
		}
		fileBrowser.AddItem("..", "", 0, nil)

		for _, descriptor := range d {
			for _, mdesc := range descriptor.GetMessageTypes() {
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
	event := ""
	messageEditor.Clear(true)
	messageEditor.SetTitle(fmt.Sprintf("Message: %s", fqmn))
	messageEditor.AddInputField("Event", event, 0, nil, func(text string) {
		event = text
	})

	message := dynamic.NewMessage(descriptors[fqmn])

	for _, field := range descriptors[fqmn].GetFields() {
		fieldName := field.GetName()

		fieldType := strings.ToLower(strings.TrimPrefix(field.GetType().String(), "TYPE_"))
		fieldModifier := strings.ToLower(strings.TrimPrefix(field.GetLabel().String(), "LABEL_"))

		f := *field
		messageEditor.AddInputField(
			fmt.Sprintf("%s (%s %s)", fieldName, fieldModifier, fieldType), "",
			0,
			nil,
			func(text string) {
				switch field.GetType() {
				case descriptorpb.FieldDescriptorProto_TYPE_INT32:
					// parse as number
					n, _ := strconv.Atoi(text)
					message.SetField(&f, n)
				case descriptorpb.FieldDescriptorProto_TYPE_INT64:
					n, _ := strconv.Atoi(text)
					message.SetField(&f, n)
				case descriptorpb.FieldDescriptorProto_TYPE_STRING:
					message.SetField(&f, text)
				}
			},
		)
	}
	messageEditor.AddButton("Send", func() {
		err := publish(event, proto2.MessageV2(message))
		if err != nil {
			logger.Errorw("failed to publish event", "error", err)
		}
		// send message over kafka
		root.AddText("Sending message!", true, tview.AlignLeft, tcell.ColorGreen)
	})

	app.SetFocus(messageEditor)
}
