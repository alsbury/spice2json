package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/authzed/spicedb/pkg/schemadsl/compiler"
)

const VERSION = "0.3.0"

func main() {
	namespace := flag.String("n", "", "default namespace")
	version := flag.Bool("v", false, "print version and exit")
	stdIn := flag.Bool("s", false, "read schema from stdin rather than a file")
	readFile := flag.Bool("f", false, "read schema from file (default)")
	readRest := flag.Bool("h", false, "read from spicedb http url to retrieve schema")
	readGrpc := flag.Bool("g", false, "read from spicedb grpc host + port to retrieve schema")
	insecureGrpc := flag.Bool("insecure", false, "connect to non TLS grpc host")
	key := flag.String("k", "", "pre-shared key for rest / grpc schema")
	flag.Parse()

	if *version == true {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	var schema string
	if *stdIn {
		stdin, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		schema = string(stdin)
	} else {
		inputSrc := flag.Arg(0)
		if inputSrc == "" {
			displayUsageInfo()
			os.Exit(1)
		}

		if !*readGrpc && !*readRest {
			*readFile = true
		}

		if *readFile {
			schema = readSchemaFromFile(inputSrc)
		} else if *readRest {
			schema = readSchemaFromUrl(inputSrc, *key)
		} else if *readGrpc {
			schema = readSchemaFromGrpc(inputSrc, *key, *insecureGrpc)
		}
	}

	in := compiler.InputSchema{
		SchemaString: schema,
	}

	def, err := compiler.Compile(in, namespace)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var buf strings.Builder
	err = WriteSchemaTo(def, &buf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	output, _ := PrettyString(buf.String())

	outputFileName := flag.Arg(1)
	if outputFileName != "" {
		data := []byte(output)
		err = os.WriteFile(outputFileName, data, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		fmt.Print(output)
	}
}

func displayUsageInfo() {
	fmt.Println("Spice2JSON " + VERSION)
	fmt.Println("Please provide a valid input schema and a path to the output json")
	fmt.Println("")
	fmt.Println("Read from file: spice2json test_schema.zaml [output.json]")
	fmt.Println("Read from stdin: spice2json -s")
	fmt.Println("Read from spicedb rest client: spice2json -h http://localhost:8443")
	fmt.Println("Read from spicedb grpc client: spice2json -g [-insecure] localhost:50051")
	flag.Usage()
}

// PrettyString https://gosamples.dev/pretty-print-json/
func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "  "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

// WriteSchemaTo Portions of this code were pulled from https://github.com/oviva-ag/spicedb
func WriteSchemaTo(schema *compiler.CompiledSchema, w io.Writer) error {
	var definitions []*Definition
	for _, def := range schema.ObjectDefinitions {
		o, err := mapDefinition(def)
		if err != nil {
			return fmt.Errorf("failed to export %q: %w", def.Name, err)
		}
		definitions = append(definitions, o)
	}

	var caveats []*Caveat
	for _, caveat := range schema.CaveatDefinitions {
		o := mapCaveat(caveat)
		caveats = append(caveats, o)
	}

	data, err := json.Marshal(&Schema{
		Definitions: definitions,
		Caveats:     caveats,
	})
	if err != nil {
		return fmt.Errorf("unable to serialize schema for export: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("unable to write schema for export: %w", err)
	}
	return nil
}
