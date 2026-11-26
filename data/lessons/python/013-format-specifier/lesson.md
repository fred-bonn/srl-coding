# Format Specifiers

In the last lesson, an f-string inserted a value into a string, and that was it.
A *format specifier* tells the f-string how to render the value: how many decimals to show, how wide the result should be.

## Example

The format specifier `.2f` renders a number in fixed-point notation with exactly two digits after the decimal point:

```python
price = 49.901
print(f"{price:.2f}")
```

Output:

```
49.90
```

The value is rounded to the two digits it shows.

Other format specifiers exist, far more than we need here.
Whenever we want more control, the Python documentation lists them all.

## Width

A number in front of the dot sets the minimum width of the rendered value.
If the value is shorter, it is padded on the left with spaces:

```python
print(f"{price:6.2f}")
print(f"{price:9.2f}")
```

Output:

```
 49.90
    49.90
```

Numbers are right-aligned by default: the padding goes on the left.
That is exactly what we want when stacking numbers in a column, because their right edges, and their decimal points, line up:

```python
large = 149.95
small = 49.901
print(f"{large:6.2f}")
print(f"{small:6.2f}")
```

Output:

```
149.95
 49.90
```

Both lines are six characters wide, so the decimal points sit in the same column.

## Newlines

The column above needed two `print` calls, one line per call.
A single string can hold several lines: `\n` is a newline character, and it ends the current line and starts a new one.

```python
print("149.95\n 49.90")
```

Output:

```
149.95
 49.90
```

`\n` counts as one character, written as a backslash and an `n`.
It can appear any number of times inside a string, f-strings included.

## A Width From a Variable

Any part of a format specifier can be a replacement field of its own, so the width can come from a variable:

```python
width = 9
print(f"{price:{width}.2f}")
```

Output:

```
    49.90
```

In the assignment the two amounts must share one width.
The starter code provides a helper, `decimal_width`, that evaluates to it:

```python
print(decimal_width(149.95, 49.901))
```

Output:

```
6
```

The tests never check the helpers, only the function the assignment asks us to implement.
We may write as many helpers of our own as we want, in this lesson and every one after it.