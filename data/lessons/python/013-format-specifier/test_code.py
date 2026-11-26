import pytest

from main import align_decimals


test_cases = [
    (149.95, 49.901, "149.95\n 49.90"),
    (1234.5678, 9.5, "1234.57\n   9.50"),
]

validation_cases = [
    pytest.param(49.901, 149.95, " 49.90\n149.95", marks=pytest.mark.validate),
    pytest.param(0.123, 45.6, " 0.12\n45.60", marks=pytest.mark.validate),
    pytest.param(1.5, 2.5, "1.50\n2.50", marks=pytest.mark.validate),
]


@pytest.mark.parametrize(
    ("input1", "input2", "expected_output"), test_cases + validation_cases
)
def test_align_decimals(input1, input2, expected_output):
    result = align_decimals(input1, input2)
    print("\n---------------------------------")
    print(f"Inputs: {input1}, {input2}")
    print(f"Expected: {expected_output!r}")
    print(f"Actual:   {result!r}")
    assert result == expected_output, "your result must be both amounts rounded to two decimals, aligned on the decimal point"
