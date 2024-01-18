## Create a CLI tools for a game of number picking.

The initial idea reminded me of paper game that I was played on elementary school with my friend.
1 folded paper with pen is the only thing that required.
Each player draw ship in one side of the paper, then take turn to make guess by bold coloring the paper then flip it, so it leave remark.
If the remark (bold coloring) touch the other side of paper and correctly guess the enemy ship position, then it destroyed.

## Reason why the name ssg

Previously there was drt and inv project which is 3 letters, so I want to keep the culture.
After some minute can't think of any 3 initial for ship fire guess game, I remember one of the invoker spell 'sunstrike'.
That's why it called ssg 'sun strike game'.

## Requirements

Besides the initial requirements, I think it will be more fun if there is game rule of winning and losing.
One of the reason it reminded me about the paper guess game.

The initial requirement was to take 2 player input from number 0 - 123.
So I will map the input number as x, and y coordinates to put where the object that can be sunstriked.

Then after 5 input from 2 player to put the object, then it ask player to guess the input to activate the sunstrike.
