# Variables

A variable is a named place to store a value, so you can refer to it later by name. You create one with an assignment: the variable name on the left, `=`, and the value on the right.

## Example

```python
greeting = "Greetings Humans"
print(greeting)
```

Output:

```
Greetings Humans
```

This is called a *variable declaration* of the variable `greeting` and we say that the variable `greeting` is *initialised* to `"Greetings Humans`.

## Comments

A line starting with `#` is a comment. Python ignores it entirely.

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

This lesson has two: you may not use `print` directly, and you may not change any code below the `# Don't touch below this line` comment.

These restrictions are also checked by the tests along with the output. So if your solution produces the correct output but violates a restriction *it still counts as a failed submission*.