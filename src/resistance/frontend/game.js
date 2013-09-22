var socket = new io.Socket(null, {port: 8081, rememberTransport: false});

socket.on('message', function(obj) {
    var row = document.getElementById("players").insertRow(-1);
    var cell = row.insertCell(0);
    cell.innerHTML = obj;
});

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