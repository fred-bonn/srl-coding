def decimal_width(a: float, b: float) -> int:
    return max(len(f"{a:.2f}"), len(f"{b:.2f}"))


def align_decimals(a: float, b: float) -> str:
    width = decimal_width(a, b)
    return f"{a:{width}.2f}\n{b:{width}.2f}"
