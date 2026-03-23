import requests

resp = requests.post("http://127.0.0.1:8080/api/auth/login", json={"username": "admin", "password": "adminpass123"})
print(resp.status_code, resp.text)
if resp.status_code == 200:
    token = resp.json()["token"]
    resp2 = requests.post("http://127.0.0.1:8080/api/dev/seed", json={"scenario": "launch-readiness"}, headers={"Authorization": "Bearer " + token})
    print(resp2.status_code, resp2.text)
