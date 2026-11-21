import subprocess
import sys

import pytest

MARKER = "# Don't touch below this line"

# The exact code below the marker in the starter, which the user must leave untouched.
FIXED_CODE = "\n\nprint(hello)\n"


def test_prints_hello_world():
    result = subprocess.run([sys.executable, "main.py"], capture_output=True, text=True)
    output = result.stdout.lower()
    print("\n---------------------------------")
    print('Expected: "hello" and "world" in the output')
    print(f"Actual:   {output!r}")
    assert "hello" in output, "your output must contain the word 'hello'"
    assert "world" in output, "your output must contain the word 'world'"


@pytest.mark.validate
def test_restrictions():
    with open("main.py") as f:
        source = f.read()
    above, _, below = source.partition(MARKER)
    print("\n---------------------------------")
    print(f"Expected: no 'print' above the marker, code below the marker unchanged: {FIXED_CODE!r}")
    print(f"Actual:   code below the marker: {below!r}")
    assert "print" not in above, "you may not use print"
    assert below == FIXED_CODE, "do not change the code below the comment"
