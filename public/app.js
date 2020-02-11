(function() {
  // var ws = new WebSocket("ws://localhost:1338");

  // var chatform = document.querySelector(".chatform");
  var loginform = document.querySelector(".loginform");
  var registerform = document.querySelector(".registerform");
  // document.querySelector("#chat").style.display = "none";
  async function sendLogging(login, password) {
    const loginResponse = await fetch("http://localhost:1337/login", {
      method: "POST",
      body: JSON.stringify({
        username: login,
        password: password
      }),
      headers: {
        "Content-Type": "application/json"
      }
    });
    const loginResponseValue = await loginResponse.text();

    if (loginResponse.status !== 200) {
      document.querySelector(".error-login").textContent = loginResponseValue;
      return;
    } else {
      document.querySelector(".error-login").textContent = "Bienvenue !";
      // Simulate a mouse click:
      window.location.href = "/pages/chat/";
    }

    // const ticketResponse = await fetch("http://localhost:1337/wsTicket", {
    //   method: "GET"
    // });

    // const ticketResponseValue = await ticketResponse.text();
    // if (ticketResponse.ok) {
    //   ws.send(ticketResponseValue);
    //   document.querySelector("#login").style.display = "none";
    //   document.querySelector("#chat").style.display = "block";
    //   document.querySelector("input[name=message]").focus();
    // } else {
    //   document.querySelector(".error").textContent = ticketResponseValue;
    // }
  }

  async function sendRegister(login, password, email) {
    const response = await fetch("http://localhost:1337/register", {
      method: "POST",
      body: JSON.stringify({
        username: login,
        password: password,
        email: email
      }),
      headers: {
        "Content-Type": "application/json"
      }
    });

    const registerResponseValue = await response.text();
    if (response.status !== 200) {
      document.querySelector(
        ".error-register"
      ).textContent = registerResponseValue;
    } else {
      document.querySelector(".error-register").textContent =
        "Le compte a été créé !";
      document.querySelector(".registerform").style.display = "none";
      document.querySelector(".container-login").style.display = "block";
    }
  }

  loginform.onsubmit = function(e) {
    e.preventDefault();
    var loginInput = loginform.querySelector("input[name=login]");
    var passwordInput = loginform.querySelector("input[name=password]");
    var login = loginInput.value;
    var password = passwordInput.value;
    sendLogging(login, password);
    loginInput.value = "";
    passwordInput.value = "";
  };

  registerform.onsubmit = function(e) {
    e.preventDefault();
    var loginInput = registerform.querySelector("input[name=login]");
    var passwordInput = registerform.querySelector("input[name=password]");
    var emailInput = registerform.querySelector("input[name=email]");
    sendRegister(loginInput.value, passwordInput.value, emailInput.value);
    loginInput.value = "";
    passwordInput.value = "";
    emailInput.value = "";
  };

  // chatform.onsubmit = function(e) {
  //   e.preventDefault();
  //   var input = document.querySelector("input[name=message]");
  //   var text = input.value;
  //   ws.send(text);
  //   input.value = "";
  //   input.focus();
  //   return false;
  // };

  // ws.onmessage = function(msg) {
  //   var response = msg.data;
  //   var messageList = document.querySelector(".messages");
  //   var li = document.createElement("li");
  //   li.textContent = response;
  //   messageList.appendChild(li);
  // };
})();
