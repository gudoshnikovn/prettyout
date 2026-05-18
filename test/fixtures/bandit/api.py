import pickle
import subprocess
import requests

def run_command(cmd):
    subprocess.call(cmd, shell=True)

def load_data(data):
    return pickle.loads(data)

def fetch(url):
    return requests.get(url, verify=False)
