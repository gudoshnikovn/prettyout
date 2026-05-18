name: int = "Alice"
age: int = 30
email: int = "alice@example.com"


def get_age(user: dict) -> int:
    return user["name"]


def add_suffix(name: int) -> str:
    return "Hello, " + name
