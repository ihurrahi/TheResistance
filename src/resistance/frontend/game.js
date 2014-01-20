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
    case "playerConnectSuccessful":
      handlePlayerConnectSuccessful(object);
      break;
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
    case "missionStarted":
      handleMissionStarted(object);
      break;
    case "queryIsOnMissionResult":
      handleQueryIsOnMissionResult(object);
      break;
    case "gameOver":
      handleGameOver(object);
      break;
    case "missions":
      handleMissions(object);
      break;
    default:
      // used for debugging
      // alert("Unknown message: " + object.message);
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

function handlePlayerConnectSuccessful(parsedMessage) {
  var actionDiv = document.getElementById("action");
  if (parsedMessage.isHost) {
    startButton = document.createElement("input");
    startButton.type = "button";
    startButton.value = "Start";
    startButton.id = "startButton";
    startButton.disabled = true;
    startButton.onclick = function () {
      startGame();
    }
    actionDiv.appendChild(startButton);
  } else {
    actionDiv.appendChild(document.createTextNode("Waiting for host to start game..."));
  }

  sendResistanceMessage("getPlayers");
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
  var role = document.createTextNode(parsedMessage.role);

  hideRole = function() {
    document.getElementById("showRoleButton").style.display = "inline";
	document.getElementById("roleInfo").removeChild(role);
  }

  window.setTimeout(hideRole, 3000);

  document.getElementById("showRoleButton").style.display = "none";
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
      for (var i = 0; i < inputs.length; i++) {
        if (inputs[i].type == "checkbox" && inputs[i].checked) {
          userIds.push(inputs[i].value);
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

function handleMissionStarted(parsedMessage) {
  sendResistanceMessage("queryIsOnMission");
}

function handleQueryIsOnMissionResult(parsedMessage) {
  clearActionDiv();
  var actionDiv = document.getElementById("action");
  if (parsedMessage.isOnMission) {
    actionDiv.appendChild(document.createTextNode("You have been selected to go on the mission."));
    addBreak(actionDiv);
    actionDiv.appendChild(document.createTextNode("Choose your mission outcome."));
    addBreak(actionDiv);

    var successButton = document.createElement("input");
    successButton.type = "button";
    successButton.value = "Success";
    successButton.onclick = function() {
      successButton.disabled = true;
      failButton.disabled = true;
      sendResistanceMessage("missionOutcome", {"outcome":true});
    }
    actionDiv.appendChild(successButton);

    var failButton = document.createElement("input");
    failButton.type = "button";
    failButton.value = "Fail";
    failButton.onclick = function() {
      successButton.disabled = true;
      failButton.disabled = true;
      sendResistanceMessage("missionOutcome", {"outcome":false});
    }
    actionDiv.appendChild(failButton);
  } else {
    actionDiv.appendChild(document.createTextNode("Waiting for mission to finish..."));
  }
}

function handleGameOver(parsedMessage) {
  alert("Game Over. The " + parsedMessage.winner + " team wins!");
  window.location.assign("/home.html");
}

function handleMissions(parsedMessage) {
  var missionInfoDiv = document.getElementById("missionInfo");
  var missionInfoTable = document.getElementById("missionTable");
  if (missionInfoTable != null) {
      missionInfoDiv.removeChild(missionInfoTable);
  }

  var table = document.createElement("table");
  table.id = "missionTable";

  // Create headers
  var header = table.createTHead();
  var row = header.insertRow(0);
  var cell1 = row.insertCell(0);
  var cell2 = row.insertCell(1);
  var cell3 = row.insertCell(2);
  var cell4 = row.insertCell(3);
  cell1.innerHTML = "Mission #";
  cell2.innerHTML = "Leader";
  cell3.innerHTML = "Result";
  cell4.innerHTML = "# Fails";

  for (var i in parsedMessage.missions) {
    var info = parsedMessage.missions[i];
    var row = table.insertRow(-1);
    var cell1 = row.insertCell(0);
    var cell2 = row.insertCell(1);
    var cell3 = row.insertCell(2);
    var cell4 = row.insertCell(3);

    cell1.innerHTML = info.missionNum;
    cell2.innerHTML = info.missionLeader.Username;
    cell3.innerHTML = info.missionResult;
	cell4.innerHTML = info.numFails;
  }

  missionInfoDiv.appendChild(table);
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
