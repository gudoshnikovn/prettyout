let baseUrl = "https://api.example.com";
const unusedToken = "secret";
const unusedTimeout = 5000;

function fetchUser(id) {
  return fetch(baseUrl + "/users/" + id + "?token=" + apiKey);
}

fetchUser(1);
