

## protocol

In protocol, all attribute must be present in json layout:

Empty array: "arrayVar": []

Empty string: "stringVar": ""

No 'nil' in values. No ",omitempty"


## Classes 

In case we implements something which look like a class, layout should be:

In a folder 'classname'

### classname.go 

Contains class interface

```
type ClassName interface {
    Function1(...) (....)
    Function2(...) (....)
}
```

### internal.go

Contains class implementation

```

type className struct {
    internalVar type,
    ....
}


func (c *className) Function1(....) (...) {
    ....
}


func (c *className) Function2(....) (...) {
    ....
}

```

NB: If implementation is complex, it can be spread over several files. But there should be a file 'internal.go' which host the structure.


### builder.go

Contains constructor:

```
func New(.....) (ClassName, error) {
    className := &className{
        ....
    }
    ....
    return className, nil
}
```
# Simple classes

In the case of simple class, an alternative layout, everything in a single class.

Another simplification may be to remove the New() function and make the implementation structure public, to allow member initialization


