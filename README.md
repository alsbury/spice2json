# spice2json
Utility to generate a simplified JSON representation of a SpiceDB Schema in order to power code
generation in other languages.

## Build Binary

Build for mac

```shell
GOARCH=arm64 go build -ldflags="-s -w"
```

Build for intel

```shell
GOARCH=amd64 go build -ldflags="-s -w"
```

Compress using [upx](https://upx.github.io/) for a smaller build

```
upx --brute spice2json
```

---

## Command Usage

Read from file, output is to stdout unless an output file is given as the second argument.
```shell
spice2json [-n namespace] input.zaml [output.json]
```

Read from stdin
```shell
spice2json -s < schema.zaml
```

Read from spicedb rest client
```shell
spice2json -h -k MyPreSharedKey http://localhost:8443
```

Read from spicedb grpc client
```shell
spice2json -g -k MyPreSharedKey [-insecure] localhost:50051
```


## Example

This is a simple example of SpiceDB Schema DSL as input
```
/** 
 * represents a user of the system 
 */
definition user {}

definition platform {
	relation administrator: user

	permission super_admin = administrator

	permission create_tenant = super_admin + administrator
}
```

JSON output from above example
```
{
  "definitions": [
    {
      "name": "user",
      "comment": "represents a user of the system"
    },
    {
      "name": "platform",
      "relations": [
        {
          "name": "administrator",
          "types": [
            {
              "type": "user"
            }
          ]
        }
      ],
      "permissions": [
        {
          "name": "super_admin",
          "userSet": {
            "operation": "union",
            "children": [
              {
                "relation": "administrator"
              }
            ]
          }
        },
        {
          "name": "create_tenant",
          "userSet": {
            "operation": "union",
            "children": [
              {
                "relation": "super_admin"
              },
              {
                "relation": "administrator"
              }
            ]
          }
        }
      ]
    }
  ]
}
```
