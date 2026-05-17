import os
import sys

def foo():
    try:
        return 1
    except Exception:
        raise ValueError("fail")
