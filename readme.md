# Brainfuck Interpreter

This is a brainfuck interpreter that is built with a lexer, a parser and a debugger protocol. 

> Important: Currently there isn't a debugger client. If you run your program with a debugger attached, the program with wait for a client to connect and do nothing else.

## Running a script

To run a brainfuck script, import the runtime and create a runtime options struct and pass it to a new runtime.

```go
package main

import (
	"github.com/CanPacis/brainfuck-interpreter/runtime"
)

func main() {
  options := runtime.RuntimeOptions{
		FilePath: "./test/add.bf",
	}

  r := runtime.NewRuntime(options)
  // then simply run it
  r.Run()
}
```

## Superset
This is actually intended to be a superset of brainfuck so there are extended capabilities of the runtime.

### Clearing (`*` operator)
You can clear your tape with a `*` opearator.

```
+
>+
>+

tape is [1, 1, 1, 0, 0 ...] right now

*

tape is cleared and [0, 0, 0, 0, 0 ...] now
```

### Pushing Directly (`|` operator)

Although it seems out of nature with the philosophy of brainfuck, you can set value of a cell directly with a `|` operator. This is provided for ease of use because the main concern of this superset is extended io.

```
|255
>|128

tape is [255, 128, 0, 0, 0 ...] now
```

### Escaping
You need to escape the operator keywords if you are using them in your comments. The parser cannot ignore them without proper escaping.

```
Bad:
Calculate 7 * 8

Good:
Calculate 7 \* 8
```

### More than a byte
The interpreters tape can hold values more than a byte. Well not exactly, to be backwards compatible with brainfuck, the tape still holds `uint8` values. But values pushed with the `|` operator can hold values of `uint32`. These pushed values still be recorded as a `uint8` value in the tape but there is another tape that holds the `uint32` values. This secondary tape is always in sync with the primary tape so if you manipulate your `uint8` values, the `uint32` correspondant will also change. 

> Any other operator apart from the `|` operator will result in a `uint8` value in the `uint32` tape. For instance:

```
|256

primary tape is [1, 0, 0, 0, ...]
secondary tape is [256, 0, 0, 0, ...]

running the '\+' operator won't make the secondary cell 257, instead it will also be 2

+

primary tape is [2, 0, 0, 0, ...]
secondary tape is [2, 0, 0, 0, ...]
```

### IO
The plan is to make brainfuck be able to read and write to more than one io target that is std. It should be able to read and write to disk, tcp or http connections or any other byte writable stream. This part is still an ongoing process.