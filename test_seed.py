import requests

try:
    resp = requests.post("http://127.0.0.1:8080/api/dev/seed", json={"scenario": "launch-readiness"})
    print("Status:", resp.status_code)
    print("Response:", resp.text)
except Exception as e:
    print("Error:", e)
