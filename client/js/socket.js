"use strict";

// TODO(eriq): Real name
// Socket.SERVER = 'ws://www.eriqaugustine.com/gamesocket';
Socket.SERVER = 'ws://66.169.236.4:3030/gamesocket';
// Socket.SERVER = 'ws://localhost:12345/testsocket';

function Socket() {
   this.ws = new WebSocket(Socket.SERVER);

   this.ws.onmessage = this.onMessage.bind(this);
   this.ws.onclose = this.onClose.bind(this);
   this.ws.onopen = this.onOpen.bind(this);
   this.ws.onerror = this.onError.bind(this);
}

Socket.prototype.onMessage = function(messageEvent) {
   var message = null;
   try {
      message = JSON.parse(messageEvent.data);
   } catch (ex) {
      error('Server message does not parse.');
      return false;
   }

   switch (message.Type) {
      case Message.TYPE_START:
         startGame([new DropGroup(message.Payload.Drops[0]),
                    new DropGroup(message.Payload.Drops[1])]);
         break;
      case Message.TYPE_NEXT_TURN:
         nextTurnInfo(new DropGroup(message.Payload.Drop),
                      message.Payload.PlayerPunishment,
                      message.Payload.OpponentPunishment);

         if (message.Payload.Lose) {
            loseGame();
         }
         break;
      case Message.TYPE_UPDATE:
         var nextDrop = message.Payload.Win ? null : new DropGroup(message.Payload.OpponentNextDropGroup);

         updateOpponent(message.Payload.OpponentPunishment,
                        message.Payload.OpponentBoard,
                        nextDrop);

         if (message.Payload.Win) {
            winGame();
         }
         break;
      case Message.TYPE_NO_CONTEST:
         noContest();
         break;
      case Message.TYPE_DROP_GROUP_UPDATE:
         opponentDropGroupUpdate(message.Payload.Locations);
         break;
      default:
         // Note: There are messages that are known, but just not expected from the server.
         error('Unknown Message Type: ' + message.Type);
         break;
   }
};

Socket.prototype.onClose = function(messageEvent) {
   debug("Connection to server closed.");
   noContest();
};

Socket.prototype.onOpen = function(messageEvent) {
   debug("Connection to server opened.");
   this.ws.send(createInitMessage());
};

Socket.prototype.onError = function(messageEvent) {
   debug(JSON.stringify(messageEvent));
};

Socket.prototype.close = function() {
   this.ws.close();
};

Socket.prototype.sendMove = function(dropGemLocations, boardHash) {
   this.ws.send(createMoveMessage(dropGemLocations, boardHash));
};

Socket.prototype.sendDropGroupUpdate = function(dropGemLocations) {
   this.ws.send(createDropGroupUpdateMessage(dropGemLocations));
};
