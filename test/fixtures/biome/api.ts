var baseUrl = "https://api.example.com";
const unusedToken = "secret";

function fetchUser(id: number) {
  return fetch(baseUrl + "/users/" + id);
}

fetchUser(1);
