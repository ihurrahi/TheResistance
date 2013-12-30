TheResistance
=============
Author: Andy Lau

Personal implementation of The Resistance game to aid with game play.

While playing this game without the actual board game, I noticed it was hard to keep track of past missions, like who went on which mission and whether or not it failed - even the most recent one. I decided it would be cool to implement something that could help with that. I still wanted the game to be social, so I didnt want it to be a complete replacement. 

Dependencies
------------
* Go
* Go-MySQL (go get github.com/go-sql-driver/mysql)
 * MySQL (4.1 or higher, see github.com/go-sql-driver/mysql, tested with MySQL 5.1)
* Go-Socket.IO (go get github.com/justinfx/go-socket.io)
 * Go.net/websocket (go get code.google.com/p/go.net)
 * Socket.IO client javascript (https://github.com/LearnBoost/socket.io-client/blob/804c4e281e67b0a74a41a01f34103461c5788612/socket.io.js)
* GoZMQ (go get github.com/alecthomas/gozmq)
 * ZeroMQ (2.1.x, 2.2.x, or 3.x, see github.com/alecthomas/gozmq, tested with ZeroMQ 2.2.0)

Running
-----------
After downloading the source and the dependencies into the src/ directory, you can build the project using

    scripts/resistance build ALL

or to build an individual module:

    scripts/resistance build HTTP
    scripts/resistance build WSP
    scripts/resistance build GAME

To start the servers:

    scripts/resistance start ALL

or to start an individual module:

    scripts/resistance start HTTP
    scripts/resistance start WSP
    scripts/resistance start GAME

To stop the servers:

    scripts/resistance stop ALL

or to stop an individual module:

    scripts/resistance stop HTTP
    scripts/resistance stop WSP
    scripts/resistance stop GAME

