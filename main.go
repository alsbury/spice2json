package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	ns "github.com/authzed/spicedb/pkg/namespace"
	corev1 "github.com/authzed/spicedb/pkg/proto/core/v1"
	iv1 "github.com/authzed/spicedb/pkg/proto/impl/v1"
	"github.com/authzed/spicedb/pkg/schemadsl/compiler"
	"github.com/authzed/spicedb/pkg/schemadsl/input"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	if len(os.Args) < 3 {
		displayUsageInfo()
		return
	}
	inputFileName := os.Args[1]          //os.Args[2]
	outputFileName := os.Args[2]         //os.Args[2]
	b, err := os.ReadFile(inputFileName) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	schemaSource := string(b) // convert content to a 'string'

	in := compiler.InputSchema{
		Source:       input.Source(inputFileName),
		SchemaString: schemaSource,
	}

	namespace := "default"

	def, _ := compiler.Compile(in, &namespace)
	var buf strings.Builder
	WriteSchemaTo(def.ObjectDefinitions, &buf)
	output, _ := PrettyString(buf.String())
	data := []byte(output)
	os.WriteFile(outputFileName, data, 0644)
}

func displayUsageInfo() {
	fmt.Println("")
	fmt.Println("Please provide a valid input schema and a path to the output json")
	fmt.Println("")
	fmt.Println("Example: spice2json myschema.zed myschema.json")
	fmt.Println("")
}

// https://gosamples.dev/pretty-print-json/
func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "  "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

/**
 * Portions of this code were pulled from https://github.com/oviva-ag/spicedb
 */
func WriteSchemaTo(definition []*corev1.NamespaceDefinition, w io.Writer) error {
	var objects []*Object
	for _, def := range definition {
		o, err := mapDefinition(def)
		if err != nil {
			return fmt.Errorf("failed to export %q: %w", def.Name, err)
		}
		objects = append(objects, o)
	}

	data, err := json.Marshal(map[string][]*Object{"definitions": objects})
	if err != nil {
		return fmt.Errorf("unable to serialize schema for export: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("unable to write schema for export: %w", err)
	}
	return nil
}

func mapDefinition(def *corev1.NamespaceDefinition) (*Object, error) {
	relations := []*Relation{}
	permissions := []*Permission{}
	for _, r := range def.Relation {
		kind := ns.GetRelationKind(r)
		if kind == iv1.RelationMetadata_PERMISSION {
			permissions = append(permissions, mapPermission(r))
		} else if kind == iv1.RelationMetadata_RELATION {
			relations = append(relations, mapRelation(r))
		} else {
			return nil, fmt.Errorf("unexpected relation %q, neither permission nor relation", r.Name)
		}
	}

	splits := strings.SplitN(def.Name, "/", 2)
	if len(splits) != 2 {
		return nil, fmt.Errorf("namespace missing for %q", def.Name)
	}
	namespace := splits[0]
	name := splits[1]

	return &Object{
		Name:        name,
		Namespace:   namespace,
		Relations:   relations,
		Permissions: permissions,
	}, nil
}

func mapRelation(relation *corev1.Relation) *Relation {
	return &Relation{Name: relation.Name}
}

func mapPermission(relation *corev1.Relation) *Permission {
	return &Permission{
		Name: relation.Name,
	}
}

type Relation struct {
	Name string `json:"name"`
}

type Permission struct {
	Name string `json:"name"`
}

type Object struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Relations   []*Relation   `json:"relations"`
	Permissions []*Permission `json:"permissions"`
}

type Schema struct {
	Objects []*Object `json:"objects"`
}
