import pytest

from main import hello_world


def test_hello_world_greets():
    output = str(hello_world()).lower()
    print("\n---------------------------------")
    print('Expected: "hello" and "world" in the returned string, in that order')
    print(f"Actual:   {output!r}")
    assert "hello" in output, "your function must return a string containing 'hello'"
    assert "world" in output, "your function must return a string containing 'world'"
    assert output.index("hello") < output.index("world"), "'hello' must come before 'world' in your returned string"


@pytest.mark.validate
def test_hello_world_returns_exact_string():
    result = hello_world()
    print("\n---------------------------------")
    print("Expected: 'Hello, World!'")
    print(f"Actual:   {result!r}")
    assert result == "Hello, World!", "your function must return 'Hello, World!'"
