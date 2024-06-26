My first [bitboard](https://www.chessprogramming.org/Bitboards)-based engine.


### How to Use
Obtain a third-party UCI-compatible chess GUI. Build this module's binary file locally (`go build`). In the GUI settings, set the binary file as the "engine to use."


### Perft Milestones
[Perft](https://www.chessprogramming.org/Perft) is a debugging function that compares a move tree's leaf node count against an accepted value. The largest performance gains were from reducing memory allocations and the associated GC time.

| Depth | Time | Speed (million leaves/s) |
|---|---|---|
perft(4) | 0.15s | 1.35
perft(5) | 1.4s | 3.5
perft(6) | 16s | 7.1

### Want to build your own engine?
I recommend starting by making a command-line game that makes random, legal moves against the player. At that point, reading about move generator debugging, search algorithms, and evaluation algorithms will help your engine make stronger moves.
#### Learning Resources
- [Chess programming wiki](https://www.chessprogramming.org/Getting_Started)
- [the wiki's Recommended Reading](https://www.chessprogramming.org/Recommended_Reading)
- [TalkChess forum](https://talkchess.com/forum3/index.php)
