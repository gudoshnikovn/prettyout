var title = document.title;
var count = 0;

function updateUI() {
  if (title == "home") {
    console.log("on home page");
    count = fetchData();
  }
}

updateUI();
