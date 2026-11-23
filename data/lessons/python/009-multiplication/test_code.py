import ast
import subprocess
import sys

import pytest

MARKER = "# Don't touch above this line"

# The exact code above the marker in the starter, which the user must leave untouched.
FIXED_CODE = "a = 2\nb = 6\n\n"

ALLOWED_VARIABLES = {"a", "b"}


def test_prints_12():
    result = subprocess.run([sys.executable, "main.py"], capture_output=True, text=True)
    output = result.stdout.lower()
    print("\n---------------------------------")
    print('Expected: "12" in the output')
    print(f"Actual:   {output!r}")
    assert "12" in output, "your output must contain the number 12"


@pytest.mark.validate
def test_restrictions():
    with open("main.py") as f:
        source = f.read()
    above, _, _ = source.partition(MARKER)
    try:
        tree = ast.parse(source)
    except SyntaxError as e:
        assert False, f"your code has a syntax error: {e}"
    print("\n---------------------------------")
    print(f"Expected: code above the marker unchanged: {FIXED_CODE!r}")
    print(f"Actual:   code above the marker: {above!r}")
    assert above == FIXED_CODE, "do not change the code above the comment"
    for node in ast.walk(tree):
        if isinstance(node, (ast.Assign, ast.AugAssign, ast.AnnAssign)):
            targets = node.targets if isinstance(node, ast.Assign) else [node.target]
            for target in targets:
                if isinstance(target, ast.Name) and target.id not in ALLOWED_VARIABLES:
                    print("\n---------------------------------")
                    print(f"Expected: no new variables, only {sorted(ALLOWED_VARIABLES)}")
                    print(f"Actual:   assignment to '{target.id}' on line {target.lineno}")
                    assert False, f"you may not declare new variables (line {target.lineno})"
        if isinstance(node, ast.Constant) and isinstance(node.value, int) and node.value == 12:
            print("\n---------------------------------")
            print("Expected: no literal 12 anywhere in the code")
            print(f"Actual:   {source!r}")
            assert False, f"you may not use the value 12 directly (line {node.lineno})"
