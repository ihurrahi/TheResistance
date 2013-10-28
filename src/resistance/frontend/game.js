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
    case "missionPreparation":
      handleMissionPreparation(object);
      break;
    case "queryLeaderResult":
      handleQueryLeaderResult(object);
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
      if (button != null) {
        button.disabled = false;
      }
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
  
  // Show button to get role
  var button = document.getElementById("showRoleButton");
  if (button == null) {
    button = document.createElement("input");
    button.type = "button";
    button.value = "Show Role";
    button.id = "showRoleButton";
    button.onclick = function () {
      sendResistanceMessage("queryRole");
    }
    document.getElementById("roleInfo").appendChild(button);
  }
}

function handleQueryRoleResult(parsedMessage) {
  document.getElementById("showRoleButton").style.display = "none";
  var role = document.createTextNode(parsedMessage.role);
  document.getElementById("roleInfo").appendChild(role);
}

function handleMissionPreparation(parsedMessage) {
  // A mission needs to be sent, but which team?
  // Send a message to see if I'm the leader.
  sendResistanceMessage("queryLeader");
}

function handleQueryLeaderResult(parsedMessage) {
  var actionDiv = document.getElementById("action");
  var br = document.createElement("br");
  if (parsedMessage.isLeader) {
    message = "You are the leader of this mission";
    actionDiv.appendChild(document.createTextNode(message));
    actionDiv.appendChild(document.createElement("br"));
    message = "Please select a team of " + parsedMessage["teamSize"];
    actionDiv.appendChild(document.createTextNode(message));
    actionDiv.appendChild(document.createElement("br"));
    
    var form = document.createElement("form");
    for (var index in parsedMessage.players) {
      var player = parsedMessage.players[index];
      
      var option = document.createElement("input");
      option.type = "checkbox";
      option.id = "teamMember" + player["UserId"];
      option.value = player["UserId"];
      
      var label = document.createElement("label");
      label.innerHTML = player["Username"];
      label.htmlFor = "teamMember" + player["UserId"];
      
      form.appendChild(option);
      form.appendChild(label);
      form.appendChild(document.createElement("br"));
    }
    var submitButton = document.createElement("input");
    submitButton.type = "button";
    submitButton.value = "Send";
    submitButton.onclick = function() {
      var userIds = [];
      var inputs = form.getElementsByTagName("input");
      for (var index in inputs) {
        if (inputs[index].type == "checkbox" && inputs[index].checked) {
          userIds.push(inputs[index].value);
        }
      }
      if (userIds.length != parsedMessage["teamSize"]) {
        alert("Please select a team of " + parsedMessage["teamSize"]) + ".";
      } else {
        sendResistanceMessage("startMission", {"team": userIds});
      }
      return true;
    }
    form.appendChild(submitButton);
    
    actionDiv.appendChild(form);
  } else {
    actionDiv.appendChild(document.createTextNode("You are not the leader."));
  }
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
