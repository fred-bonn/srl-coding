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

Many other format specifiers exist, which can be seen [here](https://docs.python.org/3/library/string.html#format-string-syntax).

## Width

An expression in front of the dot sets the minimum width of the rendered value.
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

## Newlines

The example above needed two `print` calls.
We can do this in a single `print` call using the newline character `\n`:

```python
print("149.95\n 49.90")
```

Output:

```
149.95
 49.90
```

## Helper Functions

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