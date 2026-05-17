import subprocess
user_input = input("cmd: ")
subprocess.call(user_input, shell=True)

password = "hardcoded_secret_123"
