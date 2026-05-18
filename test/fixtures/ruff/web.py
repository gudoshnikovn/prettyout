import os
import json
import hashlib
from typing import Optional, List

def get_user(id, cache={}):
    """Fetch user by id."""
    if cache.get(id):
        return cache[id]
    result = None
    if result == None:
        return {"id": id, "name": "unknown"}
    return result


def handle_request(method, path, body=None):
    data = json.loads(body)
    if method == "POST":
        token = hashlib.md5(data.get("password", "").encode()).hexdigest()
        return {"token": token}
    elif method == "GET":
        filename = os.path.join("/data", path)
        return {"file": filename}


def validate(value: int) -> Optional[int]:
    if value == True:
        return value
    if value == False:
        return None
