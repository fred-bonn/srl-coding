# String Formatting

In the last lesson, `hello_world` returned the same string on every call.
A real greeting greets *someone*.
The function needs to receive a name.

## Parameters

A function can take *parameters*: names written between the parentheses, which the body can use like any other variable.
The values we hand over in a call are called *arguments*.
When a call runs, each parameter is assigned the value of its argument:

```python
def greet(name):
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

Braces `{}` mark expressions: each expression is evaluated, and its value is inserted into the string at that spot:

The braces can hold any expression, not only a variable:

```python
a = 2
b = 3
print(f"{a} + {b} is {a + b}")
```

Output:

```
2 + 3 is 5
```
