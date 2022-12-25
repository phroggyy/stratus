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

func addRepeatedInput(form *tview.Form, field *desc.FieldDescriptor, fieldNamePrefix string, output *dynamic.Message) {
	vals := make([]interface{}, 0)

	prefixedName := getPrefixedName(fieldNamePrefix, field.GetName())
	fieldName := fmt.Sprintf("%s[%d]", prefixedName, len(vals))
	form.AddButton(
		fmt.Sprintf("Add a %s", getPrefixedName(fieldNamePrefix, field.GetName())),
		func() {
			switch field.GetType() {
			case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
				message := addMessageField(form, fieldName, field)
				vals = append(vals, message)
				output.SetField(field, vals)
			case descriptorpb.FieldDescriptorProto_TYPE_INT32:
			case descriptorpb.FieldDescriptorProto_TYPE_INT64:
				val := 0
				// we store references so that we can update the same array item
				vals = append(vals, &val)
				addNumField(form, fieldName, field, func(n int) {
					val = n
					outVals := make([]int, 0)
					for _, valRef := range vals {
						outVals = append(outVals, *valRef.(*int))
					}

					output.SetField(field, outVals)
				})
			case descriptorpb.FieldDescriptorProto_TYPE_STRING:
				val := ""
				vals = append(vals, &val)
				addTextField(form, fieldName, field, func(s string) {
					val = s
					outVals := make([]string, 0)
					for _, valRef := range vals {
						outVals = append(outVals, *valRef.(*string))
					}
					output.SetField(field, outVals)
				})
			}
		},
	)
}

func addMessageField(form *tview.Form, fieldName string, messageField *desc.FieldDescriptor) *dynamic.Message {
	messageType := messageField.GetMessageType()
	message := dynamic.NewMessage(messageType)
	for _, field := range messageType.GetFields() {
		AddInputField(
			form,
			field,
			fieldName,
			message,
		)
	}

	return message
}

func addTextField(form *tview.Form, fieldNamePrefix string, field *desc.FieldDescriptor, changed func(string)) {
	form.AddInputField(
		getFieldName(getPrefixedName(fieldNamePrefix, field.GetName()), field),
		"",
		0,
		nil,
		changed,
	)
}

func addNumField(form *tview.Form, fieldNamePrefix string, field *desc.FieldDescriptor, changed func(int)) {
	form.AddInputField(
		getFieldName(getPrefixedName(fieldNamePrefix, field.GetName()), field),
		"",
		0,
		nil,
		func(text string) {
			n, _ := strconv.Atoi(text)
			changed(n)
		},
	)
}

func AddInputField(form *tview.Form, protoField *desc.FieldDescriptor, fieldNamePrefix string, output *dynamic.Message) {
	if protoField.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		addRepeatedInput(form, protoField, fieldNamePrefix, output)
		return
	}

	f := *protoField

	switch protoField.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		message := addMessageField(form, getPrefixedName(fieldNamePrefix, protoField.GetName()), protoField)
		output.SetField(protoField, message)
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		addNumField(form, fieldNamePrefix, protoField, func(n int) {
			output.SetField(&f, n)
		})
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		addTextField(form, fieldNamePrefix, protoField, func(s string) {
			output.SetField(&f, s)
		})
	}

	//form.AddInputField(
	//	getFieldName(fieldNamePrefix, protoField),
	//	"",
	//	0,
	//	nil,
	//	func(text string) {
	//		switch protoField.GetType() {
	//		case descriptorpb.FieldDescriptorProto_TYPE_INT32:
	//			// parse as number
	//			n, _ := strconv.Atoi(text)
	//			output.SetField(&f, n)
	//		case descriptorpb.FieldDescriptorProto_TYPE_INT64:
	//			n, _ := strconv.Atoi(text)
	//			output.SetField(&f, n)
	//		case descriptorpb.FieldDescriptorProto_TYPE_STRING:
	//			output.SetField(&f, text)
	//		}
	//	},
	//)
}

func getPrefixedName(fieldNamePrefix, name string) string {
	return strings.TrimPrefix(fmt.Sprintf("%s.%s", fieldNamePrefix, name), ".")
}

func getFieldName(fieldName string, field *desc.FieldDescriptor) string {
	fieldType := strings.ToLower(strings.TrimPrefix(field.GetType().String(), "TYPE_"))
	fieldModifier := strings.ToLower(strings.TrimPrefix(field.GetLabel().String(), "LABEL_"))

	return strings.TrimPrefix(
		fmt.Sprintf("%s (%s %s)", fieldName, fieldModifier, fieldType),
		".",
	)
}
