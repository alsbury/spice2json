# spice2json
Utility to generate a basic JSON representation of a SpiceDB Schema

## Build Binary

```
go build
```

---

## Command Usage

```
spice2json <input file> <output file>
```

## Example

This is a simple example of SpiceDB Schema DSL as input
```
definition user {}

definition platform {
	relation administrator: user

	permission super_admin = administrator

	permission create_tenant = administrator
}
```

JSON output from above example
```
{
  "definitions": [
    {
      "name": "user",
      "namespace": "default",
      "relations": [],
      "permissions": []
    },
    {
      "name": "platform",
      "namespace": "default",
      "relations": [
        {
          "name": "administrator"
        }
      ],
      "permissions": [
        {
          "name": "super_admin"
        },
        {
          "name": "create_tenant"
        }
      ]
    }
  ]
}
```
