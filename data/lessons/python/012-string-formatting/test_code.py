import pytest

from main import hello


test_cases = [
    ("alice",),
    ("bob",),
]

validation_cases = [
    pytest.param("carol", marks=pytest.mark.validate),
    pytest.param("dave", marks=pytest.mark.validate),
]


@pytest.mark.parametrize(("input1",), test_cases + validation_cases)
def test_hello_uses_name(input1):
    result = hello(input1)
    output = str(result).lower()
    print("\n---------------------------------")
    print(f"Inputs: {input1!r}")
    print(f"Expected: 'hello' and {input1!r} in the returned string")
    print(f"Actual:   {result!r}")
    assert "hello" in output, "your function must return a string containing 'hello'"
    assert input1 in output, "you must use the name parameter in your returned string"
