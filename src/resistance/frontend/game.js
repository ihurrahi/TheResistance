var socket = new io.Socket(null, {port: 8081, rememberTransport: false});

socket.on('message', function(obj) {
  handleMessage(obj);
});

function handleMessage(message) {
  object = JSON.parse(message);
  
  handleAnyErrors(object);
  
  switch(object.message) {
    case "players":
      handlePlayers(object);
      break;
    default:
      alert("Unknown message: " + object.message);
  }
}

function handleAnyErrors(parsedMessage) {
  var div = document.getElementById("alerts");
  if ("errorMessage" in object) {
    div.innerHTML = "<b>Error: " + object.errorMessage + "</b>";
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
  }
}

function sendResistanceMessage(message, arguments) {
  var packet = {};
  packet["message"] = message;
  packet["userCookie"] = document.cookie;
  for (var property in arguments) {
    packet[property] = arguments[property];
  }
  socket.send(JSON.stringify(packet));
}

function doAction(gameId) {
  sendResistanceMessage("start",{"gameId":gameId});
}

function playerConnect(gameId) {
  sendResistanceMessage("playerConnect",{"gameId":gameId});
}