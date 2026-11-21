import ast
import subprocess
import sys

import pytest


def test_prints_42():
    result = subprocess.run([sys.executable, "main.py"], capture_output=True, text=True)
    output = result.stdout.lower()
    print("\n---------------------------------")
    print('Expected: "42" in the output')
    print(f"Actual:   {output!r}")
    assert "42" in output, "your output must contain the number 42"


@pytest.mark.validate
def test_no_strings():
    with open("main.py") as f:
        source = f.read()
    try:
        tree = ast.parse(source)
    except SyntaxError as e:
        assert False, f"your code has a syntax error: {e}"
    print("\n---------------------------------")
    print("Expected: no string literal anywhere in the code")
    print(f"Actual:   {source!r}")
    for node in ast.walk(tree):
        if isinstance(node, ast.Constant) and isinstance(node.value, str):
            assert False, f"you may not use strings (line {node.lineno})"
