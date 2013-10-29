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
    case "teamApproval":
      handleTeamApproval(object);
      break;
    case "approveTeamUpdate":
      handleApproveTeamUpdate(object);
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
  clearActionDiv();
  var actionDiv = document.getElementById("action");
  if (parsedMessage.isLeader) {
    message = "You are the leader of this mission";
    actionDiv.appendChild(document.createTextNode(message));
    addBreak(actionDiv);
    message = "Please select a team of " + parsedMessage["teamSize"] + ".";
    actionDiv.appendChild(document.createTextNode(message));
    addBreak(actionDiv);
    
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
      addBreak(form);
    }
    addBreak(form);
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
        alert("Please select a team of " + parsedMessage["teamSize"] + ".");
      } else {
        submitButton.disabled = true;
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

function handleTeamApproval(parsedMessage) {
  clearActionDiv();
  var actionDiv = document.getElementById("action");
  var message = "Do you approve of this team?"
  actionDiv.appendChild(document.createTextNode(message));
  addBreak(actionDiv);
  for (var index in parsedMessage.team) {
    actionDiv.appendChild(document.createTextNode(parsedMessage.team[index]));
    addBreak(actionDiv);
  }
  var yesButton = document.createElement("input");
  yesButton.type = "button";
  yesButton.value = "Yes";
  yesButton.onclick = function() {
    yesButton.disabled = true;
    noButton.disabled = true;
    sendResistanceMessage("approveTeam", {"vote":true});
  }
  actionDiv.appendChild(yesButton);
  
  var noButton = document.createElement("input");
  noButton.type = "button";
  noButton.value = "No";
  noButton.onclick = function() {
    yesButton.disabled = true;
    noButton.disabled = true;
    sendResistanceMessage("approveTeam", {"vote":false});
  }
  actionDiv.appendChild(noButton);
}

function handleApproveTeamUpdate(parsedMessage) {
  var actionDiv = document.getElementById("action");
  var message = parsedMessage.username + " voted ";
  if (parsedMessage.vote == true) {
    message += "yes";
  } else if (parsedMessage.vote == false) {
    message += "no";
  }
  addBreak(actionDiv);
  actionDiv.appendChild(document.createTextNode(message));
}

function addBreak(divElement) {
  var br = document.createElement("br");
  divElement.appendChild(br);
}

function clearActionDiv() {
  var actionDiv = document.getElementById("action");
  actionDiv.innerHTML = ""
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
