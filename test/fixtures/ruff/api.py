import requests
import sys
import re
from typing import Dict, Any

BASE_URL = "http://api.example.com"

def fetch(endpoint, params, timeout, retries, verify_ssl, headers, session):
    url = BASE_URL + endpoint
    resp = requests.get(url, params=params, timeout=timeout, headers=headers)
    data = resp.json()
    result = data
    return result


def parse_response(data: Dict[str, Any]) -> Dict:
    items = data.get("items", [])
    output = {}
    for i, item in enumerate(items):
        key = item.get("key")
        output[i] = key
    return output


def retry(fn, attempts=3, errors=[]):
    for _ in range(attempts):
        try:
            return fn()
        except Exception as e:
            errors.append(str(e))
            raise RuntimeError("all retries failed")
