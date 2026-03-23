const res = await fetch("http://127.0.0.1:8080/api/auth/login", {
  method: "POST",
  headers: {"Content-Type": "application/json"},
  body: JSON.stringify({username: "admin", password: "adminpass123"})
});
const data = await res.json();
console.log(data);

const res2 = await fetch("http://127.0.0.1:8080/api/dev/seed", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "Authorization": "Bearer " + data.token
  },
  body: JSON.stringify({scenario: "launch-readiness"})
});
console.log(await res2.text());
