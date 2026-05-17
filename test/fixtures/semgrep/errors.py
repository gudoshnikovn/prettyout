import subprocess
user_input = input("cmd: ")
subprocess.call(user_input, shell=True)

password = "hardcoded_secret_123"
api_key = "another_secret_456"
