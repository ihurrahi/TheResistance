TheResistance
=============
Author: Andy Lau

Personal implementation of The Resistance game to aid with game play.

While playing this game without the actual board game, I noticed it was hard to keep track of past missions, like who went on which mission and whether or not it failed - even the most recent one. I decided it would be cool to implement something that could help with that. I still wanted the game to be social, so I didnt want it to be a complete replacement. 

Dependencies
------------
* Go
* Go-MySQL (github.com/go-sql-driver/mysql)
* Go-Socket.IO (github.com/justinfx/go-socket.io)
** Go.net/websocket (code.google.com/p/go.net)
* ZeroMQ (github.com/alecthomas/gozmq)

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

