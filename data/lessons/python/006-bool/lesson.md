# Booleans

A boolean (`bool`) is a value that is either `True` or `False`.
It is not a string, and capitalization matters: `True` and `False` always start with a capital letter.

As with integers and floats, enclosing a boolean in quotes would turn it into a string: `True` is a boolean, `"True"` is a string.

## Example

```python
is_open = True
print(is_open)
```

Output:

```
True
```

## Truthiness

Every value in Python can be treated as a boolean, no matter its type.
We can convert any value to a boolean with the `bool` function:

```python
print(bool(42))
print(bool(0))
print(bool(3.14))
print(bool(0.0))
print(bool("Greetings Humans"))
print(bool(""))
print(bool(None))
```

Output:

```
True
False
True
False
True
False
False
```

Values that represent "nothing" are converted to `False`, and every other value is converted to `True`.
For example, the empty string `""` and `None` are converted to `False`, while every other possible string value is converted to `True`.

Values that convert to `True` are called *truthy*, and values that convert to `False` are called *falsy*.
