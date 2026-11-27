# String Formatting

In the last lesson, `hello_world` returned the same string on every call.
A real greeting greets *someone*.

## Parameters

A function can take *parameters*.
A parameter is a name written between the parentheses, which the body can use like any other variable.
The values we hand over in a call are called *arguments*.
When a call runs, each parameter is assigned the value of its argument:

```python
def greet(name: str) -> None:
    print("Greetings", name)


greet("Humans")
greet("Ada")
```

Output:

```
Greetings Humans
Greetings Ada
```

## F-Strings

An *f-string* is a string literal prefixed with the letter `f`.

```python
a = "Ada"
b = "Bob"
print(f"Hello, {a}!")
print(f"Hello, {b}!")
```

Output:

```
Hello, Ada!
Hello, Bob!
```

Braces `{}` mark expressions: each expression is evaluated, and its value is inserted into the string at that spot.

The braces can hold any expression, not only a variable:

```python
a: int = 2
b: int = 3
print(f"{a} + {b} is {a + b}")
```

Output:

```
2 + 3 is 5
```

## Type Hints

In the `greet` definition we added `name: str` and `-> None`.
These are *type hints*.
A type hint documents the type a variable or value will have.
For example, `name: str` indicates that the `name` parameter is a string, and `-> None` indicates that the function will evaluate to the `None` value when called.

Type hints are not mandatory in Python, as it is a [dynamically typed language](https://en.wikipedia.org/wiki/Dynamic_programming_language).
Nonetheless, they are useful as documentation for yourself and whoever is reading your code.
From this lesson on, the code we provide will *always* use type hints.
