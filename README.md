# FractalNet


## Location of Server:  http://35.230.55.6/



# Cellular Automata

* Each grid square is one of 4 different types:  Red, Green, Blue, or Empty.
* The "Neighborhood" of a cell is the 8 surrounding cells.


Color | Enemy | Fun Description
------|-------|------------------------
Red | Blue | Water puts out a fire.
Blue | Green | Earth grows out of water.
Green | Red | Fire burns the earth. 



## Order of Logic

**Note:**  This ordering is still under experimentation.

Currently, this is written as a if-then-else statement.  
Go through the list starting from #1.
If the conditions are satisfied, then apply that rule, 
and continue on to the next cell.


1. If there are 3 cells of the same color in the neighborhood of an empty cell, then it becomes that color.

2. If a non-empty cell (of any color) is surrounded by more than 5 non-empty cells, then it becomes an empty cell.

3. If a non-empty cell is surrounded by at least 1 "enemy", then it will be taken over by that enemy.

4. Otherwise, become an empty cell.







# Message Examples

Location of WebSocket:  ws://35.230.55.6/ws


## Chat

```JSON 
{
    "EventType": "Chat",
    "EventBody": "Hello World!",
}    
```





## Change a Square (Life Change)


Currently, the Grid is 48 x 48.  So the Location coordinates _x_ and _y_ are in the range of [0, 47].


Values match with the colors:

* 0 = empty
* 1 = Red
* 2 = Green
* 3 = Blue 



```JSON 
{
    "EventType": "LifeChange",
    "Value": 3,
    "Location": 
    {
        "X": 0,
        "Y": 47,
    }
}    
```





## Update to the Next Generation (Life Update)

```JSON 
{
    "EventType": "LifeUpdate",
}
```



