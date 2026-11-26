# Functions

A function is a named block of code that we can run whenever we like.
We have been calling functions all along: `print`, `int`, and `float` are all functions someone else wrote for us.
Now we write our own.

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
def hello_world() -> str:
    return "Hello, World!"


message = hello_world()
print(message)
```

Output:

```
Hello, World!
```

The call `hello_world()` evaluates to the string `"Hello, World!"`, and the assignment stores that value in `message`.
Nothing is printed inside the function: it only produces the value, and we decide where and when to print it.

The `-> str` after the parentheses is a *type annotation*: a promise to the reader that this function evaluates to a string.
Python does not check the promise; it is documentation, like a comment.

A function that never runs a `return` statement still evaluates to a value.
It evaluates to `None`, the "no value" value we already know:

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

From this lesson on, we no longer write whole programs.
Assignments ask us to implement a function, and the starter code already has its name and a `pass` body to replace.
The tests import our code and call our function directly, checking the value it evaluates to.
When an assignment is about printing, they check the printed output instead.
