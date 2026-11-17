# Challenge Generation Prompt

Prompt for generating Python coding challenges with an LLM. Pass the example
outputs in the `samples/` directory alongside them as few-shot samples.

---

Generate a Python coding challenge about: **{TOPIC}**

Respond with a single JSON object and nothing else. It must contain exactly
these fields:

- `description` (string): instructions for the learner describing what needs
  to be implemented. State the required behaviour clearly, including any
  special return values or edge cases, but do not reveal the solution.
- `starter_code` (string): valid, runnable Python. It defines the functions
  and classes the learner must implement as stubs: final signatures with
  parameter names and type hints, with `pass` as the body. For classes,
  provide the class skeleton with any already-implemented parts (e.g.
  `__init__`) left working and only the parts to implement stubbed. No hints
  or comments.
- `test_code` (string): a pytest test suite. It imports the function from
  `main` (the starter code will be saved as `main.py` next to the tests).
  Cases are split into two lists: `test_cases`, the free checks a learner
  can run at any time to see if their solution holds, and
  `validation_cases`, whose entries are
  `pytest.param(..., marks=pytest.mark.validate)`. Parametrized tests
  iterate over `test_cases + validation_cases`; when the learner validates
  their solution, all cases run and the outcome decides when the topic is
  reviewed again, while the free run deselects the `validate`-marked
  cases. Keep `test_cases` short and typical. `validation_cases` must
  contain exactly 2 cases, at least one of which is an edge case. Tests
  print the inputs, expected output, and actual
  result before each assert. If the function should raise an exception for
  some inputs (e.g. a `ValueError`), cover that in a separate free test
  using `pytest.raises`. If the function prints to standard output, capture
  it with the `capsys` fixture and keep the checks soft in both the free and
  the validation cases: assert that the expected values and keywords are
  contained in the output (case-insensitive), rather than matching the
  exact text, so slightly incorrect printings are not punished. It may
  only rely on `pytest`, the Python standard library, and `main`.
- `solution_code` (string): a complete, correct implementation with the same
  signature. It must be a drop-in replacement: copying its contents over
  `main.py` must make every test in `test_code` pass.
- `tags` (list of strings): short, lowercase topic tags, e.g.
  `["conditionals", "strings"]`.

Rules:

1. The function and class signatures in `starter_code` and `solution_code`
   must match exactly: same names, parameters, and type hints.
2. `test_code` must not import from or reference `solution_code`.
3. Aim for 2-3 `test_cases`; `validation_cases` is always exactly 2, with
   at least one edge case.
4. Output raw JSON only - no markdown fences, no commentary.
