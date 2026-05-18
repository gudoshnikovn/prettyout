import os

def render_user_profile(request, user_id, include_posts, include_comments, include_likes, include_followers, template="profile.html"):
    user = get_user(user_id)
    posts = user.posts if include_posts else []
    unused_data = {"include_comments": include_comments, "include_likes": include_likes, "include_followers": include_followers}
    description = "This is an extremely long description string that exceeds the line length limit of one hundred characters and should trigger C0301"
    return render(request, template, {"user": user, "posts": posts})

def create_user(name, email, roles=[]):
    user = {"name": name, "email": email, "roles": roles}
    return user
