# Division

Python has three division operators.
Depending on which operator we use, the evaluated result and its type change.

- `/`: Regular division. Evaluates to a `float`.
- `//`: Integer division. Evaluates to an integer when both operands are integers, and to a `float` otherwise.
- `%`: Remainder operator. Evaluates to an integer when both operands are integers, and to a `float` otherwise.

## Example

```python
print(19 / 4)
```

Output:

```
4.75
```

Regular division works as expected arithmetically.
The integer division `//` changes this by only keeping the whole number part of the result:

```python
print(19 // 4)
```

Output:

```
4
```

What `//` throws away is available through the remainder operator `%`, which evaluates to what is left over after the division:

```python
print(19 % 4)
```

Output:

```
3
```

## Type Casting

All three operators accept integers and floats in any combination.
When the operands have different types, Python *implicitly type casts* the integer operand to a float for us before evaluating the expression.

```python
print(11 // 4.0)
print(11.0 % 4)
```

Output:

```
2.0
3.0
```

In `11 // 4.0`, the integer `11` is cast to the float `11.0`, and the expression is evaluated as `11.0 // 4.0`.
Since both operands are now floats, the result is a float too.
That is why `//` and `%` evaluate to an integer only when both operands are integers.
With a float anywhere in the expression, everything becomes a float.

We can also cast explicitly, using the `int` and `float` functions.
The `int` function evaluates to its argument cast to an integer, and the `float` function to its argument cast to a float:

```python
print(int(11.0))
print(float(11))
```

Output:

```
11
11.0
```

Explicit casting lets us control the types ourselves.
Compare the two divisions below: the first has a float operand, so the result is a float.
In the second, `int(11.0)` evaluates to the integer `11` first, so `//` divides two integers and the result is an integer:

```python
print(11 // 4.0)
print(int(11.0) // 4)
```

Output:

```
2.0
2
```

However, this does not work on all types.
Trying to divide a number with a string will cause an *error*.
More on errors later!