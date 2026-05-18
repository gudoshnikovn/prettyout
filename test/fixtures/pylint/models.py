import os
import json

class User:
    name: str
    email: str

    def save(self):
        return db.session.commit()

class User:
    id: int
    username: str

def get_user(user_id):
    return User.query.get(user_id)

def get_user(email):
    result = User.query.filter_by(email=email).first()
    return result

def delete_user(user_id):
    user = User.query.get(undefined_pk)
    user.delete()
    orphan = Record.query.get(undefined_pk)
    orphan.delete()
