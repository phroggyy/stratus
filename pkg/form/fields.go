package form

import (
	"fmt"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/rivo/tview"
	"google.golang.org/protobuf/types/descriptorpb"
	"strconv"
	"strings"
)

func AddInputField(form *tview.Form, protoField *desc.FieldDescriptor, fieldNamePrefix string, output *dynamic.Message) {
	if protoField.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE && protoField.GetLabel() != descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		messageType := protoField.GetMessageType()
		message := dynamic.NewMessage(messageType)
		for _, field := range messageType.GetFields() {
			AddInputField(
				form,
				field,
				strings.TrimPrefix(fmt.Sprintf("%s.%s", fieldNamePrefix, protoField.GetName()), "."),
				message,
			)
		}

		output.SetField(protoField, message)
		return
	}

	f := *protoField
	form.AddInputField(
		getFieldName(fieldNamePrefix, protoField),
		//fmt.Sprintf("%s (%s %s)", fieldName, fieldModifier, fieldType),
		"",
		0,
		nil,
		func(text string) {
			switch protoField.GetType() {
			case descriptorpb.FieldDescriptorProto_TYPE_INT32:
				// parse as number
				n, _ := strconv.Atoi(text)
				output.SetField(&f, n)
			case descriptorpb.FieldDescriptorProto_TYPE_INT64:
				n, _ := strconv.Atoi(text)
				output.SetField(&f, n)
			case descriptorpb.FieldDescriptorProto_TYPE_STRING:
				output.SetField(&f, text)
			}
		},
	)
}

func getFieldName(fieldNamePrefix string, field *desc.FieldDescriptor) string {
	fieldName := field.GetName()

	fieldType := strings.ToLower(strings.TrimPrefix(field.GetType().String(), "TYPE_"))
	fieldModifier := strings.ToLower(strings.TrimPrefix(field.GetLabel().String(), "LABEL_"))

	return strings.TrimPrefix(
		fmt.Sprintf("%s.%s (%s %s)", fieldNamePrefix, fieldName, fieldModifier, fieldType),
		".",
	)
}
