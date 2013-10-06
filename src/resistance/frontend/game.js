function parseUrlParams() {
    var urlParams = {};
    var regex = /([^&=]+)=?([^&]*)/g;
    var parameters = window.location.search.substring(1);
    while (match = regex.exec(parameters)) {
      urlParams[match[1]] = match[2];
    }
    return urlParams;
}

function handleMessage(message) {
  object = JSON.parse(message);
  
  handleAnyErrors(object);
  
  switch(object.message) {
    case "players":
      handlePlayers(object);
      break;
    case "gameStarted":
      handleGameStart(object);
      break;
    case "queryRoleResult":
      handleQueryRoleResult(object);
      break;
    default:
      // used for debugging
      //alert("Unknown message: " + object.message);
  }
}

function handleAnyErrors(parsedMessage) {
  var div = document.getElementById("alerts");
  if ("errorMessage" in parsedMessage) {
    div.innerHTML = "<b>Error: " + parsedMessage.errorMessage + "</b>";
  } else {
    div.innerHTML = "";
  }
}

function handlePlayers(parsedMessage) {
  if ("players" in parsedMessage) {
    var playersTable = document.getElementById("players");
    playersTable.innerHTML = ""
    for (var i = 0; i < parsedMessage.players.length; i++) {
      var row = playersTable.insertRow(-1);
      var cell = row.insertCell(0);
      cell.innerHTML = parsedMessage.players[i];
    }
    if (parsedMessage.players.length >= 5) {
      var button = document.getElementById("startButton");
      button.disabled = false;
    }
  }
}

function handleGameStart(parsedMessage) {
  // remove the start button if it exists
  var startButton = document.getElementById("startButton");
  var actionDiv = document.getElementById("action");
  if (startButton != null && actionDiv != null) {
      actionDiv.removeChild(startButton);
  }
  
  // hack. remove
  document.getElementById("information").innerHTML += "The game has started!"
  
  // figure out what my role is
  sendResistanceMessage("queryRole");
}

function handleQueryRoleResult(parsedMessage) {
  document.getElementById("information").innerHTML += "<br>Your role:<br>"
  document.getElementById("information").innerHTML += parsedMessage.role
}

function sendResistanceMessage(message, arguments) {
  var packet = {};
  packet["message"] = message;
  packet["userCookie"] = document.cookie;
  packet["gameId"] = gameId;
  for (var property in arguments) {
    packet[property] = arguments[property];
  }
  socket.send(JSON.stringify(packet));
}

function startGame() {
  sendResistanceMessage("startGame");
}

function playerConnect() {
  sendResistanceMessage("playerConnect");
}

var socket = new io.Socket(null, {port: 8081, rememberTransport: false});

socket.on('message', function(obj) {
  handleMessage(obj);
});

gameId = parseUrlParams()["gameId"];
socket.connect();
playerConnect();
