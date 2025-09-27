# Neu language
Neu pronounced as in "Neumann" is a programming language for teaching beginners, who speak Hungarian.

It is designed to be a replacement for a learning programming language used in my university.

[Roadmap](https://trello.com/b/1dLB446k/neu-language-roadmap)

> [!IMPORTANT]
> I am fairly new to go and language design. Code quality may not be up to your standards.
> My nomenclature and notation may be unconventional, but it is consistent.

## Using Neu

Windows:
```
neu.exe <neu-program>
```

Linux/Mac
```console
./neu <neu-program>
```

Or if you have go installed:
```console
go run neu.go <neu-program>
```

## Building Neu
```console
go build -o build/neu neu.go
```
No dependencies!

## Hello world in Neu
```
PROG: hello
    KI: "Hello World"
:PROG
```

## Motivation

The original is designed as a Pascal like language with only Hungarian keywords, written in Java.

It also has some serious disadvantages in both design and in its pedagogical use. For example there are no for-loops, no functions, no file handling, and all variables have to be declared in a separate block. Since its syntax is Hungarian, there are special characters in keywords which in certain cases can get distorted if not saved with the appropriate encoding. In my opinion it "harms" some student's future of understanding programming, it is only available as a GUI program for editing/compiling/running your programs, which is bad because most students will confuse the editor with the programming language itself. (I noticed this while teaching, because every student hated c++ beacause of CodeBlocks)

Neu is designed to overcome these limitations, and prepare for a more smoother switch to later languages. Also I would like to introduce some concepts which will make using the language more intuitive.

## Syntax details

Most of the keywords are in Hungarian, but I decided to use some common English words for some of them. If you speak Hungarian, some of the keywords can sound weird, but that is done to circumvent the use of special characters (there will be no `KÜLÖNBEN` and `AMÍG`). Common keywords will receive shorter names than less common ones, to make typing it faster.

> [!CAUTION]
> Some of these listed are not implemented yet. Please refer to the roadmap to which version will they be added.

### Comments

Comments are lines marked with a `#` (hash) sign. There are no multi line comments.

Example:
```
# This is a comment
```

### Blocks

Blocks are declared with opening and closing tags. Like `TAG:` and `:TAG`, in special cases you may have `:TAG:` where it both closes the previous block and opens a new one.

Tags can also have certain parameters or variables next to the opening tag. With the `PROG` example, it takes an identifier as the name of the program. This is also how loop variables can be declared.

List of all the block tags are:
* `PROG` - main function
* `FN` - function definition (currently not implemented)
* `ISM` - for loop
* `CIKLUS` - while loop
* `HA` - if statement

### Basic I/O commands

These commands ("one line statements") are used for I/O using the command line and the file system.

List of all the commands are:
* `KI:` - print to standard out
* `BE:` - read from standard in

Usage:
```
BE: foo
```

Multiple things can be printed or read in using a `,` (comma) to separate them.

Usage:
```
KI: "My number is:", 10
```

### Variables

Variables are declared with a type a name and a value. If no value is given then the value will be a special `SEMMI` (null) value.

Variable names must start with a letter but then can include any alphanumeral character as well as underscores and questionmarks.

List of all the types are:
* `NUM` - number, can be integer or floating point
* `TXT` - text or strings
* `LOG` - logical, booleans (not to be confused with `LOG10` or `LOGN` for logarithms (currently not implemented))
* `KAR` - character
* `LIST` - array or list (currently not implemented)
* `FILE` - file variable, used for reading and writing to files (currently not implemented)
* `NUM-E` - number, must be integer (currently not implemented)
* `NUM-T` - number, must be floating point (currently not implemented)

Example:
```
NUM: age = 18
TXT: message = "Hello World"
```
Variables are always scoped to their blocks where they were declared in.

### PROG tag

Contains the main body of the program ("main function"). It must be followed by an identifier, usually the name of the file without its extension.

Example:
```
# foo.neu
PROG: foo

:PROG
```

### HA tag

This tag will run a block of code based on the logical value of the provided parameter ("if statement").

```
HA: 0<1
  KI: "true"
:HA
```
If we want to create another branch ("else") then the `:NEM-HA:` "double tag" can be used.

```
# LOG: foo

HA: foo
  KI: "true"
:NEM-HA:
  KI: "false"
:HA
```
To check multiple conditions ("else if") the `:DE-HA:` double tag can be used.
```
# NUM: foo

HA: foo>0
  KI: "positive"
:DE-HA: foo<0
  KI: "negative"
:NEM-HA:
  KI: "zero"
:HA
```

### ISM tag

This is used for repeating a given set of instructions a set amount of times ("for loop"). This tag must be followed by an integer value as the amount of repetitions. An optional indexing variable can also be declared after this tag. This variable will start at 0 and increment with every repetition.

Example:
```
# Simple repetition
ISM: 4
  KI: "this will appear four times"
:ISM

# Indexing variable
ISM: 10, idx
  KI: "this has appeared ", idx, " times previously"
:ISM
```

### CIKLUS tag

This is used for repeating a set of instructions until a condition is still true ("while loop"). This tag must be followed by a logical value as the loop condition. An optional indexing variable can also be declared, similar to the `ISM` tag.

Example:
```
NUM: foo = 0

CIKLUS: foo<10
  KI: "This will print many times"
  foo = foo+1
:CIKLUS

NUM: bar = 0

CIKLUS: bar<10, idx
  KI: "This has appeared", idx, " times previously"
  bar = bar+1
:CIKLUS
```

### File handling

**TODO (version 1)**

Files have to be declared with a valid path as string.
```
FILE: myfile = "foo.txt"
```
The first operation to a file will automatically open it, but closing it have to be done manually. A file must exist to be read from, but will be created if printed to. If a file already exists and it is printed to, then all contents will be overwritten. If the end of a file is reached then all subsequent `BE:` commands will result in a `SEMMI` value. Closing a file is done by printing a `SEMMI` value to it.

Reading is done line-by-line, and the result is always a `TXT` type.

Simple example:
```
# Create file variable
FILE: myfile = "foo.txt"

# Read in a single line and print it
TXT: line
myfile BE: line
KI: line

# Close the file
myfile KI: SEMMI
```
A more advanced example:
```
FILE: myfile = "foo.txt"
TXT: line

myfile BE: line
CIKLUS: line != SEMMI, idx
  KI: "line num: ", idx
  KI: line
  myfile BE: line
:CIKLUS
myfile KI: SEMMI
```

### Lists

**TODO (version 1)**

Lists are defined using the `LIST` type and in square brackets the type they will contain. Lists are static in size and this has to be indicated in the definition. Multidimensional arrays are possible by creating a list of lists (of lists...). Default values are left uninitialized for each element (see: `SEMMI`)

Examples:
```
# List of ten names
LIST[TXT]: names[10]

# 2x2 matrix
LIST[LIST[NUM]]: matrix[2][2]
```

Accessing an element of a list is done using its identifier and square brackets. Lists are zero indexed

```
KI: "First name in list:", names[0]
```

List definition also allows the usage of variables/expressions to determine the size of the list

```
NUM: dimension = 0
BE: dimension
LIST[NUM]: vector[dimension]

ISM: dimension, idx
  KI: idx, vector[idx]
:ISM
```

### Functions

**TODO (version 2)**

Functions are declared using the `FN` tag. Followed by their name, function parameters in parenthesis separated with a comma, an equals sign and the return type. Implicit returns can also be done if a variable is declared after the equals sign. Functions with no returns (void functions, procedures) can omit the equals sign and return, or can return a `SEMMI`.

Examples:
```
# Mutliply a number with 2
FN: mult2(NUM: x)=NUM
  VISSZA x*2
:FN

# Add to numbers together, and return implicitly
FN: sum_numbers(NUM:x,NUM:y)=NUM:sum
  sum = x+y
:FN

# No return function
FN: print_person(TXT:name,NUM:age)
  KI: "Person name: ", name
  KI: "Person age: ", age
  HA: age>18
    KI: "Can legally drink"
  :HA
:FN
```
