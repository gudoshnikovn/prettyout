from models import User


def greet(user: User) -> str:
    return "Hello, " + user.nickname  # attribute not found


def process(value: int) -> int:
    return value + "suffix"  # operator / argument type error


def show(items: list[int]) -> None:
    greet(items)  # wrong argument type
