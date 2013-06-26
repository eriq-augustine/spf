package server;

import (
   "encoding/json"
   "fmt"
   "code.google.com/p/go.net/websocket"
   "com/eriqaugustine/spf/game"
);

var connectionId int = 0;

// {connection/player id => websocket}
var connections = make(map[int]*websocket.Conn);

// {connection/player id => Game}
// Note: there are two entries per game (one per player).
var activeGames = make(map[int]*game.Game);

// TODO(eriq): queue
var waitingPlayer int = -1;

func GameServer(ws *websocket.Conn) {
   connectionId++;
   var id = connectionId;
   connections[id] = ws;

   var msg Message;

   println("On Connect");

   SocketLifeLoop:
   for {
      var err = websocket.JSON.Receive(ws, &msg);
      if err != nil {
         switch err.(type) {
            case *json.SyntaxError:
               println("unmarshal error!");
               websocket.Message.Send(ws, "unmarshal error");
               continue;
            default:
               fmt.Printf("ERROR: %s\n", err.Error());
               break SocketLifeLoop;
         }
      }

      var messagePart = msg.DecodeMessagePart();
      switch msgType := messagePart.(type) {
         case InitMessagePart:
            if (waitingPlayer == -1) {
               waitingPlayer = id;
            } else {
               var newGame *game.Game = game.NewGame(id, waitingPlayer);

               activeGames[id] = newGame;
               activeGames[waitingPlayer] = newGame;

               waitingPlayer = -1;
               BroadcastStart(newGame);
            }
         case MoveMessagePart:
            var movePart, _ = messagePart.(MoveMessagePart);

            // TEST
            println("Move");

            // TODO(eriq): will return nil on hash miss.
            var dropGroup, punishments =
               activeGames[id].MoveUpdate(id, movePart.Locations, movePart.BoardHash);
            signalNextTurn(id, dropGroup, punishments);
         default:
            fmt.Println("Unknow type: ", msgType);
            continue;
      }
   }

   // TODO(eriq): Verify that all the cleanup happens.
   delete(connections, id);
   delete(activeGames, id);

   println("On Close");
}

func BroadcastStart(currentGame *game.Game) {
   // Initial drop will be the same for both players.
   var message = NewMessage(MESSAGE_TYPE_START,
                            StartMessagePart{currentGame.InitialDrops()});

   for _, playerId := range currentGame.Players {
      websocket.JSON.Send(connections[playerId], message);
   }
}

func signalNextTurn(playerId int, dropGroup *[2]game.Gem, punishments int) {
   var currentGame *game.Game = activeGames[playerId];

   // Tell the player the next turn info
   var playerMessage =
      NewMessage(MESSAGE_TYPE_NEXT_TURN,
                 NextTurnMessagePart{*dropGroup, punishments,
                                     currentGame.Punishments[currentGame.GetOpponentOrdinal(playerId)]});

   websocket.JSON.Send(connections[playerId], playerMessage);

   // TODO(eriq): Update the opponent.
}
