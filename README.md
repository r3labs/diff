# diff

## Synopsis

A library for diffing golang structures and values. It supports basic diffing, returning true or false if a change is detected, or a full changelog of all items that have been modified.

## Build status

* Master [![CircleCI](https://circleci.com/gh/r3labs/diff/tree/master.svg?style=svg)](https://circleci.com/gh/r3labs/diff/tree/master)

## Installation

```
go get github.com/r3labs/diff
```

## Usage

### Basic Example

Diffing a basic set of values can be accomplished using the diff functions. Any items that specify a "diff" tag using a name will be compared.

```go
import "github.com/r3labs/diff"

type Order struct {
    ID    string `diff:"id"`
    Items []int  `diff:"items"`
}

func main() {
    a := Order{
        ID: "1234",
        Items: []int{1, 2, 3, 4},
    }

    b := Order{
        ID: "1234",
        Items: []int{1, 2, 4},
    }

    changelog, err := diff.Diff(a, b)
    ...
}
```

In this example, the output generated in the changelog will indicate that the third element with a value of '3' was removed from items.
When marshalling the changelog to json, the output will look like:

```json
[
    {
        "type": "delete",
        "path": [
            "items", "2"
        ],
        "from": 3,
        "to": null
    }
]
```

A more complicated example might look like:

```go
import "github.com/r3labs/diff"


type Tag struct {
    Name  string `diff:"name,identifier"`
    Value string `diff:"value"`
}

type Fruit struct {
    ID        int      `diff:"id"`
    Name      string   `diff:"name"`
    Healthy   bool     `diff:"healthy"`
    Nutrients []string `diff:"nutrients"`
    Tags      []Tag    `diff:"tags"`
}

func main() {
    a := Fruit{
        ID:      1,
        Name:    "Green Apple",
        Healthy: true,
        Nutrients: []string{
            "vitamin c",
            "vitamin d",
        },
        Tags: []Tag{
            {
                Name:  "kind",
                Value: "fruit",
            },
        },
    }

    b := Fruit{
        ID:      2,
        Name:    "Red Apple",
        Healthy: true,
        Nutrients: []string{
            "vitamin c",
            "vitamin d",
            "vitamin e",
        },
        Tags: []Tag{
            {
                Name:  "popularity",
                Value: "high",
            },
            {
                Name:  "kind",
                Value: "fruit",
            },
        },
    }

    changelog, err := diff.Diff(a, b)
    ...
}
```

This yields a changelog of:

```json
[
    {
        "type": "update",
        "path": [
            "id"
        ],
        "from": 1,
        "to": 2
    },
    {
        "type": "update",
        "path": [
            "name"
        ],
        "from": "Green Apple",
        "to": "Red Apple"
    },
    {
        "type": "create",
        "path": [
            "nutrients",
            "2"
        ],
        "from": null,
        "to": "vitamin e"
    },
    {
        "type": "create",
        "path": [
            "tags",
            "popularity"
        ],
        "from": null,
        "to": {
            "Name": "popularity",
            "Value": "high"
        }
    }
]
```

### Tag Options

The following tag options can be set:

* `-` excludes the field from the diff
* `identifier` indicates that this field is used as an identifier when the parent struct is a member of a slice. In a slice where there are identifiable items, ordering will be ignored and matched components will be diff'ed together.


## Supported Types

A diffable value can be/contain any of the following types:

* struct
* slice
* string
* int
* bool
* map
* pointer


## Running Tests

```
make test
```

## Contributing

Please read through our
[contributing guidelines](CONTRIBUTING.md).
Included are directions for opening issues, coding standards, and notes on
development.

Moreover, if your pull request contains patches or features, you must include
relevant unit tests.

## Versioning

For transparency into our release cycle and in striving to maintain backward
compatibility, this project is maintained under [the Semantic Versioning guidelines](http://semver.org/).

## Copyright and License

Code and documentation copyright since 2015 r3labs.io authors.

Code released under
[the Mozilla Public License Version 2.0](LICENSE).
