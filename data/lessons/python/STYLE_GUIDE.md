# Lesson Style Guide

How to author and maintain the lessons in this directory.
Everything here is generalized from the implemented lessons; when adding a lesson, make it match.

## Lesson structure

Every lesson is a directory with exactly six files:

| File               | Purpose                                              |
|--------------------|------------------------------------------------------|
| `lesson.md`        | The lesson text: explanation and examples.            |
| `assignment.md`    | The task the learner must solve, plus restrictions.   |
| `metadata.json`    | Lesson metadata (difficulty, topics).                 |
| `starter_code.py`  | Starting point copied into the learner's `main.py`. For unit lessons, a stub with the signature and `pass` bodies. |
| `solution_code.py` | A reference solution that passes every test.                       |
| `test_code.py`     | The tests the solution is checked against.            |

Directory names are `NNN-topic-name`: zero-padded number, lowercase, hyphenated.
Numbers define the order lessons are taken in. When inserting a lesson, renumber the later ones with `git mv`.

## lesson.md

- Title is the bare topic, capitalized: `# Booleans`, `# None`. No "in Python" suffix.
- One concept per lesson. A lesson introduces exactly one type or idea, and builds on the previous ones.
- Voice: personal ("we")
- Spelling: American.
- Paragraphs are written one sentence per line.
- Every lesson has an `## Example` section: a short, runnable snippet followed by an `Output:` block showing its exact output. Run every example and verify the output block before submitting.
- Extra sections (`## Truthiness`, `## Comments`, ...) are allowed when the concept needs them, in the same style.
- Inline code comments use two spaces before `#` (PEP 8).

## assignment.md

Template:

```markdown
Write a Python program that prints `X`.

Restrictions:

- You may not use strings.
```

- One task sentence, then an optional `Restrictions:` list.
- **Every restriction must map to atleast one test.** A restriction the tests do not enforce is not a restriction; either enforce it or drop it.
- Files end with a trailing newline.

## test_code.py

Two test styles exist, matching the two kinds of lessons:

**Printing lessons** are lessons that run the learner's `main.py` as a subprocess:

```python
result = subprocess.run([sys.executable, "main.py"], capture_output=True, text=True)
```

- Assertions go against the captured stdout.
- Output tests are lenient for `print`-based lessons: case and punctuation tolerance as described in lesson 001's `## Submission` section.
- Restrictions on source code are checked by parsing `main.py` with `ast` and walking the tree, failing with the offending line number.

**Unit lessons** (functions, classes) import the learner's code directly and test return values or instance state:

```python
from main import Order
```

- Test inputs live in a module-level `test_cases` list of tuples; use `@pytest.mark.parametrize` over them.
- Additional edge cases go in a `validation_cases` list built with `pytest.param(..., marks=pytest.mark.validate)` and are concatenated onto `test_cases` in the parametrize call.
- Tests may capture stdout with pytest's `capsys` fixture when the assignment is about printing.

Rules common to both styles:

- Restriction tests are marked `@pytest.mark.validate` (whole-program) or use validation cases with `marks=pytest.mark.validate` (unit).
- Before each assertion, print a diagnostic block: a `----` separator, the inputs, the expected result, and the actual result.
- Failure messages are lowercase and addressed to the learner: `"you may not use strings (line 3)"`.

## metadata.json

```json
{
    "difficulty": 1,
    "topics": [
        "none"
    ]
}
```

- `difficulty`: integer, relative to the other lessons.
- `topics`: lowercase, one entry per introduced type or concept.
- 4-space indent, matching the existing files.

## Python code style

TODO: to be filled in.
