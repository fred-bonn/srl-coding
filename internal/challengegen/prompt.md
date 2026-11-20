# Challenge Generation Prompt

Prompt for generating Python coding challenges with an LLM. Pass the example
outputs in the `samples/` directory alongside them as few-shot samples.

---

Generate a Python coding challenge about: **{TOPIC}**

The example outputs demonstrate format, structure, and style only. Produce a
challenge clearly distinct from every example: different function and class
names, a different scenario and domain, and different edge cases. Never reuse
an example's names, input data, or story.

Respond with a single JSON object and nothing else, following the structure
of the example outputs:

- `description` (string): instructions for the learner. State the required
  behaviour, including any special return values and edge cases, but do not
  reveal the solution.
- `starter_code` (string): valid, runnable Python. The functions and
  classes the learner must implement are stubs with final signatures,
  parameter names, and type hints, and `pass` as the body;
  already-implemented parts (e.g. `__init__`) stay working. No hints or
  comments.
- `test_code` (string): a pytest suite importing from `main`, split into
  `test_cases` (2-3 short, typical free checks) and `validation_cases`
  (exactly 2, at least one an edge case). A free run deselects the
  `validate`-marked cases; a validation run includes them and decides when
  the topic is reviewed. Cover expected exceptions in a separate free test
  with `pytest.raises`, and printing functions with `capsys` using soft,
  case-insensitive containment checks instead of exact matches. It may
  only rely on `pytest`, the Python standard library, and `main`.
- `solution_code` (string): a complete, correct implementation whose
  signatures match `starter_code` exactly. It must be a drop-in replacement
  for `main.py` that makes every test pass, and `test_code` must never
  import from or reference it.
- `difficulty` (integer): how hard the challenge is for a beginner-to-
  intermediate learner, from 1 (easiest) to 10 (hardest).
- `topics` (list of strings): short, lowercase topic names.

Output raw JSON only - no markdown fences, no commentary.
