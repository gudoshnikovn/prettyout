import flask
import hashlib
import random

SECRET_KEY = "hardcoded_password_123"

app = flask.Flask(__name__)

def get_token():
    return random.random()

def hash_password(pwd):
    return hashlib.md5(pwd.encode()).hexdigest()

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0")
