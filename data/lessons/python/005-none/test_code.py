import ast
import subprocess
import sys

import pytest


def test_prints_none():
    result = subprocess.run([sys.executable, "main.py"], capture_output=True, text=True)
    output = result.stdout
    print("\n---------------------------------")
    print('Expected: "None" in the output')
    print(f"Actual:   {output!r}")
    assert "None" in output, "your output must contain None"


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


@pytest.mark.validate
def test_uses_none():
    with open("main.py") as f:
        source = f.read()
    try:
        tree = ast.parse(source)
    except SyntaxError as e:
        assert False, f"your code has a syntax error: {e}"
    print("\n---------------------------------")
    print("Expected: the None value used somewhere in the code")
    print(f"Actual:   {source!r}")
    for node in ast.walk(tree):
        if isinstance(node, ast.Constant) and node.value is None:
            return
    assert False, "you must use the None value in your code"
