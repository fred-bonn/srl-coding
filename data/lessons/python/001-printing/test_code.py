import subprocess
import sys


def test_print_hello_world():
    result = subprocess.run([sys.executable, "main.py"], capture_output=True, text=True)
    output = result.stdout.lower()
    print("\n---------------------------------")
    print('Expected: "hello" and "world" in the output, in that order')
    print(f"Actual:   {output!r}")
    assert "hello" in output, "your output must contain the word 'hello'"
    assert "world" in output, "your output must contain the word 'world'"
    assert output.index("hello") < output.index("world"), "'hello' must appear before 'world' in your output"
