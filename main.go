package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/nutmeglabs/banda/gen/idl/extensions"
)

// We need to keep a mapping of fully-qualified type name to descriptor so we can lookup whether a field is
// a map or not. We don't want to do code generation for map fields.
var typeNameToDesc map[string]*descriptor.DescriptorProto

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var req plugin.CodeGeneratorRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Fatalf("unable to parse protobuf: %v", err)
	}

	files := make([]*plugin.CodeGeneratorResponse_File, 0, len(req.ProtoFile))

	initTypeNameToDesc(req.ProtoFile)

	for _, f := range req.ProtoFile {
		// Exclude "behind-the-scenes" protos from code generation.
		if strings.Contains(*f.Name, "google/protobuf") || strings.Contains(*f.Name, "extensions.proto") {
			continue
		}

		code := bytes.NewBuffer(nil)

		fileHeaderTemplate.Execute(code, &File{
			Name:    *f.Name,
			Package: *f.Package,
		})

		for _, msg := range f.MessageType {
			emitMessageType(code, "", msg)
		}

		outputFilename := strings.TrimSuffix(*f.Name, ".proto") + "_trans.pb.go"
		files = append(files,
			&plugin.CodeGeneratorResponse_File{
				Name:    proto.String(outputFilename),
				Content: proto.String(strings.TrimLeft(code.String(), "\n")),
			})
	}

	emitFiles(files)
}

func emitMessageType(code io.Writer, msgNamePrefix string, msg *descriptor.DescriptorProto) error {
	// Maps are represented as 'repeated MapNameEntry map_name' fields in proto descriptors but they become
	// Golang maps in generated code. Because of this, we can skip the code generation for any of these intermediate
	// map entry message types and only do code generation for the repeated map entry field.
	if msg.GetOptions().GetMapEntry() == true {
		return nil
	}

	msgName := msgNamePrefix + *msg.Name

	for _, nestedType := range msg.NestedType {
		emitMessageType(code, msgName+"_", nestedType)
	}

	var translateFields []*Field
	var compositeFields []*Field
	for _, field := range msg.Field {
		ext, err := proto.GetExtension(field.GetOptions(), extensions.E_Translated)
		if err == nil {
			// Ignore "[(extensions.translated) = false]"
			if *ext.(*bool) != true {
				continue
			}

			if *field.Type != descriptor.FieldDescriptorProto_TYPE_STRING {
				log.Fatalf("error: 'translated' option used with non-string field (msg = %v, field = %v)", *msg.Name, *field.Name)
			}
			translateFields = append(translateFields, &Field{
				Name:    camelCase(*field.Name),
				IsArray: *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED,
			})
		}

		// Collect composite fields excluding Google well-known types. Well-known types are guaranteed not to contain translation
		// annotations.
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE && !isGoogleWKT(field) {
			desc := typeNameToDesc[*field.TypeName]
			isMap := desc != nil && desc.GetOptions().GetMapEntry()

			// Skip any maps that don't have composite value types since we need an enclosing message type in order to have a translated string field in a map.
			if isMap && desc.Field[1].GetType() != descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				continue
			}

			isArray := !isMap && *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED

			compositeFields = append(compositeFields, &Field{
				Name:    camelCase(*field.Name),
				IsArray: isArray,
				IsMap:   isMap,
			})
		}
	}

	m := &Message{
		Name:             msgName,
		TranslatedFields: translateFields,
		CompositeFields:  compositeFields,
	}

	extractTranslationsTemplate.Execute(code, m)
	translateTemplate.Execute(code, m)
	getTranslationKeysTemplate.Execute(code, m)

	return nil
}

// initTypeNameToDesc builds the map from fully qualified type names to descriptors.
// The key names for the map come from the input data, which puts a period at the beginning.
func initTypeNameToDesc(files []*descriptor.FileDescriptorProto) {
	typeNameToDesc = make(map[string]*descriptor.DescriptorProto)
	for _, f := range files {
		// The names in this loop are defined by the proto world, not us, so the
		// package name may be empty.  If so, the dotted package name of X will
		// be ".X"; otherwise it will be ".pkg.X".
		dottedPkg := "." + f.GetPackage()
		if dottedPkg != "." {
			dottedPkg += "."
		}

		for _, desc := range f.MessageType {
			addTypeNames(dottedPkg, desc)
		}
	}
}

func addTypeNames(prefix string, desc *descriptor.DescriptorProto) {
	name := prefix + *desc.Name
	typeNameToDesc[name] = desc

	for _, msg := range desc.NestedType {
		addTypeNames(name+".", msg)
	}
}

func emitFiles(out []*plugin.CodeGeneratorResponse_File) {
	buf, err := proto.Marshal(&plugin.CodeGeneratorResponse{File: out})
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stdout.Write(buf); err != nil {
		log.Fatal(err)
	}
}

func isGoogleWKT(field *descriptor.FieldDescriptorProto) bool {
	return strings.HasPrefix(*field.TypeName, ".google.protobuf")
}
