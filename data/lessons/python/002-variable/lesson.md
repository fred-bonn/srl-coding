# Variables

A variable is a named place to store a value, so we can refer to it later by name.

## Example

We create a variable with the *assignment operator* `=`:

```python
greeting = "Greetings Humans"
print(greeting)
```

Output:

```
Greetings Humans
```

In this example, the variable `greeting` is on the left-hand side of `=`, and the string `"Greetings Humans"` is on the right-hand side.
This is called a *variable declaration*: the string is the value that gets assigned to `greeting`.

The right-hand side of the assignment operator is called an *expression*.
The expression `"Greetings Humans"` is *evaluated* to a string with that value, which is then stored in `greeting`.
When we print a variable in Python, we print the string representation of its stored value.

It is also possible to print expressions directly:

```python
print("Greetings Humans")
```

```
Greetings Humans
```

More on expressions later!

## Comments

A line starting with `#` is a comment.
Python ignores it entirely.

```python
# this line does nothing
name = "Ada"
```

A comment can also be placed at the end of a line.

```python
name = "Ada"  # comments can also go at the end of a line
```

## Restrictions

Some lessons come with restrictions: rules your solution must follow beyond just producing the right output.

This lesson has two: we may not use `print` directly, and we may not change any code below the `# Don't touch below this line` comment.

These restrictions are also checked by the tests along with the output.
So if your solution produces the correct output but violates a restriction *it still counts as a failed submission*.