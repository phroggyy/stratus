package protobuf

import (
	"fmt"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetBufDependencies(bufPath string) ([]string, error) {
	bufFile, err := os.Open(bufPath)
	if err != nil {
		panic(err)
	}
	d := yaml.NewDecoder(bufFile)

	bufContents := map[string]interface{}{}
	if err := d.Decode(&bufContents); err != nil {
		return nil, err
	}

	if deps, ok := bufContents["deps"].([]string); ok {
		return deps, nil
	}

	return nil, nil
}

func DownloadBufDependencies(bufPath, outputDir string) error {
	deps, err := GetBufDependencies(bufPath)

	if err != nil {
		return err
	}

	for _, dep := range deps {
		err := exec.Command("buf", "export", dep, "-o", outputDir).Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetMessagesFromFile parses a protobuf file and returns a map of Fully Qualified Message Names (FQMNs)
// to the messages, allowing for proper import parsing.
func GetMessagesFromFile(protoFilepath, vendorPath, relativePathRoot string) (map[string]*desc.MessageDescriptor, error) {
	p := protoparse.Parser{
		Accessor: func(filename string) (io.ReadCloser, error) {
			if strings.HasPrefix(filename, "/") {
				return os.Open(filename)
			}

			if vendorPath != "" {
				if f, err := os.Open(filepath.Join(vendorPath, filename)); err == nil {
					return f, nil
				}
			}

			return os.Open(filepath.Join(relativePathRoot, filename))
		},
	}

	d, err := p.ParseFiles(protoFilepath)
	if err != nil {
		return nil, err
	}

	messages := map[string]*desc.MessageDescriptor{}

	for _, descriptor := range d {
		for _, mdesc := range descriptor.GetMessageTypes() {
			// fully qualified message name
			fqmn := fmt.Sprintf("%s.%s", descriptor.GetPackage(), mdesc.GetName())
			messages[fqmn] = mdesc
		}
	}

	return messages, nil
}
