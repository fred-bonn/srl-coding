# Functions

A function is a named block of code that we can run whenever we like.
For example, `print` is a built-in function available to us in Python.
Now we get to write our own functions.

## Example

We define a function with the `def` keyword, a name, parentheses, and a colon.
The code we want the function to run is indented underneath, and is called the function's *body*:

```python
def greet():
    print("Greetings Humans")


greet()
```

Output:

```
Greetings Humans
```

Defining a function does not run its body.
Nothing happens until we *call* the function by writing its name followed by parentheses.
Each call runs the body from top to bottom, and we can call a function as many times as we like:

```python
def greet():
    print("Greetings Humans")


greet()
greet()
```

Output:

```
Greetings Humans
Greetings Humans
```

## Return Values

So far our functions have only printed.
More often, we want a call to *evaluate to a value*, so we can store it in a variable or use it inside a larger expression.
The `return` statement ends the function and makes the call evaluate to the value written after it:

```python
def hello_world():
    return "Hello, World!"


message = hello_world()
print(message)
```

Output:

```
Hello, World!
```

The call `hello_world()` evaluates to the string `"Hello, World!"`, and the assignment stores that value in `message`.

A function that does not return a value will default to returning the `None` value:

```python
def greet():
    print("Greetings Humans")


result = greet()
print(result)
```

Output:

```
Greetings Humans
None
```

## Assignments

In this lesson we introduce a new assignment format.
We are now asking you to implement a single function.
The starter code already has the function you need to implement as a *stub*, which means that it returns `None`.
A Python function that doesn't return any value will default to returning `None`.
The tests import your code and call your function directly, checking the value it evaluates to.
When an assignment is about printing, they check the printed output instead.
