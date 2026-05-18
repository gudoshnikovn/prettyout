from typing import Optional


class User:
    name: str
    age: int

    def get_name(self) -> str:
        return 42  # wrong return type

    def get_age(self) -> int:
        pass  # implicit None return


x: int = "hello"  # assignment type error
y: str = 123      # another assignment error

result = "text" + 5  # operator error
