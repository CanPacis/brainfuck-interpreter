# Brainfuck Interpreter

This is a brainfuck interpreter that is built with a lexer, a parser and a debugger protocol.

> Currently there isn't a debugger client. If you run your program with a debugger attached, the program will wait for a client to connect and do nothing else.

## Running a script

To run a brainfuck script, import the engine and create an engine options struct and pass it to a new engine.

```go
package main

import (
  "github.com/CanPacis/brainfuck-interpreter/engine"
)

func main() {
  options := engine.EngineOptions{
    FilePath: "./bf/add.bf",
  }

  e := engine.NewEngine(options)
  // then simply run it
  e.Run()
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

### Escaping

You need to escape the operator keywords if you are using them in your comments. The parser cannot ignore them without proper escaping.

```
Bad:
Calculate 7 * 8

Good:
Calculate 7 \* 8
```

### IO

The plan is to make brainfuck be able to read and write to more than one io target that is std. It should be able to read and write to disk, tcp or http connections or any other byte writable stream. This part is still an ongoing process.
