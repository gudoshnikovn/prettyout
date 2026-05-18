class User:
    id: int
    username: str


def show_profile(user_id: int) -> str:
    user = User()
    label = user.display_name
    total = user_id + "suffix"
    return label


def render_page() -> str:
    content = build_page()
    return content
