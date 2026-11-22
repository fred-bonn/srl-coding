# Printing

Printing is how a program shows output to the user.
In Python, we use the built-in `print` function.
It takes one or more values and writes them to standard output, followed by a newline.

## Example

```python
print("Greetings")
print("Humans")
```

Output:

```
Greetings
Humans
```

## Submission

While working on a solution we can run our code as often as we like.
Running executes our program and checks it against some of the lesson's tests, so we get feedback while we work. Runs cost nothing.

When we submit a solution to an assignment our code is checked against all of the lesson's tests, including tests that were withheld from our runs.
Each test compares what our program actually did with what was expected.
Afterward we get a report back, listing each test as passed or failed.
For every failed test it also reports what the expected result was and what our code produced instead.

Running and submitting are not the same thing: running is practice, submitting is turning our homework in for grading.
We can run as much as we need, and submit when we are confident in our solution.

For tests involving `print` the tests strive to be lenient so we do not have to produce the exact same output.
For example, for this lesson, the test checks whether the uncapitalized words `hello` and `world` appear in the output in that order and ignores all punctuation.