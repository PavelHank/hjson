# hjson
parse json string hommization for golang. You can fetch any value by key path. 

# How To Install?

```shell
> go get github.com/PavelHank/hjson
```

# Example

We can achieve any json node by key path, there are some example to explain how to use the library. Below example depend the facade json data:

```json
{
    "name": "Pavel Hank",
    "birthday": 1269003023,
    "favorite_sports": [
        "Basketball",
        "Football",
        "Badminton"
    ],
    "preferences":{
        "eat": [
            "rice",
            "noodles",
            "milk"
        ],
        "sports":{
             "Basketball":"crazy",
            "Football":"watch games some time",
            "Badminton":"beloved"
        }
    },
    "is_student": false,
    "is_programmer":true,
    "bio":"poor life, and poor soul, just live in this world like a wanderer, no one pays attention, no one knows",
    "deposit": 262.62
}
```

## GetString

```go
import "github.com/PavelHank/hjson"

func main(){
    // GetString by key.
    fmt.Println(hjson.GetString([]byte(json), "name"))
    // Output:
    // Pavel Hank <nil>

    // GetString by key path in deepper level.
    fmt.Println(hjson.GetString([]byte(json), "preferences.sports.Basketball"))
    // Output:
    // crazy <nil>

    // GetString from json array
    fmt.Println(hjson.GetString([]byte(json), "preferences.eat.$0"))
    // Output:
    // rice <nil>
}
```

## GetInt/GetFloat/GetBool
used it like `GetString` function.

## Bug Report
report any bug with new issues or hit my email: pavelhank@outlook.com