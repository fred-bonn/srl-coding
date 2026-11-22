import subprocess
import sys


def test_prints_true():
    result = subprocess.run([sys.executable, "main.py"], capture_output=True, text=True)
    output = result.stdout.lower()
    print("\n---------------------------------")
    print('Expected: "true" in the output')
    print(f"Actual:   {output!r}")
    assert "true" in output, "your output must contain True"
